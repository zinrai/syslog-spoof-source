package main

import (
	"fmt"
	"net"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

// buildPacket assembles an IPv4 and UDP packet carrying payload.
//
// The kernel would fill in the IP total length and checksum on send, but they
// are computed here so the packet is well-formed before it reaches the socket.
func buildPacket(src, dst net.IP, sport, dport uint16, ipID uint16, payload []byte) ([]byte, error) {
	ip := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Id:       ipID,
		Flags:    layers.IPv4DontFragment,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    src.To4(),
		DstIP:    dst.To4(),
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(sport),
		DstPort: layers.UDPPort(dport),
	}
	// Required before serialization, otherwise the UDP checksum cannot be
	// computed: it covers a pseudo header taken from the IP layer.
	if err := udp.SetNetworkLayerForChecksum(ip); err != nil {
		return nil, fmt.Errorf("cannot bind the UDP checksum to the IP layer: %w", err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	if err := gopacket.SerializeLayers(buf, opts, ip, udp, gopacket.Payload(payload)); err != nil {
		return nil, fmt.Errorf("cannot serialize the packet: %w", err)
	}
	return buf.Bytes(), nil
}
