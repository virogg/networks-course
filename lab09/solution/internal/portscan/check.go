package portscan

import (
	"net"
	"strconv"
	"time"
)

type checker struct {
	tcp func(ip string, port int) bool
	udp func(ip string, port int) bool
}

func newChecker(mode Mode, timeout time.Duration) checker {
	if mode == ModeRemote {
		return checker{
			tcp: func(ip string, port int) bool {
				return checkConnect(ip, port, timeout)
			},
			udp: func(string, int) bool {
				return false
			},
		}
	}
	return checker{tcp: checkBindTCP, udp: checkBindUDP}
}

func checkBindTCP(ip string, port int) bool {
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	l.Close() //nolint:errcheck
	return true
}

func checkBindUDP(ip string, port int) bool {
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	c, err := net.ListenPacket("udp", addr)
	if err != nil {
		return false
	}
	c.Close() //nolint:errcheck
	return true
}

func checkConnect(ip string, port int, timeout time.Duration) bool {
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	c, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	c.Close() //nolint:errcheck
	return true
}
