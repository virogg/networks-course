package copies

import (
	"errors"
	"fmt"
	"strings"
)

type Kind string

const (
	KindHello Kind = "HELLO"
	KindAlive Kind = "ALIVE"
	KindBye   Kind = "BYE"
)

type Message struct {
	Kind Kind
	ID   string
}

func (m Message) Encode() []byte {
	return fmt.Appendf(nil, "%s %s", m.Kind, m.ID)
}

func Decode(b []byte) (Message, error) {
	parts := strings.Fields(string(b))
	if len(parts) != 2 {
		return Message{}, errors.New("malformed message")
	}
	k := Kind(parts[0])
	switch k {
	case KindHello, KindAlive, KindBye:
	default:
		return Message{}, fmt.Errorf("unknown kind %q", parts[0])
	}
	return Message{Kind: k, ID: parts[1]}, nil
}
