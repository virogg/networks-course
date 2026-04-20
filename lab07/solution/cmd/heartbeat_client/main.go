package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/networks-course/lab07/solution/internal/heartbeat/client"
)

func main() {
	host := flag.String("host", "127.0.0.1", "heartbeat server host")
	port := flag.Int("port", 8083, "heartbeat server port")
	interval := flag.Duration("interval", time.Second, "heartbeat interval")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c := client.NewClient(
		client.WithPort(*port),
		client.WithHost(*host),
		client.WithInterval(*interval),
	)
	if err := c.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
