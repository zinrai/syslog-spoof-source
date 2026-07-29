package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Format string

const (
	FormatRFC3164 Format = "rfc3164"
	FormatRFC5424 Format = "rfc5424"
)

// buildMessage assembles one syslog message.
//
// The hostname and body are placeholders, not an attempt to imitate real
// device logs: the hostname just echoes the source address, and the body
// carries a sequence number the receiving side uses to spot loss and
// reordering.
func buildMessage(f Format, src net.IP, seq uint64, pri int, size int, now time.Time) []byte {
	host := "dev-" + strings.ReplaceAll(src.String(), ".", "-")
	body := fmt.Sprintf("seq=%d src=%s", seq, src)

	var head string
	switch f {
	case FormatRFC5424:
		// <PRI>VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID SP SD [SP MSG]
		head = fmt.Sprintf("<%d>1 %s %s syslog-spoof-source %d - - ",
			pri, now.Format("2006-01-02T15:04:05.000Z07:00"), host, seq%65536)
	default:
		// <PRI>MMM dd HH:mm:ss HOSTNAME TAG[PID]: MSG
		// The layout uses _2 rather than 02: RFC 3164 pads a single digit
		// day with a space, so the second of the month is "Jan  2".
		head = fmt.Sprintf("<%d>%s %s syslog-spoof-source[%d]: ",
			pri, now.Format("Jan _2 15:04:05"), host, seq%65536)
	}

	msg := head + body
	if pad := size - len(msg); pad > 0 {
		msg += " " + strings.Repeat("x", pad-1)
	}
	return []byte(msg)
}
