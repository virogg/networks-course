package snw

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

type Config struct {
	Timeout     time.Duration
	ChunkSize   int
	LossProb    float64
	CorruptProb float64
	SendFile    string
	RecvFile    string
	IsInitiator bool // client sends HELLO; servers waits for 1st packet
	Seed        int64
	Tag         string // log prefix ("client"/"server") - for testing
}

// `conn` is expected to be opened already
//
// For client `remote` has to be set beforehand.
// For server `remote` will be opened on first arrived packet addr
type Peer struct {
	cfg Config

	conn   *net.UDPConn
	remote *net.UDPAddr

	remoteMu sync.RWMutex

	ackCh        chan Frame
	dataCh       chan Frame
	remoteReady  chan struct{}
	peerObserved chan struct{}

	rng   *rand.Rand
	rngMu sync.Mutex
}

func NewPeer(conn *net.UDPConn, remote *net.UDPAddr, cfg Config) *Peer {
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = 1024
	}
	if cfg.ChunkSize > MaxPayload {
		log.Printf("[%s] chunk-size %d > MaxPayload %d, clamping",
			cfg.Tag, cfg.ChunkSize, MaxPayload)
		cfg.ChunkSize = MaxPayload
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = time.Second
	}
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}
	if cfg.Tag == "" {
		cfg.Tag = "peer"
	}
	p := &Peer{
		cfg:          cfg,
		conn:         conn,
		remote:       remote,
		ackCh:        make(chan Frame, 16),
		dataCh:       make(chan Frame, 16),
		remoteReady:  make(chan struct{}),
		peerObserved: make(chan struct{}),
		rng:          rand.New(rand.NewSource(cfg.Seed)),
	}
	if cfg.IsInitiator && remote != nil {
		close(p.remoteReady)
	}
	return p
}

func (p *Peer) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-runCtx.Done()
		_ = p.conn.SetReadDeadline(time.Unix(1, 0))
	}()

	var (
		userWg    sync.WaitGroup
		serviceWg sync.WaitGroup
		errMu     sync.Mutex
		firstErr  error
	)
	setErr := func(e error) {
		if e == nil || errors.Is(e, context.Canceled) {
			return
		}
		errMu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		errMu.Unlock()
	}

	serviceWg.Go(func() {
		setErr(p.readLoop(runCtx))
	})

	if p.cfg.IsInitiator {
		serviceWg.Go(func() {
			p.helloRepeater(runCtx)
		})
	}

	if p.cfg.SendFile != "" {
		userWg.Go(func() {
			setErr(p.sender(runCtx))
		})
	}
	if p.cfg.RecvFile != "" {
		userWg.Go(func() {
			setErr(p.receiver(runCtx))
		})
	}

	go func() {
		userWg.Wait()
		cancel()
	}()

	serviceWg.Wait()
	userWg.Wait()
	return firstErr
}

const maxUDPDatagram = 65535 // 2^16-1

func (p *Peer) readLoop(ctx context.Context) error {
	buf := make([]byte, maxUDPDatagram)
	bootstrapped := false
	for {
		n, addr, err := p.conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if isTimeout(err) {
				continue
			}
			return fmt.Errorf("read: %w", err)
		}

		if !bootstrapped {
			if !p.cfg.IsInitiator {
				p.remoteMu.Lock()
				p.remote = addr
				p.remoteMu.Unlock()
				log.Printf("[%s] bootstrap: peer = %s", p.cfg.Tag, addr)
				close(p.remoteReady)
			}
			close(p.peerObserved)
			bootstrapped = true
		} else if !p.cfg.IsInitiator {
			r := p.getRemote()
			if r != nil && !addrEqual(addr, r) {
				log.Printf("[%s] drop frame from unknown %s", p.cfg.Tag, addr)
				continue
			}
		}

		frame, err := Decode(buf[:n])
		if err != nil {
			log.Printf("[%s] drop bad frame: %v", p.cfg.Tag, err)
			continue
		}

		switch frame.Type {
		case FrameAck:
			select {
			case p.ackCh <- frame:
			default:
			}
		case FrameData:
			select {
			case p.dataCh <- frame:
			case <-ctx.Done():
				return nil
			}
		case FrameHello:
		default:
			log.Printf("[%s] unknown frame type=%d", p.cfg.Tag, frame.Type)
		}
	}
}

func (p *Peer) helloRepeater(ctx context.Context) {
	if err := p.sendFrame(Frame{Type: FrameHello}); err != nil {
		log.Printf("[%s] hello send: %v", p.cfg.Tag, err)
	}
	t := time.NewTicker(p.cfg.Timeout)
	defer t.Stop()
	for {
		select {
		case <-p.peerObserved:
			return
		case <-ctx.Done():
			return
		case <-t.C:
			log.Printf("[%s] HELLO retransmit", p.cfg.Tag)
			if err := p.sendFrame(Frame{Type: FrameHello}); err != nil {
				log.Printf("[%s] hello send: %v", p.cfg.Tag, err)
			}
		}
	}
}

func (p *Peer) sender(ctx context.Context) error {
	select {
	case <-p.remoteReady:
	case <-ctx.Done():
		return ctx.Err()
	}

	f, err := os.Open(p.cfg.SendFile)
	if err != nil {
		return fmt.Errorf("open send-file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err == nil {
		log.Printf("[%s] sender: file=%s size=%d chunk=%d",
			p.cfg.Tag, p.cfg.SendFile, info.Size(), p.cfg.ChunkSize)
	}

	buf := make([]byte, p.cfg.ChunkSize)
	seq := uint8(0)
	chunks := 0
	bytesSent := 0
	for {
		n, rerr := io.ReadFull(f, buf)
		isEOF := false
		switch {
		case rerr == nil:
		case errors.Is(rerr, io.EOF):
			n = 0
			isEOF = true
		case errors.Is(rerr, io.ErrUnexpectedEOF):
			isEOF = true
		default:
			return fmt.Errorf("read send-file: %w", rerr)
		}

		payload := make([]byte, n)
		copy(payload, buf[:n])
		flags := uint16(0)
		if isEOF {
			flags |= FlagEOF
		}
		fr := Frame{Type: FrameData, Seq: seq, Flags: flags, Payload: payload}

		if err := p.sendAndAwaitAck(ctx, fr); err != nil {
			return err
		}

		bytesSent += n
		chunks++
		if isEOF {
			log.Printf("[%s] sender: done, %d chunks, %d bytes", p.cfg.Tag, chunks, bytesSent)
			return nil
		}
		seq ^= 1
	}
}

func (p *Peer) sendAndAwaitAck(ctx context.Context, fr Frame) error {
	for attempt := 1; ; attempt++ {
		if err := p.sendFrame(fr); err != nil {
			return fmt.Errorf("send: %w", err)
		}

		t := time.NewTimer(p.cfg.Timeout)
		waitOk := false
	wait:
		for {
			select {
			case ack := <-p.ackCh:
				if ack.Seq == fr.Seq {
					waitOk = true
					t.Stop()
					break wait
				}
				log.Printf("[%s] sender: drop ACK seq=%d (want %d)",
					p.cfg.Tag, ack.Seq, fr.Seq)
			case <-t.C:
				break wait
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			}
		}
		if waitOk {
			return nil
		}
		log.Printf("[%s] sender: timeout seq=%d, retransmit (attempt %d)",
			p.cfg.Tag, fr.Seq, attempt+1)
	}
}

func (p *Peer) receiver(ctx context.Context) error {
	select {
	case <-p.remoteReady:
	case <-ctx.Done():
		return ctx.Err()
	}

	f, err := os.Create(p.cfg.RecvFile)
	if err != nil {
		return fmt.Errorf("create recv-file: %w", err)
	}
	defer f.Close()

	expected := uint8(0)
	bytesWritten := 0
	for {
		select {
		case fr := <-p.dataCh:
			if fr.Type != FrameData {
				continue
			}
			if fr.Seq == expected {
				if len(fr.Payload) > 0 {
					if _, err := f.Write(fr.Payload); err != nil {
						return fmt.Errorf("write recv-file: %w", err)
					}
					bytesWritten += len(fr.Payload)
				}
				if err := p.sendFrame(Frame{Type: FrameAck, Seq: fr.Seq}); err != nil {
					return fmt.Errorf("send ack: %w", err)
				}
				if fr.HasEOF() {
					log.Printf("[%s] receiver: EOF, %d bytes saved to %s",
						p.cfg.Tag, bytesWritten, p.cfg.RecvFile)
					return nil
				}
				expected ^= 1
			} else {
				log.Printf("[%s] receiver: duplicate seq=%d (want %d), re-ACK",
					p.cfg.Tag, fr.Seq, expected)
				_ = p.sendFrame(Frame{Type: FrameAck, Seq: fr.Seq})
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Peer) sendFrame(fr Frame) error {
	r := p.getRemote()
	if r == nil {
		return errors.New("remote address unknown")
	}

	raw := fr.Encode()

	if p.cfg.CorruptProb > 0 && fr.Type == FrameData && len(fr.Payload) > 0 {
		if p.draw() < p.cfg.CorruptProb {
			idx := HeaderSize + p.intn(len(fr.Payload))
			raw[idx] ^= 0x01
			log.Printf("[%s] CORRUPT: flipped bit at byte %d of seq=%d",
				p.cfg.Tag, idx-HeaderSize, fr.Seq)
		}
	}

	if p.cfg.LossProb > 0 && p.draw() < p.cfg.LossProb {
		log.Printf("[%s] LOSS: drop %s seq=%d", p.cfg.Tag, fr.Type, fr.Seq)
		return nil
	}

	if _, err := p.conn.WriteToUDP(raw, r); err != nil {
		return err
	}
	return nil
}

func (p *Peer) getRemote() *net.UDPAddr {
	p.remoteMu.RLock()
	defer p.remoteMu.RUnlock()
	return p.remote
}

func (p *Peer) draw() float64 {
	p.rngMu.Lock()
	defer p.rngMu.Unlock()
	return p.rng.Float64()
}

func (p *Peer) intn(n int) int {
	p.rngMu.Lock()
	defer p.rngMu.Unlock()
	return p.rng.Intn(n)
}

func addrEqual(a, b *net.UDPAddr) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.IP.Equal(b.IP) && a.Port == b.Port
}

func isTimeout(err error) bool {
	if ne, ok := errors.AsType[net.Error](err); ok {
		return ne.Timeout()
	}
	return false
}
