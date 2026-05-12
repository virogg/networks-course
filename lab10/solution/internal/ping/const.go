package ping

import "fmt"

const (
	TypeEchoReply              = 0
	TypeDestinationUnreachable = 3
	TypeEchoRequest            = 8
	TypeTimeExceeded           = 11
)

func DescribeError(typ, code uint8) string {
	switch typ {
	case TypeDestinationUnreachable:
		switch code {
		case 0:
			return "Destination network unreachable"
		case 1:
			return "Destination host unreachable"
		case 2:
			return "Protocol unreachable"
		case 3:
			return "Port unreachable"
		case 4:
			return "Fragmentation needed (DF set)"
		case 5:
			return "Source route failed"
		case 6:
			return "Destination network unknown"
		case 7:
			return "Destination host unknown"
		case 9:
			return "Network administratively prohibited"
		case 10:
			return "Host administratively prohibited"
		case 13:
			return "Communication administratively prohibited"
		default:
			return fmt.Sprintf("Destination unreachable (code=%d)", code)
		}
	case TypeTimeExceeded:
		switch code {
		case 0:
			return "TTL exceeded in transit"
		case 1:
			return "Fragment reassembly time exceeded"
		default:
			return fmt.Sprintf("Time exceeded (code=%d)", code)
		}
	}
	return ""
}
