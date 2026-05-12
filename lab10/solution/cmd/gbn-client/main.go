package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/networks-course/lab10/solution/internal/gbn"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9000", "server UDP address")
	file := flag.String("file", "", "file to send")
	chunk := flag.Int("chunk", 1024, "chunk size in bytes")
	window := flag.Uint("window", 4, "window size (number of unacked packets)")
	timeout := flag.Duration("timeout", 500*time.Millisecond, "retransmit timeout")
	loss := flag.Float64("loss-rate", 0, "simulated DATA loss rate [0,1)")
	flag.Parse()

	if *file == "" {
		fmt.Fprintln(os.Stderr, "--file is required")
		os.Exit(2)
	}

	c, err := gbn.NewClient(
		gbn.WithClientRemoteAddr(*addr),
		gbn.WithClientFile(*file),
		gbn.WithClientChunkSize(*chunk),
		gbn.WithClientWindowSize(uint32(*window)),
		gbn.WithClientTimeout(*timeout),
		gbn.WithClientLossRate(*loss),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer c.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := c.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
