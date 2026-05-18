package draw

import (
	"encoding/binary"
	"io"
)

const (
	Width  = 800
	Height = 600
)

type Segment struct {
	X1, Y1, X2, Y2 float32
	R, G, B        uint8
}

// encodes segment to TCP stream
func WriteSegment(w io.Writer, s Segment) error {
	return binary.Write(w, binary.BigEndian, &s)
}

// decodes segment from TCP stream
func ReadSegment(r io.Reader) (Segment, error) {
	var s Segment
	err := binary.Read(r, binary.BigEndian, &s)
	return s, err
}
