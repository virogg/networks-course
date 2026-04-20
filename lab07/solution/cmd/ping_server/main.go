package main

import (
	"flag"
	"log"

	"github.com/virogg/networks-course/lab07/solution/internal/ping/server"
)

func main() {
	port := flag.Int("port", 8082, "UDP port for ping server")
	flag.Parse()

	srv := server.NewServer(
		server.WithPort(*port),
	)
	log.Fatal(srv.ListenAndServe())
}
