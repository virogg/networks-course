package gbn

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"time"
)

const (
	timeWaitDuration = 5 * time.Second

	listenFmt            = "LISTEN %s -> %s loss=%.2f"
	closeTimeWaitFmt     = "CLOSE time-wait elapsed"
	recvBadServerFmt     = "RECV bad packet: %v"
	recvDataInOrderFmt   = "RECV DATA seq=%d (in-order, %d bytes)"
	recvDataOOOFmt       = "RECV DATA seq=%d (out-of-order, expected=%d, drop)"
	recvDataAfterFinFmt  = "RECV DATA seq=%d after FIN (drop)"
	recvFinFmt           = "RECV FIN seq=%d"
	recvFinDupFmt        = "RECV FIN seq=%d (duplicate, replay FIN-ACK)"
	recvFinDropFmt       = "RECV FIN seq=%d (expected=%d, drop)"
	recvUnexpectedSrvFmt = "RECV unexpected %s seq=%d"
	sendLostSrvFmt       = "SEND %s seq=%d (LOST simulated)"
	sendOkSrvFmt         = "SEND %s seq=%d"
	writeErrSrvFmt       = "write error: %v"
	stateFmt             = "expected=%d"
)

type Server struct {
	cfg  serverConfig
	log  *EventLogger
	conn *net.UDPConn
	out  *os.File

	expected    uint32
	ackSent     bool
	lastAck     uint32
	clientAddr  *net.UDPAddr
	finSeq      uint32
	finReceived bool
	twDeadline  time.Time
}

func NewServer(opts ...ServerOption) (*Server, error) {
	cfg := defaultServerConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	addr, err := net.ResolveUDPAddr("udp", cfg.ListenAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	if cfg.Logger == nil {
		cfg.Logger = NewLogger(os.Stderr, "server")
	}
	return &Server{cfg: cfg, log: cfg.Logger, conn: conn}, nil
}

func (s *Server) Close() error { return s.conn.Close() }

func (s *Server) Run(ctx context.Context) error {
	out, err := os.Create(s.cfg.OutPath)
	if err != nil {
		return err
	}
	defer out.Close()
	s.out = out

	s.log.Event(
		fmt.Sprintf(listenFmt, s.conn.LocalAddr(), s.cfg.OutPath, s.cfg.LossRate),
		s.state(),
	)

	buf := make([]byte, 64*1024)
	for {
		readDeadline := time.Now().Add(pollInterval)
		if s.finReceived && s.twDeadline.Before(readDeadline) {
			readDeadline = s.twDeadline
		}
		if err := s.conn.SetReadDeadline(readDeadline); err != nil {
			return err
		}
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := errors.AsType[net.Error](err); ok && ne.Timeout() {
				if s.finReceived && !time.Now().Before(s.twDeadline) {
					s.log.Event(closeTimeWaitFmt, s.state())
					return nil
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					continue
				}
			}
			return err
		}
		if s.clientAddr == nil {
			s.clientAddr = addr
		}
		pkt, err := Unmarshal(buf[:n])
		if err != nil {
			s.log.Event(fmt.Sprintf(recvBadServerFmt, err), s.state())
			continue
		}

		switch pkt.Type {
		case TypeData:
			if werr := s.handleData(pkt); werr != nil {
				return werr
			}
		case TypeFin:
			s.handleFin(pkt)
		default:
			s.log.Event(fmt.Sprintf(recvUnexpectedSrvFmt, pkt.Type, pkt.Seq), s.state())
		}
	}
}

func (s *Server) handleData(pkt Packet) error {
	if s.finReceived {
		s.log.Event(fmt.Sprintf(recvDataAfterFinFmt, pkt.Seq), s.state())
		s.replayAck()
		return nil
	}
	if pkt.Seq != s.expected {
		s.log.Event(fmt.Sprintf(recvDataOOOFmt, pkt.Seq, s.expected), s.state())
		s.replayAck()
		return nil
	}
	if _, err := s.out.Write(pkt.Payload); err != nil {
		return err
	}
	s.expected++
	s.log.Event(fmt.Sprintf(recvDataInOrderFmt, pkt.Seq, len(pkt.Payload)), s.state())
	s.lastAck = pkt.Seq
	s.ackSent = true
	s.sendAck(pkt.Seq, TypeAck)
	return nil
}

func (s *Server) handleFin(pkt Packet) {
	if s.finReceived && pkt.Seq == s.finSeq {
		s.log.Event(fmt.Sprintf(recvFinDupFmt, pkt.Seq), s.state())
		s.sendAck(pkt.Seq, TypeFinAck)
		s.twDeadline = time.Now().Add(timeWaitDuration)
		return
	}
	if pkt.Seq != s.expected {
		s.log.Event(fmt.Sprintf(recvFinDropFmt, pkt.Seq, s.expected), s.state())
		s.replayAck()
		return
	}
	s.expected++
	s.finSeq = pkt.Seq
	s.finReceived = true
	s.twDeadline = time.Now().Add(timeWaitDuration)
	s.log.Event(fmt.Sprintf(recvFinFmt, pkt.Seq), s.state())
	s.sendAck(pkt.Seq, TypeFinAck)
}

func (s *Server) replayAck() {
	if !s.ackSent {
		return
	}
	s.sendAck(s.lastAck, TypeAck)
}

func (s *Server) sendAck(seq uint32, t Type) {
	if s.cfg.LossRate > 0 && rand.Float64() < s.cfg.LossRate {
		s.log.Event(fmt.Sprintf(sendLostSrvFmt, t, seq), s.state())
		return
	}
	p := Packet{Type: t, Seq: seq}
	if _, err := s.conn.WriteToUDP(p.Marshal(), s.clientAddr); err != nil {
		s.log.Event(fmt.Sprintf(writeErrSrvFmt, err), s.state())
		return
	}
	s.log.Event(fmt.Sprintf(sendOkSrvFmt, t, seq), s.state())
}

func (s *Server) state() string {
	return fmt.Sprintf(stateFmt, s.expected)
}
