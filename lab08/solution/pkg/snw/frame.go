package snw

import (
	"encoding/binary"
	"errors"

	"github.com/virogg/networks-course/lab08/solution/pkg/checksum"
)

type FrameType uint8

const (
	FrameData  FrameType = 0
	FrameAck   FrameType = 1
	FrameHello FrameType = 2
)

func (t FrameType) String() string {
	switch t {
	case FrameData:
		return "DATA"
	case FrameAck:
		return "ACK"
	case FrameHello:
		return "HELLO"
	default:
		return "UNK"
	}
}

const (
	FlagEOF uint16 = 1 << 0
)

const HeaderSize = 8

const MaxPayload = 1500 - HeaderSize - 28 // 28 = IPv4(20) + UDP(8)

// Frame
//
// | 1 byte | 1 byte | 2 bytes BE | 2 bytes BE | 2 bytes BE | N bytes |
// |  type  |  seq   |   length   |  checksum  |   flags    | payload |
type Frame struct {
	Type    FrameType
	Seq     uint8
	Flags   uint16
	Payload []byte
}

var (
	ErrShortFrame     = errors.New("frame too short")
	ErrLengthMismatch = errors.New("length field mismatch")
	ErrBadChecksum    = errors.New("checksum mismatch")
)

func (f Frame) Encode() []byte {
	buf := make([]byte, HeaderSize+len(f.Payload))
	buf[0] = byte(f.Type)
	buf[1] = f.Seq
	binary.BigEndian.PutUint16(buf[2:4], uint16(len(f.Payload)))
	binary.BigEndian.PutUint16(buf[4:6], 0) // for checksum
	binary.BigEndian.PutUint16(buf[6:8], f.Flags)
	copy(buf[HeaderSize:], f.Payload)
	sum := checksum.Compute(buf)
	binary.BigEndian.PutUint16(buf[4:6], sum)
	return buf
}

func Decode(data []byte) (Frame, error) {
	if len(data) < HeaderSize {
		return Frame{}, ErrShortFrame
	}
	length := binary.BigEndian.Uint16(data[2:4])
	if int(length) != len(data)-HeaderSize {
		return Frame{}, ErrLengthMismatch
	}
	sum := binary.BigEndian.Uint16(data[4:6])
	tmp := make([]byte, len(data))
	copy(tmp, data)
	binary.BigEndian.PutUint16(tmp[4:6], 0)
	if !checksum.Verify(tmp, sum) {
		return Frame{}, ErrBadChecksum
	}
	flags := binary.BigEndian.Uint16(data[6:8])
	var payload []byte
	if length > 0 {
		payload = make([]byte, length)
		copy(payload, data[HeaderSize:])
	}
	return Frame{
		Type:    FrameType(data[0]),
		Seq:     data[1],
		Flags:   flags,
		Payload: payload,
	}, nil
}

func (f Frame) HasEOF() bool {
	return f.Flags&FlagEOF != 0
}
