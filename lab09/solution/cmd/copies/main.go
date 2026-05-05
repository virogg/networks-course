package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/networks-course/lab09/solution/internal/copies"
)

func main() {
	port := flag.Int("port", 9999, "broadcast port")
	interval := flag.Duration("interval", 2*time.Second, "broadcast interval")
	dead := flag.Int("dead-multiplier", 3, "drop peer after dead-multiplier * interval without messages")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := copies.NewApp(
		copies.WithPort(*port),
		copies.WithInterval(*interval),
		copies.WithDeadMultiplier(*dead),
	)

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
