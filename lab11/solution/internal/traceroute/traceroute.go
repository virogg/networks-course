package traceroute

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/ipv4"
)

type Tracer struct {
	cfg config
	id  uint16
	seq uint16
	out io.Writer
}

func New(opts ...Option) *Tracer {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.Queries < 1 {
		cfg.Queries = 1
	}
	return &Tracer{
		cfg: cfg,
		id:  uint16(os.Getpid() & 0xffff),
		out: os.Stdout,
	}
}

type probe struct {
	from net.IP // router/host that answered (nil = no reply)
	rtt  time.Duration
	done bool // destination reached on this probe
}

func (t *Tracer) Run(ctx context.Context) error {
	addr, err := net.ResolveIPAddr("ip4", t.cfg.Host)
	if err != nil {
		return fmt.Errorf("resolve %q: %w", t.cfg.Host, err)
	}

	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("listen icmp (privileged raw socket — run with sudo): %w", err)
	}
	defer conn.Close()
	pc := ipv4.NewPacketConn(conn)

	fmt.Fprintf(t.out, "traceroute to %s (%s), %d hops max, %d probes per hop\n", t.cfg.Host, addr.IP, t.cfg.MaxHops, t.cfg.Queries)

	buf := make([]byte, 1500)
	for ttl := 1; ttl <= t.cfg.MaxHops; ttl++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := pc.SetTTL(ttl); err != nil {
			return fmt.Errorf("set ttl=%d: %w", ttl, err)
		}

		probes, reached := t.traceHop(pc, ttl, addr, buf)
		t.printHop(ttl, probes)
		if reached {
			return nil
		}
	}
	return nil
}

func (t *Tracer) traceHop(pc *ipv4.PacketConn, ttl int, dst *net.IPAddr, buf []byte) ([]probe, bool) {
	probes := make([]probe, t.cfg.Queries)
	sendTime := make([]time.Time, t.cfg.Queries)
	idxOfSeq := make(map[uint16]int, t.cfg.Queries)

	var payload [16]byte
	for i := range probes {
		t.seq++
		sendTime[i] = time.Now()
		n, err := pc.WriteTo(MarshalEcho(t.id, t.seq, payload[:]), nil, dst)
		if err != nil {
			t.logf("ttl=%d seq=%d: send failed: %v", ttl, t.seq, err)
			continue
		}
		idxOfSeq[t.seq] = i
		t.logf("ttl=%d seq=%d: sent %d bytes to %s", ttl, t.seq, n, dst.IP)
	}

	deadline := time.Now().Add(t.cfg.Timeout)
	reached := false
	got := 0
	for pending := len(idxOfSeq); pending > 0; {
		if err := pc.SetReadDeadline(deadline); err != nil {
			break
		}
		n, _, src, err := pc.ReadFrom(buf)
		if err != nil {
			t.logf("ttl=%d: read stopped: %v", ttl, err)
			break
		}
		got++
		msg, err := ParseMessage(buf[:n])
		if err != nil {
			t.logf("ttl=%d: recv %d bytes from %s: parse error: %v", ttl, n, src, err)
			continue
		}
		seq, ok := matchSeq(msg, t.id)
		if !ok {
			t.logf("ttl=%d: recv %d bytes from %s: type=%d code=%d, not our probe", ttl, n, src, msg.Type, msg.Code)
			continue
		}
		i, ok := idxOfSeq[seq]
		if !ok || probes[i].from != nil {
			continue
		}
		ip := src.(*net.IPAddr).IP
		done := msg.Type == TypeEchoReply || (msg.Type == TypeDestinationUnreachable && ip.Equal(dst.IP))
		probes[i] = probe{from: ip, rtt: time.Since(sendTime[i]), done: done}
		reached = reached || done
		pending--
		t.logf("ttl=%d seq=%d: reply from %s type=%d", ttl, seq, ip, msg.Type)
	}
	t.logf("ttl=%d: %d probe(s) sent, %d packet(s) received", ttl, len(idxOfSeq), got)
	return probes, reached
}

func (t *Tracer) logf(format string, args ...any) {
	if t.cfg.Verbose {
		fmt.Fprintf(os.Stderr, "[traceroute] "+format+"\n", args...)
	}
}

func matchSeq(msg *ICMPMessage, id uint16) (uint16, bool) {
	switch msg.Type {
	case TypeEchoReply:
		if msg.ID == id {
			return msg.Seq, true
		}
	case TypeTimeExceeded, TypeDestinationUnreachable:
		if msg.Embedded != nil && msg.Embedded.ID == id {
			return msg.Embedded.Seq, true
		}
	}
	return 0, false
}

func (t *Tracer) printHop(ttl int, probes []probe) {
	var b strings.Builder
	fmt.Fprintf(&b, "%2d  ", ttl)
	var lastIP net.IP
	for _, p := range probes {
		if p.from == nil {
			b.WriteString(" *")
			continue
		}
		if !p.from.Equal(lastIP) {
			if lastIP != nil {
				b.WriteString("\n    ")
			}
			b.WriteString(t.format(p.from))
			lastIP = p.from
		}
		fmt.Fprintf(&b, "  %.3f ms", float64(p.rtt)/float64(time.Millisecond))
	}
	fmt.Fprintln(t.out, b.String())
}

// task Б
func (t *Tracer) format(ip net.IP) string {
	if !t.cfg.Resolve {
		return ip.String()
	}
	names, err := net.LookupAddr(ip.String())
	if err != nil || len(names) == 0 {
		return ip.String()
	}
	return fmt.Sprintf("%s (%s)", strings.TrimSuffix(names[0], "."), ip)
}
