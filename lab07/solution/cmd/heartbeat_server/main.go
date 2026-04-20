package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/networks-course/lab07/solution/internal/heartbeat/server"
)

func main() {
	port := flag.Int("port", 8083, "UDP port for heartbeat server")
	dead := flag.Duration("dead", 5*time.Second, "mark client down after no packets for this duration")
	check := flag.Duration("check", 500*time.Millisecond, "liveness check interval")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := server.NewServer(
		server.WithPort(*port),
		server.WithDeadAfter(*dead),
		server.WithCheckEvery(*check),
	)
	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatal(err)
	}
}
