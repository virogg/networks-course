package draw

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
)

const sendQueue = 4096

type rgb struct{ R, G, B uint8 }

type clientGame struct {
	canvas  *Canvas
	out     chan<- Segment
	color   rgb
	prevX   float32
	prevY   float32
	drawing bool
	closed  atomic.Bool
}

func (g *clientGame) disconnect(reason string) {
	if g.closed.CompareAndSwap(false, true) {
		log.Printf("draw server disconnected (%s) — closing", reason)
	}
}

func (g *clientGame) Update() error {
	if g.closed.Load() {
		return ebiten.Termination
	}

	x, y := ebiten.CursorPosition()
	fx, fy := float32(x), float32(y)

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.drawing {
			seg := Segment{
				X1: g.prevX, Y1: g.prevY, X2: fx, Y2: fy,
				R: g.color.R, G: g.color.G, B: g.color.B,
			}
			g.canvas.AddSegment(seg)
			select {
			case g.out <- seg: // stream to server
			default: // queue full
			}
		}
		g.prevX, g.prevY, g.drawing = fx, fy, true
	} else {
		g.drawing = false
	}
	return nil
}

func (g *clientGame) Draw(screen *ebiten.Image) {
	g.canvas.Render(screen)
}
func (g *clientGame) Layout(_, _ int) (int, int) {
	return Width, Height
}

func RunClient(addr string, r, g, b uint8) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to draw server %q: %w", addr, err)
	}
	defer conn.Close()
	log.Printf("connected to draw server %s", conn.RemoteAddr())

	out := make(chan Segment, sendQueue)
	game := &clientGame{
		canvas: NewCanvas(),
		out:    out,
		color:  rgb{R: r, G: g, B: b},
	}

	go func() {
		for seg := range out {
			if err := WriteSegment(conn, seg); err != nil {
				game.disconnect(err.Error())
				return
			}
		}
	}()

	go func() {
		if _, err := conn.Read(make([]byte, 1)); err != nil {
			game.disconnect(err.Error())
		}
	}()

	ebiten.SetWindowSize(Width, Height)
	ebiten.SetWindowTitle("Remote drawing — client")
	return ebiten.RunGame(game)
}
