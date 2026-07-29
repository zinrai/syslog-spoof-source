package main

import (
	"net"
	"strings"
	"testing"
	"time"
)

// RFC 5424 fixes VERSION at 1. Generators that emit anything else, such as
// Vector's demo_logs and flog, produce messages that a strict receiver is
// entitled to reject. Avoiding that is why this tool exists, so it is asserted
// rather than assumed.
func TestRFC5424VersionIsOne(t *testing.T) {
	now := time.Date(2026, 7, 29, 6, 9, 41, 464000000, time.UTC)
	for seq := uint64(0); seq < 100; seq++ {
		msg := string(buildMessage(FormatRFC5424, net.ParseIP("192.0.2.1"), seq, 134, 0, now))
		if !strings.HasPrefix(msg, "<134>1 ") {
			t.Fatalf("seq %d: VERSION is not 1: %q", seq, msg)
		}
	}
}

// RFC 3164 pads a single digit day with a space, giving "Jul  9" not "Jul 9".
// The layout string uses _2 for this; getting it wrong makes the timestamp
// unparseable to a strict RFC 3164 receiver.
func TestRFC3164DayPadding(t *testing.T) {
	now := time.Date(2026, 7, 9, 6, 9, 41, 0, time.UTC)
	msg := string(buildMessage(FormatRFC3164, net.ParseIP("192.0.2.1"), 0, 134, 0, now))
	if !strings.HasPrefix(msg, "<134>Jul  9 06:09:41 ") {
		t.Fatalf("day is not space-padded: %q", msg)
	}
}
