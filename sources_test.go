package main

import (
	"net"
	"testing"
)

// Every address handed out has to fall inside the CIDR. Handing out one from
// outside would make the tool send from a range the operator never named.
// The /31 and /32 cases matter: an earlier version returned an address past
// the end of a /32.
func TestSourceIPsStayInsideCIDR(t *testing.T) {
	for _, c := range []struct {
		cidr  string
		count int
	}{
		{"192.0.2.0/24", 254},
		{"192.0.2.0/25", 126},
		{"10.0.0.0/30", 2},
		{"10.0.0.0/31", 2},
		{"10.0.0.1/32", 1},
	} {
		ips, err := sourceIPs(c.cidr, c.count)
		if err != nil {
			t.Errorf("%s count %d: %v", c.cidr, c.count, err)
			continue
		}
		_, ipnet, _ := net.ParseCIDR(c.cidr)
		for _, ip := range ips {
			if !ipnet.Contains(ip) {
				t.Errorf("%s count %d: %s falls outside the range", c.cidr, c.count, ip)
				break
			}
		}
	}
}

// Asking for more than the range holds is an error, not a silent overflow
// into the neighbouring network.
func TestSourceIPsRejectsOverflow(t *testing.T) {
	for _, c := range []struct {
		cidr  string
		count int
	}{
		{"192.0.2.0/24", 255},
		{"10.0.0.0/30", 3},
		{"10.0.0.1/32", 2},
	} {
		if _, err := sourceIPs(c.cidr, c.count); err == nil {
			t.Errorf("%s count %d: expected an error", c.cidr, c.count)
		}
	}
}
