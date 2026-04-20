package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/virogg/networks-course/lab07/solution/pkg/ping"
)

const bufSize = 4096

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

func (c *Client) Run() error {
	serverAddr := &net.UDPAddr{IP: net.ParseIP(c.cfg.Host), Port: c.cfg.Port}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return fmt.Errorf("dial udp: %w", err)
	}
	defer conn.Close()

	fmt.Fprintf(os.Stdout, "PING %s:%d\n", c.cfg.Host, c.cfg.Port)

	var rtts []time.Duration
	received := 0

	buf := make([]byte, bufSize)
	for seq := 1; seq <= c.cfg.Count; seq++ {
		sendTime := time.Now()
		msg := fmt.Sprintf("Ping %d %d", seq, sendTime.UnixNano())

		if _, err := conn.Write([]byte(msg)); err != nil {
			fmt.Printf("seq=%d write error: %v\n", seq, err)
			continue
		}

		if err := conn.SetReadDeadline(sendTime.Add(c.cfg.Timeout)); err != nil {
			return fmt.Errorf("set deadline: %w", err)
		}

		n, err := conn.Read(buf)
		rtt := time.Since(sendTime)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Printf("seq=%d Request timed out\n", seq)
				continue
			}
			fmt.Printf("seq=%d read error: %v\n", seq, err)
			continue
		}

		received++
		rtts = append(rtts, rtt)
		fmt.Printf("seq=%d reply=%q rtt=%v\n", seq, string(buf[:n]), rtt)
	}

	if c.cfg.Stats {
		ping.PrintStats(c.cfg.Host, c.cfg.Count, received, rtts)
	}
	return nil
}
