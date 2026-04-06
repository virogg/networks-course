package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 8080, "listen port")
	flag.Parse()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: *port})
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer conn.Close()

	fmt.Printf("listening for broadcasts on :%d\n", *port)

	buf := make([]byte, 1024)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("read: %v", err)
			continue
		}
		fmt.Printf("[%s] %s\n", src, string(buf[:n]))
	}
}
