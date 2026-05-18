package echo6

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

const CRLF = "\r\n"

func Echo(addr, message string) (reply, remote string, err error) {
	conn, err := net.Dial("tcp6", addr)
	if err != nil {
		return "", "", fmt.Errorf("dial tcp6 %q: %w", addr, err)
	}
	defer conn.Close()
	remote = conn.RemoteAddr().String()

	if _, err := fmt.Fprintln(conn, message); err != nil {
		return "", remote, fmt.Errorf("send: %w", err)
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", remote, fmt.Errorf("receive: %w", err)
	}
	return strings.TrimRight(line, CRLF), remote, nil
}
