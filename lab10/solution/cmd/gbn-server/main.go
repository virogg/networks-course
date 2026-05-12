package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/virogg/networks-course/lab10/solution/internal/gbn"
)

func main() {
	addr := flag.String("addr", ":9000", "UDP listen address")
	out := flag.String("out", "received.bin", "output file path")
	loss := flag.Float64("loss-rate", 0, "simulated ACK loss rate [0,1)")
	flag.Parse()

	srv, err := gbn.NewServer(
		gbn.WithServerListenAddr(*addr),
		gbn.WithServerOutPath(*out),
		gbn.WithServerLossRate(*loss),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer srv.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
