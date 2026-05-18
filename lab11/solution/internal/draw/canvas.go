package draw

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const strokeWidth = 3

type Canvas struct {
	mu  sync.Mutex
	img *ebiten.Image
}

func NewCanvas() *Canvas {
	img := ebiten.NewImage(Width, Height)
	img.Fill(color.White)
	return &Canvas{img: img}
}

func (c *Canvas) AddSegment(s Segment) {
	c.mu.Lock()
	defer c.mu.Unlock()
	clr := color.RGBA{R: s.R, G: s.G, B: s.B, A: 255}
	vector.StrokeLine(c.img, s.X1, s.Y1, s.X2, s.Y2, strokeWidth, clr, true)
}

func (c *Canvas) Render(screen *ebiten.Image) {
	c.mu.Lock()
	defer c.mu.Unlock()
	screen.DrawImage(c.img, nil)
}
