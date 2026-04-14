package ftp

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func SendCmd(conn net.Conn, cmd string) error {
	_, err := fmt.Fprintf(conn, "%s\r\n", cmd)
	return err
}

func ReadResponse(r *bufio.Reader) (int, string, error) {
	var lines []string
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return 0, "", err
		}
		line = strings.TrimRight(line, "\r\n")
		lines = append(lines, line)
		if len(line) >= 4 && line[3] == ' ' {
			break
		}
	}
	full := strings.Join(lines, "\n")
	code, _ := strconv.Atoi(full[:3])
	return code, full, nil
}

func ParsePORTAddr(args string) (string, error) {
	parts := strings.Split(args, ",")
	if len(parts) != 6 {
		return "", fmt.Errorf("invalid PORT args: %s", args)
	}
	p1, _ := strconv.Atoi(strings.TrimSpace(parts[4]))
	p2, _ := strconv.Atoi(strings.TrimSpace(parts[5]))
	port := p1*256 + p2
	host := fmt.Sprintf("%s.%s.%s.%s", parts[0], parts[1], parts[2], parts[3])
	return fmt.Sprintf("%s:%d", host, port), nil
}

func BuildPORTArgs(addr net.Addr) string {
	tcpAddr := addr.(*net.TCPAddr)
	ip := tcpAddr.IP.To4()
	p1 := tcpAddr.Port / 256
	p2 := tcpAddr.Port % 256
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d", ip[0], ip[1], ip[2], ip[3], p1, p2)
}
