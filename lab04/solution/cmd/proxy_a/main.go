package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/virogg/networks-course/solution/pkg/logger"
	"github.com/virogg/networks-course/solution/pkg/server"
)

func main() {
	port := flag.Int("port", 8888, "listen port")
	lvl := flag.String("level", "local", "logging level")
	flag.Parse()

	logger, err := logger.New(*lvl)
	if err != nil {
		log.Fatalf("logger: %v", err)
	}

	http.HandleFunc("/", server.Handler(server.Config{Logger: logger}))

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("proxy_a listening on %s — GET/POST, no cache", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
