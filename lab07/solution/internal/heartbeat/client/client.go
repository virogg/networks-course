package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

type Client struct {
	cfg config
}

func NewClient(opts ...Option) *Client {
	client := new(Client)
	for _, opt := range opts {
		opt(&client.cfg)
	}
	return client
}

func (client *Client) Run(ctx context.Context) error {
	serverAddr := &net.UDPAddr{IP: net.ParseIP(client.cfg.Host), Port: client.cfg.Port}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return fmt.Errorf("dial udp: %w", err)
	}
	defer conn.Close()

	log.Printf("heartbeat client -> %s every %s", serverAddr, client.cfg.Interval)

	ticker := time.NewTicker(client.cfg.Interval)
	defer ticker.Stop()

	seq := uint64(1)
	send := func() {
		msg := fmt.Sprintf("%d %d", seq, time.Now().UnixNano())
		if _, err := conn.Write([]byte(msg)); err != nil {
			log.Printf("seq=%d write error: %v", seq, err)
			return
		}
		log.Printf("sent: %s", msg)
		seq++
	}

	send()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			send()
		}
	}
}
