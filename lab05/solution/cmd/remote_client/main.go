package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	host := flag.String("host", "localhost", "server host")
	port := flag.String("port", "8080", "server port")
	cmd := flag.String("cmd", "echo hello", "command to run on server")
	flag.Parse()

	conn, err := net.Dial("tcp", *host+":"+*port)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close() //nolint:checkerr

	fmt.Fprint(conn, *cmd)

	if tc, ok := conn.(*net.TCPConn); ok {
		tc.CloseWrite() //nolint:checkerr
	}

	out, err := io.ReadAll(conn)
	if err != nil {
		log.Fatalf("read: %v", err)
	}
	fmt.Print(string(out))
}
