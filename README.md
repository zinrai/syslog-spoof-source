# syslog-spoof-source

Generate **UDP** syslog from many different source addresses.

```
$ syslog-spoof-source -dst 10.0.0.5 -sources 10.99.0.0/16 -count 12000 -rate 3000 -duration 60s
sent=180000 sources=12000 elapsed=60.00s rate=3000/s dst=10.0.0.5:514
```

## What it is for

When testing a syslog receiver, rate and connection count are already covered by `loggen`, which ships with syslog-ng. What no tool seems to offer is **varying the source address**.

That matters. A receiver that keys on the sender behaves differently under many distinct sources than under one, and a single load generator only ever produces one.

This covers that one gap. It is not a replacement for `loggen`.

## Limits

**Keep it inside your own network.** Spoofed packets are dropped by reverse-path filtering on the way, so put the generator on the same segment as the receiver.

**Mind the receiver's own limits.** Many distinct sources can hit defaults a single source never reaches, so when logs are sent but never arrive, check the receiver's tuning before suspecting the sender.

## Usage

Run with `-h` for the full flag list. The common shapes:

```bash
# Many sources at a steady rate, for as long as you watch
$ syslog-spoof-source -dst 10.0.0.5 -sources 10.99.0.0/16 -count 12000 -rate 3000

# A fixed number of messages, to reconcile sent against received
$ syslog-spoof-source -dst 10.0.0.5 -count 12000 -messages 36000 -rate 3000

# RFC 5424 instead of RFC 3164, sized to a target bandwidth rather than a rate
$ syslog-spoof-source -dst 10.0.0.5 -format rfc5424 -rate 3000 -size 1500
```

Source addresses are taken in order starting one past the network address, so `-sources 10.99.0.0/16 -count 12000` covers `10.99.0.1` through `10.99.46.224`. Facility and severity are not flags: each message draws a random facility (0-23) and severity (0-7), so the receiving side sees a realistic spread rather than one constant value.

## What the messages look like

```
# rfc3164
<134>Jul 29 03:01:41 dev-10-99-0-1 syslog-spoof-source[0]: seq=0 src=10.99.0.1

# rfc5424
<134>1 2026-07-29T03:01:41.464Z dev-10-99-0-1 syslog-spoof-source 0 - - seq=0 src=10.99.0.1
```

The hostname just echoes the source address, so the receiving side can see at a glance where a message came from. It is not meant to imitate a real device. The `seq` value counts every message sent, so the receiving side can use it to detect loss and reordering. Because sources cycle through the list, one address sees `seq` values spaced `-count` apart.

## License

This project is licensed under the [MIT License](LICENSE).
