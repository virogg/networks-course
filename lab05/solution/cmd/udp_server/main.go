package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"
)

func subnetBroadcast(ip net.IP) net.IP {
	ifaces, err := net.Interfaces()
	if err != nil {
		return net.IPv4bcast
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil || !ip4.Equal(ip.To4()) {
				continue
			}
			mask := ipNet.Mask
			if len(mask) == 16 {
				mask = mask[12:]
			}
			bcast := make(net.IP, 4)
			for i := range bcast {
				bcast[i] = ip4[i] | ^mask[i]
			}
			return bcast
		}
	}
	return net.IPv4bcast
}

func main() {
	port := flag.Int("port", 8080, "broadcast port")
	flag.Parse()

	tmp, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalf("routing lookup: %v", err)
	}
	localIP := tmp.LocalAddr().(*net.UDPAddr).IP
	tmp.Close() //nolint:checkerr

	bcastIP := subnetBroadcast(localIP)

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: localIP, Port: 0})
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer conn.Close() //nolint:checkerr

	rawConn, err := conn.SyscallConn()
	if err != nil {
		log.Fatalf("syscall conn: %v", err)
	}
	var setErr error
	rawConn.Control(func(fd uintptr) { //nolint:checkerr
		setErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	})
	if setErr != nil {
		log.Fatalf("set broadcast: %v", setErr)
	}

	dst := &net.UDPAddr{IP: bcastIP, Port: *port}
	fmt.Printf("broadcasting to %s every 1s\n", dst)

	for {
		msg := time.Now().Format("2006-01-02 15:04:05")
		if _, err := conn.WriteToUDP([]byte(msg), dst); err != nil {
			log.Printf("write: %v", err)
		}
		fmt.Printf("sent: %s\n", msg)
		time.Sleep(time.Second)
	}
}
