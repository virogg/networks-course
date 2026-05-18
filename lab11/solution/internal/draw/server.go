package draw

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/hajimehoshi/ebiten/v2"
)

type serverGame struct {
	canvas *Canvas
}

func (g *serverGame) Update() error              { return nil }
func (g *serverGame) Draw(screen *ebiten.Image)  { g.canvas.Render(screen) }
func (g *serverGame) Layout(_, _ int) (int, int) { return Width, Height }

func RunServer(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %q: %w", addr, err)
	}
	log.Printf("draw server listening on %s", ln.Addr())

	canvas := NewCanvas()
	go acceptLoop(ln, canvas)

	ebiten.SetWindowSize(Width, Height)
	ebiten.SetWindowTitle("Remote drawing — server (replica)")
	return ebiten.RunGame(&serverGame{canvas: canvas})
}

func acceptLoop(ln net.Listener, canvas *Canvas) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			return
		}
		go handleConn(conn, canvas)
	}
}

func handleConn(conn net.Conn, canvas *Canvas) {
	defer conn.Close()
	log.Printf("client connected: %s", conn.RemoteAddr())
	for {
		seg, err := ReadSegment(conn)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("read from %s: %v", conn.RemoteAddr(), err)
			}
			log.Printf("client disconnected: %s", conn.RemoteAddr())
			return
		}
		canvas.AddSegment(seg)
	}
}
