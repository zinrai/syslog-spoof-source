// Command syslog-spoof-source generates syslog messages from many different
// source addresses and sends them over UDP.
//
// UDP only: TCP needs the three way handshake to complete at the source
// address, so it cannot be spoofed.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Overwritten at release time by goreleaser via -ldflags "-X main.version=...".
// They keep these defaults for a plain go build.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type options struct {
	dst      string
	port     int
	sources  string
	count    int
	rate     float64
	duration time.Duration
	messages uint64
	format   string
	size     int
}

func main() {
	var o options
	flag.StringVar(&o.dst, "dst", "", "destination address (required)")
	flag.IntVar(&o.port, "port", 514, "destination port")
	flag.StringVar(&o.sources, "sources", "10.99.0.0/16", "CIDR the source addresses are taken from")
	flag.IntVar(&o.count, "count", 100, "number of distinct source addresses")
	flag.Float64Var(&o.rate, "rate", 1000, "messages per second, 0 for unthrottled")
	flag.DurationVar(&o.duration, "duration", 10*time.Second, "how long to keep sending")
	flag.Uint64Var(&o.messages, "messages", 0, "total messages to send, takes precedence over -duration")
	flag.StringVar(&o.format, "format", "rfc3164", "message format: rfc3164 or rfc5424")
	flag.IntVar(&o.size, "size", 0, "minimum message size in bytes, padded to reach it")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("syslog-spoof-source %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	if err := run(o); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
