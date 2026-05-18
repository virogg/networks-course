package traceroute

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	TypeEchoReply              = 0
	TypeDestinationUnreachable = 3
	TypeEchoRequest            = 8
	TypeTimeExceeded           = 11
)

func Checksum(b []byte) uint16 {
	var sum uint32
	n := len(b)
	for i := 0; i+1 < n; i += 2 {
		sum += uint32(b[i])<<8 | uint32(b[i+1])
	}
	if n%2 == 1 {
		sum += uint32(b[n-1]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}

func MarshalEcho(id, seq uint16, payload []byte) []byte {
	pkt := make([]byte, 8+len(payload))
	pkt[0] = TypeEchoRequest
	pkt[1] = 0
	binary.BigEndian.PutUint16(pkt[4:6], id)
	binary.BigEndian.PutUint16(pkt[6:8], seq)
	copy(pkt[8:], payload)
	binary.BigEndian.PutUint16(pkt[2:4], Checksum(pkt))
	return pkt
}

type ICMPMessage struct {
	Type     uint8
	Code     uint8
	ID       uint16
	Seq      uint16
	Payload  []byte
	Embedded *ICMPMessage
}

var ErrBadChecksum = errors.New("icmp: bad checksum")

func ParseMessage(b []byte) (*ICMPMessage, error) {
	if len(b) < 8 {
		return nil, fmt.Errorf("icmp: short packet (%d bytes)", len(b))
	}
	m := &ICMPMessage{Type: b[0], Code: b[1], Payload: b[8:]}
	switch m.Type {
	case TypeEchoReply, TypeEchoRequest:
		if Checksum(b) != 0 {
			return nil, ErrBadChecksum
		}
		m.ID = binary.BigEndian.Uint16(b[4:6])
		m.Seq = binary.BigEndian.Uint16(b[6:8])
	case TypeDestinationUnreachable, TypeTimeExceeded:
		orig := b[8:]
		if len(orig) < 20 {
			return m, nil
		}
		ihl := int(orig[0]&0x0f) * 4
		if ihl < 20 || len(orig) < ihl+8 {
			return m, nil
		}
		echo := orig[ihl:]
		m.Embedded = &ICMPMessage{
			Type: echo[0],
			Code: echo[1],
			ID:   binary.BigEndian.Uint16(echo[4:6]),
			Seq:  binary.BigEndian.Uint16(echo[6:8]),
		}
	}
	return m, nil
}
