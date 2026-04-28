package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/virogg/networks-course/lab08/solution/pkg/snw"
)

type Server struct {
	cfg config
}

func NewServer(opts ...Option) *Server {
	s := new(Server)
	for _, opt := range opts {
		opt(&s.cfg)
	}
	return s
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	laddr, err := net.ResolveUDPAddr("udp", s.cfg.Addr)
	if err != nil {
		return fmt.Errorf("resolve addr: %w", err)
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	log.Printf("snw_server listening on %s (timeout=%s loss=%.2f corrupt=%.2f chunk=%d)",
		conn.LocalAddr(), s.cfg.Timeout, s.cfg.LossProb, s.cfg.CorruptProb, s.cfg.ChunkSize)

	peer := snw.NewPeer(conn, nil, snw.Config{
		Timeout:     s.cfg.Timeout,
		ChunkSize:   s.cfg.ChunkSize,
		LossProb:    s.cfg.LossProb,
		CorruptProb: s.cfg.CorruptProb,
		SendFile:    s.cfg.SendFile,
		RecvFile:    s.cfg.RecvFile,
		IsInitiator: false,
		Seed:        s.cfg.Seed,
		Tag:         "server",
	})

	return peer.Run(ctx)
}
