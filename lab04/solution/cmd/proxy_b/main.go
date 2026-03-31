package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/virogg/networks-course/solution/pkg/cache"
	"github.com/virogg/networks-course/solution/pkg/logger"
	"github.com/virogg/networks-course/solution/pkg/server"
)

func main() {
	port := flag.Int("port", 8889, "listen port")
	lvl := flag.String("level", "local", "logging level")
	cacheDir := flag.String("cache", "cache_b", "cache directory")
	flag.Parse()

	logger, err := logger.New(*lvl)
	if err != nil {
		log.Fatalf("logger: %v", err)
	}

	c, err := cache.New(*cacheDir)
	if err != nil {
		log.Fatalf("cache: %v", err)
	}

	http.HandleFunc("/", server.Handler(server.Config{Logger: logger, Cache: c}))

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("proxy_b listening on %s — GET/POST, cache in %s", addr, *cacheDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}
