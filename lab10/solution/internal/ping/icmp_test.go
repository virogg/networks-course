package ping

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksum_ZeroBuffer(t *testing.T) {
	require.Equal(t, uint16(0xffff), Checksum(make([]byte, 16)))
}

func TestChecksum_KnownVector(t *testing.T) {
	b := []byte{0x00, 0x01, 0xf2, 0x03, 0xf4, 0xf5, 0xf6, 0xf7}
	require.Equal(t, uint16(0x220d), Checksum(b))
}

func TestChecksum_OddLength(t *testing.T) {
	b := []byte{0x12, 0x34, 0x56}
	require.Equal(t, uint16(0x97cb), Checksum(b))
}

func TestMarshalEchoRoundTrip(t *testing.T) {
	payload := []byte("hello world payload")
	pkt := MarshalEcho(0x1234, 0x0007, payload)
	require.Zero(t, Checksum(pkt), "packet does not self-verify")

	msg, err := ParseMessage(pkt)
	require.NoError(t, err)
	assert.Equal(t, uint8(TypeEchoRequest), msg.Type)
	assert.Equal(t, uint8(0), msg.Code)
	assert.Equal(t, uint16(0x1234), msg.ID)
	assert.Equal(t, uint16(0x0007), msg.Seq)
	assert.Equal(t, payload, msg.Payload)
}

func TestParseMessage_BadChecksum(t *testing.T) {
	pkt := MarshalEcho(1, 2, []byte("x"))
	pkt[3] ^= 0xff
	_, err := ParseMessage(pkt)
	require.ErrorIs(t, err, ErrBadChecksum)
}

func TestDescribeError(t *testing.T) {
	tests := []struct {
		name      string
		typ, code uint8
		want      string
	}{
		{"net unreach", 3, 0, "Destination network unreachable"},
		{"host unreach", 3, 1, "Destination host unreachable"},
		{"port unreach", 3, 3, "Port unreachable"},
		{"ttl exceeded", 11, 0, "TTL exceeded in transit"},
		{"non-error type", 8, 0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, DescribeError(tc.typ, tc.code))
		})
	}
}
