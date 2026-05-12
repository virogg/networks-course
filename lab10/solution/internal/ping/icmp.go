package ping

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	cs := Checksum(pkt)
	binary.BigEndian.PutUint16(pkt[2:4], cs)
	return pkt
}

type ICMPMessage struct {
	Type     uint8
	Code     uint8
	ID       uint16
	Seq      uint16
	Payload  []byte
	Embedded *ICMPMessage // unreachable/time-exceeded: original echo header
}

var ErrBadChecksum = errors.New("icmp: bad checksum")

func ParseMessage(b []byte) (*ICMPMessage, error) {
	if len(b) < 8 {
		return nil, fmt.Errorf("icmp: short packet (%d bytes)", len(b))
	}
	cs := Checksum(b)
	if cs != 0 {
		return nil, ErrBadChecksum
	}
	m := &ICMPMessage{
		Type:    b[0],
		Code:    b[1],
		Payload: b[8:],
	}
	switch m.Type {
	case TypeEchoReply, TypeEchoRequest:
		m.ID = binary.BigEndian.Uint16(b[4:6])
		m.Seq = binary.BigEndian.Uint16(b[6:8])
	case TypeDestinationUnreachable, TypeTimeExceeded:
		body := b[8:]
		if len(body) < 4 {
			return m, nil
		}
		ipStart := 4
		if len(body) < ipStart+20 {
			return m, nil
		}
		ihl := int(body[ipStart]&0x0f) * 4
		if ihl < 20 || len(body) < ipStart+ihl+8 {
			return m, nil
		}
		orig := body[ipStart+ihl:]
		m.Embedded = &ICMPMessage{
			Type: orig[0],
			Code: orig[1],
			ID:   binary.BigEndian.Uint16(orig[4:6]),
			Seq:  binary.BigEndian.Uint16(orig[6:8]),
		}
	}
	return m, nil
}
