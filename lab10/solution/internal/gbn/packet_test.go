package gbn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPacketRoundTrip_Data(t *testing.T) {
	payload := []byte("the quick brown fox")
	wire := Packet{Type: TypeData, Seq: 42, Payload: payload}.Marshal()
	out, err := Unmarshal(wire)
	require.NoError(t, err)
	assert.Equal(t, TypeData, out.Type)
	assert.Equal(t, uint32(42), out.Seq)
	assert.Equal(t, payload, out.Payload)
}

func TestPacketRoundTrip_AckEmpty(t *testing.T) {
	wire := Packet{Type: TypeAck, Seq: 7}.Marshal()
	out, err := Unmarshal(wire)
	require.NoError(t, err)
	assert.Equal(t, TypeAck, out.Type)
	assert.Equal(t, uint32(7), out.Seq)
	assert.Empty(t, out.Payload)
}

func TestPacketRoundTrip_FIN(t *testing.T) {
	wire := Packet{Type: TypeFin, Seq: 99}.Marshal()
	out, err := Unmarshal(wire)
	require.NoError(t, err)
	assert.Equal(t, TypeFin, out.Type)
	assert.Equal(t, uint32(99), out.Seq)
}

func TestUnmarshal_BadCRC(t *testing.T) {
	wire := Packet{Type: TypeData, Seq: 1, Payload: []byte("abc")}.Marshal()
	wire[len(wire)-1] ^= 0xff
	_, err := Unmarshal(wire)
	require.ErrorIs(t, err, ErrCRC)
}

func TestUnmarshal_Short(t *testing.T) {
	_, err := Unmarshal([]byte{0x01, 0x00})
	require.ErrorIs(t, err, ErrShort)
}

func TestUnmarshal_LengthMismatch(t *testing.T) {
	wire := Packet{Type: TypeData, Seq: 1, Payload: []byte("abc")}.Marshal()
	truncated := append(wire[:HeaderLen+1], wire[len(wire)-TrailerLen:]...)
	_, err := Unmarshal(truncated)
	require.Error(t, err)
}
