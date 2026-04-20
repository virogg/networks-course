package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	readDeadline = 500 * time.Millisecond
	bufSize      = 4096

	heartbeatFmt = "heartbeat from %s seq=%d oneway=%v missed_since_start=%d loss=%.1f%%"
)

type clientState struct {
	lastSeq  uint64
	lastSeen time.Time
	received uint64
	lost     uint64
}

type Server struct {
	cfg     config
	mu      sync.Mutex
	clients map[string]*clientState
}

func NewServer(opts ...Option) *Server {
	srv := &Server{
		clients: make(map[string]*clientState),
	}
	for _, opt := range opts {
		opt(&srv.cfg)
	}
	return srv
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: s.cfg.Port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	log.Printf("heartbeat server listening on %s (dead=%s, check=%s)", conn.LocalAddr(), s.cfg.DeadAfter, s.cfg.CheckEvery)

	go s.watchDead(ctx)

	buf := make([]byte, bufSize)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
			return err
		}
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := errors.AsType[net.Error](err); ok && ne.Timeout() {
				if ctx.Err() != nil {
					return nil
				}
				continue
			}
			log.Printf("read error: %v", err)
			continue
		}
		s.handlePacket(clientAddr, string(buf[:n]))
	}
}

func (s *Server) handlePacket(addr *net.UDPAddr, msg string) {
	parts := strings.Fields(msg)
	if len(parts) != 2 {
		log.Printf("bad packet from %s: %q", addr, msg)
		return
	}
	seq, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		log.Printf("bad seq from %s: %v", addr, err)
		return
	}
	tsNanos, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		log.Printf("bad ts from %s: %v", addr, err)
		return
	}
	sendTime := time.Unix(0, tsNanos)
	now := time.Now()
	oneWay := now.Sub(sendTime)

	key := addr.String()
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.clients[key]
	if !ok {
		st = new(clientState)
		s.clients[key] = st
		log.Printf("new client %s", key)
	}

	var gap uint64
	if seq > st.lastSeq+1 && st.lastSeq != 0 {
		gap = seq - st.lastSeq - 1
		st.lost += gap
	}
	st.lastSeq = seq
	st.lastSeen = now
	st.received++

	total := st.received + st.lost
	lossPct := 0.0
	if total > 0 {
		lossPct = float64(st.lost) / float64(total) * 100
	}
	log.Printf(heartbeatFmt, key, seq, oneWay, st.lost, lossPct)
	if gap > 0 {
		log.Printf("  -> missed %d packet(s) before seq=%d", gap, seq)
	}
}

func (s *Server) watchDead(ctx context.Context) {
	t := time.NewTicker(s.cfg.CheckEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			s.mu.Lock()
			for key, st := range s.clients {
				if now.Sub(st.lastSeen) > s.cfg.DeadAfter {
					log.Printf("client %s assumed DOWN (no packets for %v)", key, now.Sub(st.lastSeen))
					delete(s.clients, key)
				}
			}
			s.mu.Unlock()
		}
	}
}
