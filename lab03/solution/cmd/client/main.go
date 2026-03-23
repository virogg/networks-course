package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: <client.exe> server_host server_port filename")
	}
	host := os.Args[1]
	port := os.Args[2]
	filename := os.Args[3]

	addr := net.JoinHostPort(host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("failed to connect to %s: %v\n", addr, err)
	}
	defer conn.Close()

	if !strings.HasPrefix(filename, "/") {
		filename = "/" + filename
	}

	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\n\r\n", filename, host)
	_, err = conn.Write([]byte(req))
	if err != nil {
		log.Fatalf("failed to send request: %v", err)
	}

	log.Printf("Http request: %s", req)

	log.Println("Http response:")
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			fmt.Print(line)
		}
		if err != nil {
			break
		}
	}
}
