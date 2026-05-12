package gbn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"os"
	"time"
)

type Client struct {
	cfg  clientConfig
	conn *net.UDPConn
	log  *EventLogger

	base    uint32
	nextSeq uint32
	chunks  [][]byte

	timer        *time.Timer
	ackCh        chan Packet
	doneCh       chan error
	finAcked     bool
	timeoutCount int
}

const (
	pollInterval = 200 * time.Millisecond
	maxTimeouts  = 20

	startFmt          = "START file=%s chunks=%d window=%d timeout=%s"
	doneFmt           = "DONE all data + FIN acked"
	sendLostFmt       = "%s %s seq=%d (LOST simulated)"
	sendOkFmt         = "%s %s seq=%d"
	writeErrFmt       = "write error: %v"
	timeoutFmt        = "TIMEOUT base=%d -> resend %s (attempt %d/%d)"
	timeoutAbortFmt   = "gbn: aborting after %d consecutive timeouts (peer unresponsive)"
	recvBadFmt        = "RECV bad packet: %v"
	recvAckFmt        = "RECV ACK seq=%d (cumulative)"
	recvAckDupFmt     = "RECV ACK seq=%d (duplicate, ignored)"
	recvFinAckFmt     = "RECV FIN-ACK seq=%d"
	recvUnexpectedFmt = "RECV unexpected %s seq=%d"
	windowStateFmt    = "window=[base=%d next=%d size=%d] acked=%s inflight=%s pending=%s"
)

func NewClient(opts ...ClientOption) (*Client, error) {
	cfg := defaultClientConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	addr, err := net.ResolveUDPAddr("udp", cfg.RemoteAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	if cfg.Logger == nil {
		cfg.Logger = NewLogger(os.Stderr, "client")
	}
	return &Client{
		cfg:    cfg,
		conn:   conn,
		log:    cfg.Logger,
		ackCh:  make(chan Packet, 16),
		doneCh: make(chan error, 1),
	}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) finSeq() uint32 { return uint32(len(c.chunks)) }

func (c *Client) loadFile() error {
	f, err := os.Open(c.cfg.File)
	if err != nil {
		return err
	}
	defer f.Close()
	for {
		buf := make([]byte, c.cfg.ChunkSize)
		n, err := io.ReadFull(f, buf)
		if n > 0 {
			c.chunks = append(c.chunks, buf[:n])
		}
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Run(ctx context.Context) error {
	if err := c.loadFile(); err != nil {
		return err
	}
	c.log.Event(
		fmt.Sprintf(startFmt, c.cfg.File, len(c.chunks), c.cfg.WindowSize, c.cfg.Timeout),
		c.windowState(),
	)

	c.timer = time.NewTimer(c.cfg.Timeout)
	c.timer.Stop()

	go c.recvLoop(ctx)

	c.fillWindow()

	for !c.finAcked {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-c.doneCh:
			if err != nil {
				return err
			}
		case <-c.timer.C:
			if err := c.handleTimeout(); err != nil {
				return err
			}
		case pkt := <-c.ackCh:
			c.handleAck(pkt)
			c.fillWindow()
		}
	}
	c.log.Event(doneFmt, c.windowState())
	return nil
}

func (c *Client) fillWindow() {
	for c.nextSeq < c.base+c.cfg.WindowSize && c.nextSeq <= c.finSeq() {
		seq := c.nextSeq
		c.nextSeq++
		c.sendSeq(seq, false)
	}
	if c.base < c.nextSeq {
		c.timer.Reset(c.cfg.Timeout)
	}
}

func (c *Client) sendSeq(seq uint32, isResend bool) {
	var p Packet
	if seq == c.finSeq() {
		p = Packet{Type: TypeFin, Seq: seq}
	} else {
		p = Packet{Type: TypeData, Seq: seq, Payload: c.chunks[seq]}
	}
	tag := "SEND"
	if isResend {
		tag = "RESEND"
	}
	if c.cfg.LossRate > 0 && rand.Float64() < c.cfg.LossRate {
		c.log.Event(fmt.Sprintf(sendLostFmt, tag, p.Type, seq), c.windowState())
		return
	}
	if _, err := c.conn.Write(p.Marshal()); err != nil {
		c.log.Event(fmt.Sprintf(writeErrFmt, err), c.windowState())
		return
	}
	c.log.Event(fmt.Sprintf(sendOkFmt, tag, p.Type, seq), c.windowState())
}

func (c *Client) handleTimeout() error {
	if c.base >= c.nextSeq {
		return nil
	}
	c.timeoutCount++
	if c.timeoutCount > maxTimeouts {
		return fmt.Errorf(timeoutAbortFmt, maxTimeouts)
	}
	c.log.Event(
		fmt.Sprintf(timeoutFmt, c.base, formatRange(c.base, c.nextSeq), c.timeoutCount, maxTimeouts),
		c.windowState(),
	)
	for seq := c.base; seq < c.nextSeq; seq++ {
		c.sendSeq(seq, true)
	}
	c.timer.Reset(c.cfg.Timeout)
	return nil
}

func (c *Client) recvLoop(ctx context.Context) {
	buf := make([]byte, 64*1024)
	for {
		if err := c.conn.SetReadDeadline(time.Now().Add(pollInterval)); err != nil {
			c.doneCh <- err
			return
		}
		n, err := c.conn.Read(buf)
		if err != nil {
			if ne, ok := errors.AsType[net.Error](err); ok && ne.Timeout() {
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}
			return
		}
		pkt, err := Unmarshal(buf[:n])
		if err != nil {
			c.log.Event(fmt.Sprintf(recvBadFmt, err), c.windowState())
			continue
		}
		select {
		case c.ackCh <- pkt:
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) handleAck(pkt Packet) {
	switch pkt.Type {
	case TypeAck:
		if pkt.Seq+1 > c.base {
			c.base = pkt.Seq + 1
			c.timeoutCount = 0
			c.log.Event(fmt.Sprintf(recvAckFmt, pkt.Seq), c.windowState())
			c.timer.Stop()
			if c.base != c.nextSeq {
				c.timer.Reset(c.cfg.Timeout)
			}
		} else {
			c.log.Event(fmt.Sprintf(recvAckDupFmt, pkt.Seq), c.windowState())
		}
	case TypeFinAck:
		c.log.Event(fmt.Sprintf(recvFinAckFmt, pkt.Seq), c.windowState())
		c.finAcked = true
		c.timer.Stop()
	default:
		c.log.Event(fmt.Sprintf(recvUnexpectedFmt, pkt.Type, pkt.Seq), c.windowState())
	}
}

func (c *Client) windowState() string {
	return fmt.Sprintf(windowStateFmt,
		c.base, c.nextSeq, c.cfg.WindowSize,
		formatRange(0, c.base),
		formatRange(c.base, c.nextSeq),
		formatRange(c.nextSeq, c.finSeq()+1),
	)
}
