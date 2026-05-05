package copies

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"maps"
	"net"
	"slices"
	"sync"
	"time"
)

const bufSize = 1024

type Peer struct {
	ID       string
	LastSeen time.Time
}

type Snapshot struct {
	Self  string
	Peers []Peer
}

type App struct {
	cfg config

	id        string
	conn      net.PacketConn
	bcastAddr *net.UDPAddr

	mu    sync.Mutex
	peers map[string]*Peer
}

func (a *App) Interval() time.Duration { return a.cfg.Interval }

func (a *App) Snapshot() Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := Snapshot{Self: a.id}
	for _, id := range slices.Sorted(maps.Keys(a.peers)) {
		s.Peers = append(s.Peers, *a.peers[id])
	}
	return s
}

func NewApp(opts ...Option) *App {
	a := &App{peers: make(map[string]*Peer)}
	for _, opt := range opts {
		opt(&a.cfg)
	}
	if a.cfg.Port == 0 {
		a.cfg.Port = 9999
	}
	if a.cfg.Interval == 0 {
		a.cfg.Interval = 2 * time.Second
	}
	if a.cfg.DeadMultiplier == 0 {
		a.cfg.DeadMultiplier = 3
	}
	return a
}

func (a *App) Run(ctx context.Context) error {
	if err := a.setup(ctx); err != nil {
		return err
	}
	defer a.conn.Close() //nolint:errcheck

	log.Printf("started: id=%s broadcast=%s interval=%s dead-after=%s",
		a.id, a.bcastAddr, a.cfg.Interval, a.deadAfter())

	if err := a.send(Message{Kind: KindHello, ID: a.id}); err != nil {
		log.Printf("send HELLO: %v", err)
	}

	var wg sync.WaitGroup
	wg.Go(func() { a.readLoop(ctx) })
	wg.Go(func() { a.tickLoop(ctx, a.sendAlive) })
	wg.Go(func() { a.tickLoop(ctx, a.reap) })
	wg.Go(func() { a.tickLoop(ctx, a.printList) })

	<-ctx.Done()

	a.shutdown()
	_ = a.conn.SetReadDeadline(time.Now())
	wg.Wait()
	return nil
}

func (a *App) setup(ctx context.Context) error {
	localIP, err := pickLocalIP()
	if err != nil {
		return err
	}
	tag, err := randTag()
	if err != nil {
		return fmt.Errorf("random tag: %w", err)
	}
	a.id = fmt.Sprintf("%s:%d", localIP, tag)

	lc := net.ListenConfig{Control: bindOpts}
	pc, err := lc.ListenPacket(ctx, "udp4", fmt.Sprintf(":%d", a.cfg.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	a.conn = pc
	a.bcastAddr = &net.UDPAddr{IP: net.IPv4bcast, Port: a.cfg.Port}
	return nil
}

func (a *App) send(m Message) error {
	_, err := a.conn.WriteTo(m.Encode(), a.bcastAddr)
	return err
}

func (a *App) readLoop(ctx context.Context) {
	buf := make([]byte, bufSize)
	for {
		if ctx.Err() != nil {
			return
		}
		n, _, err := a.conn.ReadFrom(buf)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if nerr, ok := errors.AsType[net.Error](err); ok && nerr.Timeout() {
				continue
			}
			log.Printf("read: %v", err)
			return
		}
		msg, err := Decode(buf[:n])
		if err != nil {
			log.Printf("decode: %v", err)
			continue
		}
		if msg.ID == a.id {
			continue
		}
		if a.handle(msg) {
			if err := a.send(Message{Kind: KindAlive, ID: a.id}); err != nil {
				log.Printf("reply ALIVE: %v", err)
			}
		}
	}
}

func (a *App) handle(m Message) (replyAlive bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch m.Kind {
	case KindBye:
		if _, ok := a.peers[m.ID]; ok {
			delete(a.peers, m.ID)
			log.Printf("peer LEFT: %s", m.ID)
		}
	case KindHello, KindAlive:
		if p, ok := a.peers[m.ID]; ok {
			p.LastSeen = time.Now()
		} else {
			a.peers[m.ID] = &Peer{ID: m.ID, LastSeen: time.Now()}
			log.Printf("peer JOINED: %s", m.ID)
		}
		replyAlive = m.Kind == KindHello
	}
	return replyAlive
}

func (a *App) tickLoop(ctx context.Context, fn func()) {
	t := time.NewTicker(a.cfg.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			fn()
		}
	}
}

func (a *App) sendAlive() {
	if err := a.send(Message{Kind: KindAlive, ID: a.id}); err != nil {
		log.Printf("send ALIVE: %v", err)
	}
}

func (a *App) reap() {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	maps.DeleteFunc(a.peers, func(_ string, p *Peer) bool {
		if now.Sub(p.LastSeen) > a.deadAfter() {
			log.Printf("peer TIMEOUT: %s", p.ID)
			return true
		}
		return false
	})
}

func (a *App) printList() {
	a.mu.Lock()
	defer a.mu.Unlock()
	ids := slices.Sorted(maps.Keys(a.peers))
	log.Printf("active copies: %d (incl self)", len(ids)+1)
	log.Printf("  self: %s", a.id)
	for _, id := range ids {
		log.Printf("  peer: %s", id)
	}
}

func (a *App) shutdown() {
	for range 2 {
		if err := a.send(Message{Kind: KindBye, ID: a.id}); err != nil {
			log.Printf("send BYE: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (a *App) deadAfter() time.Duration {
	return a.cfg.Interval * time.Duration(a.cfg.DeadMultiplier)
}

func randTag() (uint16, error) {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b[:]), nil
}

func pickLocalIP() (net.IP, error) {
	c, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		return nil, fmt.Errorf("routing lookup: %w", err)
	}
	defer c.Close() //nolint:errcheck
	return c.LocalAddr().(*net.UDPAddr).IP, nil
}
