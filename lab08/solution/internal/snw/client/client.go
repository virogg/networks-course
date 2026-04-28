package client

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/virogg/networks-course/lab08/solution/pkg/snw"
)

type Client struct {
	cfg config
}

func NewClient(opts ...Option) *Client {
	c := new(Client)
	for _, opt := range opts {
		opt(&c.cfg)
	}
	return c
}

func (c *Client) Run(ctx context.Context) error {
	remote, err := net.ResolveUDPAddr("udp",
		fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port))
	if err != nil {
		return fmt.Errorf("resolve remote: %w", err)
	}
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	log.Printf("snw_client local=%s remote=%s (timeout=%s loss=%.2f corrupt=%.2f chunk=%d)",
		conn.LocalAddr(), remote, c.cfg.Timeout, c.cfg.LossProb, c.cfg.CorruptProb, c.cfg.ChunkSize)

	peer := snw.NewPeer(conn, remote, snw.Config{
		Timeout:     c.cfg.Timeout,
		ChunkSize:   c.cfg.ChunkSize,
		LossProb:    c.cfg.LossProb,
		CorruptProb: c.cfg.CorruptProb,
		SendFile:    c.cfg.SendFile,
		RecvFile:    c.cfg.RecvFile,
		IsInitiator: true,
		Seed:        c.cfg.Seed,
		Tag:         "client",
	})

	return peer.Run(ctx)
}
