package server

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
)

const (
	bufSize         = 4096
	lossProbability = 0.2 // 20%
)

type Server struct {
	port int
}

func NewServer(opts ...Option) *Server {
	srv := new(Server)
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

func (s *Server) ListenAndServe() error {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: s.port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	log.Printf("ping server listening on %s", conn.LocalAddr())

	buf := make([]byte, bufSize)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}
		msg := string(buf[:n])

		if rand.Float64() < lossProbability {
			log.Printf("DROP from %s: %q", clientAddr, msg)
			continue
		}

		reply := strings.ToUpper(msg)
		if _, err := conn.WriteToUDP([]byte(reply), clientAddr); err != nil {
			log.Printf("write error to %s: %v", clientAddr, err)
			continue
		}
		log.Printf("echo to %s: %q -> %q", clientAddr, msg, reply)
	}
}
