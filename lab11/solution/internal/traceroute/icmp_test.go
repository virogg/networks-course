package traceroute

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func timeExceededPacket(echoID, echoSeq uint16) []byte {
	pkt := make([]byte, 8+20+8)
	pkt[0] = TypeTimeExceeded
	ip := pkt[8:28]
	ip[0] = 0x45
	echo := pkt[28:36]
	echo[0] = TypeEchoRequest
	binary.BigEndian.PutUint16(echo[4:6], echoID)
	binary.BigEndian.PutUint16(echo[6:8], echoSeq)
	binary.BigEndian.PutUint16(pkt[2:4], Checksum(pkt))
	return pkt
}

func TestChecksumZeroOnValidPacket(t *testing.T) {
	require.Zero(t, Checksum(MarshalEcho(1, 2, []byte("hello"))))
}

func TestMarshalParseEchoRoundTrip(t *testing.T) {
	msg, err := ParseMessage(MarshalEcho(0x1234, 42, []byte("payload")))
	require.NoError(t, err)
	require.Equal(t, uint8(TypeEchoRequest), msg.Type)
	require.Equal(t, uint16(0x1234), msg.ID)
	require.Equal(t, uint16(42), msg.Seq)
}

func TestParseTimeExceededExtractsEmbeddedEcho(t *testing.T) {
	msg, err := ParseMessage(timeExceededPacket(0xABCD, 7))
	require.NoError(t, err)
	require.Equal(t, uint8(TypeTimeExceeded), msg.Type)
	require.NotNil(t, msg.Embedded)
	require.Equal(t, uint16(0xABCD), msg.Embedded.ID)
	require.Equal(t, uint16(7), msg.Embedded.Seq)
}

func TestMatchSeq(t *testing.T) {
	const id = 0x4567

	echoReply := MarshalEcho(id, 5, nil)
	echoReply[0] = TypeEchoReply
	binary.BigEndian.PutUint16(echoReply[2:4], 0)
	binary.BigEndian.PutUint16(echoReply[2:4], Checksum(echoReply))
	reply, err := ParseMessage(echoReply)
	require.NoError(t, err)
	seq, ok := matchSeq(reply, id)
	require.True(t, ok)
	require.Equal(t, uint16(5), seq)

	exceeded, err := ParseMessage(timeExceededPacket(id, 9))
	require.NoError(t, err)
	seq, ok = matchSeq(exceeded, id)
	require.True(t, ok)
	require.Equal(t, uint16(9), seq)

	_, ok = matchSeq(exceeded, id+1)
	require.False(t, ok)
}
