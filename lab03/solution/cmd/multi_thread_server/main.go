package main

import (
	"fmt"
	"log"
	"os"

	"github.com/virogg/networks-course/solution/pkg/server"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: <server.exe> server_port")
		os.Exit(1)
	}

	port := os.Args[1]
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

		go server.HandleConnection(conn)
	}
}
