package gbn

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

type Type uint8

const (
	TypeData   Type = 0x01
	TypeAck    Type = 0x02
	TypeFin    Type = 0x03
	TypeFinAck Type = 0x04
)

func (t Type) String() string {
	switch t {
	case TypeData:
		return "DATA"
	case TypeAck:
		return "ACK"
	case TypeFin:
		return "FIN"
	case TypeFinAck:
		return "FIN-ACK"
	}
	return fmt.Sprintf("UNKNOWN(0x%x)", uint8(t))
}

const (
	HeaderLen  = 1 + 4 + 4
	TrailerLen = 4
)

type Packet struct {
	Type    Type
	Seq     uint32
	Payload []byte
}

var (
	ErrShort = errors.New("gbn: short packet")
	ErrCRC   = errors.New("gbn: bad CRC")
)

func (p Packet) Marshal() []byte {
	buf := make([]byte, HeaderLen+len(p.Payload)+TrailerLen)
	buf[0] = byte(p.Type)
	binary.BigEndian.PutUint32(buf[1:5], p.Seq)
	binary.BigEndian.PutUint32(buf[5:9], uint32(len(p.Payload)))
	copy(buf[HeaderLen:], p.Payload)
	body := buf[:len(buf)-TrailerLen]
	binary.BigEndian.PutUint32(buf[len(buf)-TrailerLen:], crc32.ChecksumIEEE(body))
	return buf
}

func Unmarshal(b []byte) (Packet, error) {
	if len(b) < HeaderLen+TrailerLen {
		return Packet{}, ErrShort
	}
	plen := binary.BigEndian.Uint32(b[5:9])
	if int(plen) != len(b)-HeaderLen-TrailerLen {
		return Packet{}, fmt.Errorf("gbn: length mismatch (declared=%d actual=%d)", plen, len(b)-HeaderLen-TrailerLen)
	}
	body := b[:len(b)-TrailerLen]
	got := binary.BigEndian.Uint32(b[len(b)-TrailerLen:])
	if crc32.ChecksumIEEE(body) != got {
		return Packet{}, ErrCRC
	}
	payload := make([]byte, plen)
	copy(payload, b[HeaderLen:HeaderLen+plen])
	return Packet{
		Type:    Type(b[0]),
		Seq:     binary.BigEndian.Uint32(b[1:5]),
		Payload: payload,
	}, nil
}
