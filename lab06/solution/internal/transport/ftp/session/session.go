package session

import (
	"bufio"
	"fmt"
	"net"
	"path/filepath"
)

type Session struct {
	RootDir  string
	Conn     net.Conn
	Reader   *bufio.Reader
	Cwd      string
	DataAddr string
	Authed   bool
	User     string
}

func New(rootDir string, conn net.Conn) *Session {
	return &Session{
		RootDir: rootDir,
		Conn:    conn,
		Reader:  bufio.NewReader(conn),
		Cwd:     "/",
	}
}

func (s *Session) Send(code int, msg string) {
	fmt.Fprintf(s.Conn, "%d %s\r\n", code, msg)
}

func (s *Session) VirtualPath(p string) string {
	if !filepath.IsAbs(p) {
		p = filepath.Join(s.Cwd, p)
	}
	return filepath.Clean(p)
}

func (s *Session) RealPath(p string) string {
	return filepath.Join(s.RootDir, s.VirtualPath(p))
}

func (s *Session) OpenDataConn() (net.Conn, error) {
	if s.DataAddr == "" {
		return nil, fmt.Errorf("no PORT specified")
	}
	return net.Dial("tcp", s.DataAddr)
}
