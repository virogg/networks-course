package ping

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/net/ipv4"
)

const (
	pingHeaderFmt     = "PING %s (%s): %d data bytes\n"
	echoReplyFmt      = "%d bytes from %s: icmp_seq=%d ttl=- time=%.3f ms\n"
	icmpErrFmt        = "From %s: icmp_seq=%d %s (type=%d code=%d)\n"
	sendFailedFmt     = "send seq=%d failed: %v\n"
	requestTimeoutFmt = "Request timeout for icmp_seq=%d\n"
	readErrFmt        = "read: %v\n"
	deadlineErrFmt    = "set deadline: %v\n"
)

type Pinger struct {
	cfg   config
	id    uint16
	stats Stats
	out   io.Writer
}

func New(opts ...Option) *Pinger {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.Size < 8 {
		cfg.Size = 56
	}
	return &Pinger{
		cfg: cfg,
		id:  uint16(os.Getpid() & 0xffff),
		out: os.Stdout,
	}
}

func (p *Pinger) Stats() *Stats { return &p.stats }

func (p *Pinger) Run(ctx context.Context) error {
	addr, err := net.ResolveIPAddr("ip4", p.cfg.Host)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("listen icmp (privileged raw socket — run with sudo): %w", err)
	}
	defer conn.Close()
	pc := ipv4.NewPacketConn(conn)

	fmt.Fprintf(p.out, pingHeaderFmt, p.cfg.Host, addr.IP, p.cfg.Size)

	ticker := time.NewTicker(p.cfg.Interval)
	defer ticker.Stop()

	var seq uint16
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		seq++
		p.stats.Sent++
		sendTime := time.Now()
		if err := p.send(pc, addr, seq, sendTime); err != nil {
			fmt.Fprintf(p.out, sendFailedFmt, seq, err)
		} else {
			p.recv(pc, seq, sendTime)
		}

		if p.cfg.Count > 0 && p.stats.Sent >= p.cfg.Count {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (p *Pinger) send(pc *ipv4.PacketConn, addr *net.IPAddr, seq uint16, t time.Time) error {
	payload := make([]byte, p.cfg.Size)
	binary.BigEndian.PutUint64(payload[:8], uint64(t.UnixNano()))
	for i := 8; i < len(payload); i++ {
		payload[i] = byte(i)
	}
	pkt := MarshalEcho(p.id, seq, payload)
	_, err := pc.WriteTo(pkt, nil, addr)
	return err
}

func (p *Pinger) recv(pc *ipv4.PacketConn, seq uint16, sendTime time.Time) {
	buf := make([]byte, 1500)
	deadline := sendTime.Add(p.cfg.Timeout)
	for {
		if err := pc.SetReadDeadline(deadline); err != nil {
			fmt.Fprintf(p.out, deadlineErrFmt, err)
			return
		}
		n, _, src, err := pc.ReadFrom(buf)
		if err != nil {
			if ne, ok := errors.AsType[net.Error](err); ok && ne.Timeout() {
				fmt.Fprintf(p.out, requestTimeoutFmt, seq)
				return
			}
			fmt.Fprintf(p.out, readErrFmt, err)
			return
		}
		msg, err := ParseMessage(buf[:n])
		if err != nil {
			continue
		}
		switch msg.Type {
		case TypeEchoReply:
			if msg.ID != p.id || msg.Seq != seq {
				continue
			}
			rtt := computeRTT(msg.Payload, sendTime)
			p.stats.Record(rtt)
			fmt.Fprintf(p.out, echoReplyFmt, n, src, seq, float64(rtt)/float64(time.Millisecond))
			return
		case TypeDestinationUnreachable, TypeTimeExceeded:
			if msg.Embedded == nil || msg.Embedded.ID != p.id || msg.Embedded.Seq != seq {
				continue
			}
			desc := DescribeError(msg.Type, msg.Code)
			fmt.Fprintf(p.out, icmpErrFmt, src, seq, desc, msg.Type, msg.Code)
			return
		}
	}
}

func computeRTT(payload []byte, sendTime time.Time) time.Duration {
	if len(payload) >= 8 {
		ts := int64(binary.BigEndian.Uint64(payload[:8]))
		return time.Since(time.Unix(0, ts))
	}
	return time.Since(sendTime)
}
