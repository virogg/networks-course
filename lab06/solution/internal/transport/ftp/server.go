package ftp

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/virogg/networks-course/lab06/solution/internal/transport/ftp/session"
	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
)

type Server struct {
	RootDir string
	User    string
	Pass    string
}

func NewServer(rootDir, user, pass string) *Server {
	rootDir, _ = filepath.Abs(rootDir)
	os.MkdirAll(rootDir, 0755)
	return &Server{RootDir: rootDir, User: user, Pass: pass}
}

func (s *Server) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Printf("FTP server on %s, root=%s, user=%s", addr, s.RootDir, s.User)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	log.Printf("[%s] connected", remote)
	defer log.Printf("[%s] disconnected", remote)

	sess := session.New(s.RootDir, conn)
	sess.Send(ftp.StatusServiceReady, "FTP server ready")

	for {
		line, err := sess.Reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToUpper(parts[0])
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}
		if cmd == "PASS" {
			log.Printf("[%s] PASS ****", remote)
		} else {
			log.Printf("[%s] %s", remote, line)
		}

		switch cmd {
		case "USER":
			sess.User = arg
			sess.Send(ftp.StatusPasswordRequired, "Password required")
		case "PASS":
			if sess.User == s.User && arg == s.Pass {
				sess.Authed = true
				sess.Send(ftp.StatusLoginSuccessful, "Login successful")
			} else {
				sess.Send(ftp.StatusNotLoggedIn, "Login incorrect")
			}
		case "QUIT":
			sess.Send(ftp.StatusServiceClosing, "Bye")
			return
		default:
			if !sess.Authed {
				sess.Send(ftp.StatusNotLoggedIn, "Not logged in")
				continue
			}
			switch cmd {
			case "PWD":
				sess.Send(ftp.StatusCurrentDirectory, fmt.Sprintf("\"%s\"", sess.Cwd))
			case "CWD":
				sess.HandleCWD(arg)
			case "PORT":
				sess.HandlePORT(arg)
			case "NLST":
				sess.HandleNLST()
			case "RETR":
				sess.HandleRETR(arg)
			case "STOR":
				sess.HandleSTOR(arg)
			case "DELE":
				sess.HandleDELE(arg)
			case "TYPE":
				sess.Send(ftp.StatusOK, "Type set")
			case "SYST":
				sess.Send(ftp.StatusSystemType, "UNIX Type: L8")
			default:
				sess.Send(ftp.StatusNotImplemented, "Command not implemented")
			}
		}
	}
}
