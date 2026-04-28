package checksum

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompute_EvenLength(t *testing.T) {
	data := []byte{0x00, 0x01, 0xF2, 0x03, 0xF4, 0xF5, 0xF6, 0xF7}
	sum := Compute(data)
	assert.True(t, Verify(data, sum))
}

func TestCompute_OddLength_Padding(t *testing.T) {
	data := []byte("hello, world!")
	sum := Compute(data)
	assert.True(t, Verify(data, sum))
}

func TestCompute_BitFlipInData_Detected(t *testing.T) {
	data := []byte("checksum-test-payload")
	sum := Compute(data)
	corrupted := bytes.Clone(data)
	corrupted[5] ^= 0x01
	assert.False(t, Verify(corrupted, sum))
}

func TestCompute_BitFlipInChecksum_Detected(t *testing.T) {
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	sum := Compute(data)
	bad := sum ^ 0x0001
	assert.False(t, Verify(data, bad))
}

// RFC 1071:
// > To check a checksum, the 1's complement sum is computed over
// > the same set of octets, including the checksum field.
// > If the result is all 1 bits (-0 in 1's complement arithmetic), the check succeeds.
func TestCompute_RFCInvariant(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	sum := Compute(data)
	combined := append(bytes.Clone(data), byte(sum>>8), byte(sum))
	require.Equal(t, uint16(0), Compute(combined))
}

func TestCompute_EmptyInput(t *testing.T) {
	require.Equal(t, uint16(0xFFFF), Compute(nil))
}

func TestCompute_KnownVectors(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want uint16
	}{
		{"single zero byte", []byte{0x00}, 0xFFFF},
		{"two bytes 0x0001", []byte{0x00, 0x01}, 0xFFFE},
		{"all 0xFF (4 bytes)", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0x0000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, Compute(tc.data))
			assert.True(t, Verify(tc.data, tc.want))
		})
	}
}
