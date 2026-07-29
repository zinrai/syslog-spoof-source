package main

import (
	"bytes"
	"net"
	"testing"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

// The source address is the whole point of the tool, so it has to survive
// serialization unchanged rather than being replaced by the sending host's.
func TestPacketCarriesTheGivenSource(t *testing.T) {
	pkt, err := buildPacket(net.ParseIP("192.0.2.1"), net.ParseIP("10.0.0.5"), 40000, 514, 1234, []byte("<134>x"))
	if err != nil {
		t.Fatal(err)
	}

	p := gopacket.NewPacket(pkt, layers.LayerTypeIPv4, gopacket.Default)
	ip, _ := p.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if ip == nil {
		t.Fatal("no IPv4 layer")
	}
	if got := ip.SrcIP.String(); got != "192.0.2.1" {
		t.Errorf("source: got %s", got)
	}
	if got := ip.DstIP.String(); got != "10.0.0.5" {
		t.Errorf("destination: got %s", got)
	}
}

// A wrong checksum is dropped silently by the receiver. gopacket computes it,
// but only when SetNetworkLayerForChecksum has been called; this guards
// against that call being lost. Verified against a decode across payload
// lengths, since odd lengths exercise the padding path.
func TestPacketChecksumsVerify(t *testing.T) {
	for _, payload := range [][]byte{
		[]byte(""),
		[]byte("<134>Jul 29 06:09:41 dev-1 app[0]: seq=0"),  // even length
		[]byte("<134>Jul 29 06:09:41 dev-1 app[0]: seq=01"), // odd length
		bytes.Repeat([]byte("x"), 1400),
	} {
		pkt, err := buildPacket(net.ParseIP("192.0.2.1"), net.ParseIP("10.0.0.5"), 40000, 514, 1, payload)
		if err != nil {
			t.Fatal(err)
		}
		p := gopacket.NewPacket(pkt, layers.LayerTypeIPv4, gopacket.Default)
		if err := p.ErrorLayer(); err != nil {
			t.Errorf("payload %d bytes: decode failed: %v", len(payload), err.Error())
		}
	}
}
