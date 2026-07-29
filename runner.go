package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/unix"
)

type config struct {
	dst    net.IP
	srcs   []net.IP
	format Format
}

func run(o options) error {
	cfg, err := o.validate()
	if err != nil {
		return err
	}

	sender, err := newRawSender(cfg.dst)
	if err != nil {
		return err
	}
	defer sender.Close()

	sent, elapsed := o.sendLoop(cfg, sender)

	fmt.Fprintf(os.Stderr, "sent=%d sources=%d elapsed=%.2fs rate=%.0f/s dst=%s:%d\n",
		sent, len(cfg.srcs), elapsed.Seconds(), float64(sent)/elapsed.Seconds(), o.dst, o.port)
	return nil
}

func (o options) validate() (config, error) {
	var c config

	if o.dst == "" {
		return c, fmt.Errorf("-dst is required")
	}
	c.dst = net.ParseIP(o.dst)
	if c.dst == nil || c.dst.To4() == nil {
		return c, fmt.Errorf("-dst must be an IPv4 address: %s", o.dst)
	}

	c.format = Format(o.format)
	if c.format != FormatRFC3164 && c.format != FormatRFC5424 {
		return c, fmt.Errorf("-format must be rfc3164 or rfc5424: %s", o.format)
	}

	srcs, err := sourceIPs(o.sources, o.count)
	if err != nil {
		return c, err
	}
	c.srcs = srcs
	return c, nil
}

func (o options) sendLoop(c config, sender *rawSender) (uint64, time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, unix.SIGTERM)

	deadline := time.Time{}
	if o.messages == 0 {
		deadline = time.Now().Add(o.duration)
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	start := time.Now()
	var sent uint64

	for {
		select {
		case <-stop:
			return sent, time.Since(start)
		default:
		}
		if o.messages > 0 && sent >= o.messages {
			break
		}
		if o.messages == 0 && !time.Now().Before(deadline) {
			break
		}

		o.pace(sent, start)

		src := c.srcs[int(sent%uint64(len(c.srcs)))]
		// Real devices differ in facility and severity, so both are drawn
		// per message: facility 0-23, severity 0-7. This gives the receiving
		// side a realistic spread to filter on, e.g. severity = "err".
		pri := rnd.Intn(24)*8 + rnd.Intn(8)
		msg := buildMessage(c.format, src, sent, pri, o.size, time.Now())
		sport := uint16(1024 + rnd.Intn(65535-1024))
		pkt, err := buildPacket(src, c.dst, sport, uint16(o.port), uint16(sent&0xffff), msg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "build failed:", err)
			break
		}
		if err := sender.send(pkt); err != nil {
			fmt.Fprintf(os.Stderr, "send failed on message %d: %v\n", sent+1, err)
			break
		}
		sent++
	}
	return sent, time.Since(start)
}

// pace holds the average rate without a ticker: ticker resolution bottoms out
// around a millisecond, which caps throughput near 1000 per second. Comparing
// against a deadline lets the wait go negative at high rates, so sending
// continues without pausing and the average still holds.
func (o options) pace(sent uint64, start time.Time) {
	if o.rate <= 0 {
		return
	}
	target := time.Duration(float64(sent) / o.rate * float64(time.Second))
	if behind := target - time.Since(start); behind > 200*time.Microsecond {
		time.Sleep(behind)
	}
}
