package main

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

// rawSender writes packets to a raw socket with IP_HDRINCL enabled.
//
// Not a plain UDP socket: the kernel picks the source address there, and
// picking it is the entire point of this tool. IP_HDRINCL hands the IP header
// over to the caller while leaving routing, ARP and the link layer to the
// kernel, which is why nothing here deals with MAC addresses.
//
// Requires CAP_NET_RAW. Run as root, or grant the binary the capability with
// setcap cap_net_raw+ep.
type rawSender struct {
	fd  int
	dst unix.SockaddrInet4
}

func newRawSender(dst net.IP) (*rawSender, error) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_RAW)
	if err != nil {
		return nil, fmt.Errorf("cannot open a raw socket, CAP_NET_RAW is required: %w", err)
	}
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_HDRINCL, 1); err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("cannot set IP_HDRINCL: %w", err)
	}

	s := &rawSender{fd: fd}
	copy(s.dst.Addr[:], dst.To4())
	return s, nil
}

func (s *rawSender) send(pkt []byte) error {
	return unix.Sendto(s.fd, pkt, 0, &s.dst)
}

func (s *rawSender) Close() error { return unix.Close(s.fd) }
