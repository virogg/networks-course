package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/virogg/networks-course/solution/pkg/blacklist"
	"github.com/virogg/networks-course/solution/pkg/cache"
	"github.com/virogg/networks-course/solution/pkg/logger"
	"github.com/virogg/networks-course/solution/pkg/server"
)

func main() {
	port := flag.Int("port", 8890, "listen port")
	lvl := flag.String("level", "local", "logging level")
	cacheDir := flag.String("cache", "cache_c", "cache directory")
	blFile := flag.String("blacklist", "blacklist.json", "blacklist config file")
	flag.Parse()

	logger, err := logger.New(*lvl)
	if err != nil {
		log.Fatalf("logger: %v", err)
	}

	c, err := cache.New(*cacheDir)
	if err != nil {
		log.Fatalf("cache: %v", err)
	}

	bl := blacklist.Load(*blFile)
	log.Printf("loaded %d blacklist entries from %s", len(bl), *blFile)

	http.HandleFunc("/", server.Handler(server.Config{Logger: logger, Cache: c, Blacklist: bl}))

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("proxy_c listening on %s — GET/POST, cache, blacklist", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
