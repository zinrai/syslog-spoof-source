package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

// sourceIPs returns n addresses taken from the CIDR.
//
// For a /30 or wider the network and broadcast addresses are skipped, so the
// range starts one past the network address. A /31 is a point to point link
// where both addresses are usable (RFC 3021), and a /32 is a single host, so
// neither reserves anything and both start at the network address.
//
// Handing out an address from outside the CIDR would make the tool send from
// a range the operator never named, so the count is checked against what the
// prefix actually holds.
func sourceIPs(cidr string, n int) ([]net.IP, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	if ipnet.IP.To4() == nil {
		return nil, fmt.Errorf("an IPv4 CIDR is required: %s", cidr)
	}
	if n < 1 {
		return nil, fmt.Errorf("at least one address is required, %d requested", n)
	}

	ones, bits := ipnet.Mask.Size()
	total := uint64(1) << uint(bits-ones)

	var offset, available uint64
	if ones >= 31 {
		offset, available = 0, total
	} else {
		offset, available = 1, total-2
	}
	if uint64(n) > available {
		return nil, fmt.Errorf("%s holds %d usable addresses, %d requested", cidr, available, n)
	}

	base := binary.BigEndian.Uint32(ipnet.IP.To4())
	ips := make([]net.IP, 0, n)
	for i := 0; i < n; i++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, base+uint32(offset)+uint32(i))
		ips = append(ips, ip)
	}
	return ips, nil
}
