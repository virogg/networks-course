package echo6

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Server struct {
	out io.Writer
}

func NewServer() *Server {
	return &Server{out: os.Stdout}
}

func (s *Server) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp6", addr)
	if err != nil {
		return fmt.Errorf("listen tcp6 %q: %w", addr, err)
	}
	defer ln.Close()
	fmt.Fprintf(s.out, "echo6 server listening on %s (IPv6)\n", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	fmt.Fprintf(s.out, "connection from %s\n", conn.RemoteAddr())

	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		msg := sc.Text()
		reply := strings.ToUpper(msg)
		fmt.Fprintf(s.out, "  %q -> %q\n", msg, reply)
		if _, err := fmt.Fprintln(conn, reply); err != nil {
			return
		}
	}
}
