package snw

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrame_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		fr   Frame
	}{
		{"DATA seq=0 with payload + EOF", Frame{
			Type:    FrameData,
			Seq:     0,
			Flags:   FlagEOF,
			Payload: []byte("the quick brown fox jumps over the lazy dog"),
		}},
		{"DATA seq=1 mid-stream", Frame{
			Type:    FrameData,
			Seq:     1,
			Flags:   0,
			Payload: []byte{0x00, 0xFF, 0x10, 0x20, 0x80, 0x7F},
		}},
		{"ACK seq=0", Frame{Type: FrameAck, Seq: 0}},
		{"ACK seq=1", Frame{Type: FrameAck, Seq: 1}},
		{"HELLO", Frame{Type: FrameHello}},
		{"DATA empty payload + EOF", Frame{Type: FrameData, Seq: 1, Flags: FlagEOF}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.fr.Encode()
			got, err := Decode(raw)
			require.NoError(t, err)
			assert.Equal(t, tc.fr.Type, got.Type)
			assert.Equal(t, tc.fr.Seq, got.Seq)
			assert.Equal(t, tc.fr.Flags, got.Flags)
			assert.Equal(t, tc.fr.Payload, got.Payload)
			assert.Equal(t, tc.fr.HasEOF(), got.HasEOF())
		})
	}
}

func TestFrame_Decode_BadChecksum_PayloadFlip(t *testing.T) {
	fr := Frame{Type: FrameData, Seq: 0, Payload: []byte("payload to corrupt")}
	raw := fr.Encode()
	raw[HeaderSize+3] ^= 0x01
	_, err := Decode(raw)
	require.ErrorIs(t, err, ErrBadChecksum)
}

func TestFrame_Decode_BadChecksum_HeaderFlip(t *testing.T) {
	fr := Frame{Type: FrameData, Seq: 1, Flags: FlagEOF, Payload: []byte("x")}
	raw := fr.Encode()
	raw[6] ^= 0x01
	_, err := Decode(raw)
	require.ErrorIs(t, err, ErrBadChecksum)
}

func TestFrame_Decode_BadChecksum_ChecksumFlip(t *testing.T) {
	fr := Frame{Type: FrameAck, Seq: 0}
	raw := fr.Encode()
	raw[4] ^= 0x01
	_, err := Decode(raw)
	require.ErrorIs(t, err, ErrBadChecksum)
}

func TestFrame_Decode_ShortFrame(t *testing.T) {
	_, err := Decode([]byte{0x00, 0x01, 0x02})
	require.ErrorIs(t, err, ErrShortFrame)
}

func TestFrame_Decode_LengthMismatch(t *testing.T) {
	fr := Frame{Type: FrameData, Seq: 0, Payload: []byte("hello")}
	raw := fr.Encode()
	binary.BigEndian.PutUint16(raw[2:4], uint16(len(fr.Payload)+1))
	_, err := Decode(raw)
	require.ErrorIs(t, err, ErrLengthMismatch)
}

func TestFrameType_String(t *testing.T) {
	assert.Equal(t, "DATA", FrameData.String())
	assert.Equal(t, "ACK", FrameAck.String())
	assert.Equal(t, "HELLO", FrameHello.String())
	assert.Equal(t, "UNK", FrameType(99).String())
}
