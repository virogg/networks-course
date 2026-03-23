package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/virogg/networks-course/solution/pkg/parse"
	"github.com/virogg/networks-course/solution/pkg/server"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: <server.exe> server_port concurrency_level")
		os.Exit(1)
	}

	port := os.Args[1]
	concurrencyLevel, err := parse.PositiveInt(os.Args[2])
	if err != nil {
		log.Fatalf("failed to parse concurrency_level: %v", err)
	}
	sema := make(chan struct{}, concurrencyLevel)

	lis, cleanup, err := server.Listen(port)
	defer cleanup()
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		sema <- struct{}{}
		go func(c net.Conn) {
			defer func() { <-sema }()
			server.HandleConnection(c)
		}(conn)
	}
}
