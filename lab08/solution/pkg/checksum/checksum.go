// Internet 16-bit checksum (RFC 1071)
package checksum

func Compute(data []byte) uint16 {
	var sum uint32
	n := len(data)
	for i := 0; i+1 < n; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
		for sum>>16 != 0 {
			sum = (sum & 0xFFFF) + (sum >> 16)
		}
	}
	if n%2 == 1 {
		sum += uint32(data[n-1]) << 8
		for sum>>16 != 0 {
			sum = (sum & 0xFFFF) + (sum >> 16)
		}
	}
	return ^uint16(sum)
}

func Verify(data []byte, sum uint16) bool {
	return Compute(data) == sum
}
