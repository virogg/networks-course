package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/networks-course/lab10/solution/internal/ping"
)

func main() {
	interval := flag.Duration("interval", time.Second, "interval between echo requests")
	timeout := flag.Duration("timeout", time.Second, "per-request reply timeout")
	count := flag.Int("count", 0, "number of requests to send (0 = until interrupted)")
	size := flag.Int("size", 56, "ICMP payload size in bytes (>=8)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <host> [flags]\n", os.Args[0])
		flag.PrintDefaults()
	}
	args := os.Args[1:]
	var positional []string
	for {
		if err := flag.CommandLine.Parse(args); err != nil {
			os.Exit(2)
		}
		rest := flag.Args()
		if len(rest) == 0 {
			break
		}
		positional = append(positional, rest[0])
		args = rest[1:]
	}

	if len(positional) != 1 {
		flag.Usage()
		os.Exit(2)
	}
	host := positional[0]

	p := ping.New(
		ping.WithHost(host),
		ping.WithInterval(*interval),
		ping.WithTimeout(*timeout),
		ping.WithCount(*count),
		ping.WithSize(*size),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	start := time.Now()
	err := p.Run(ctx)
	fmt.Println()
	fmt.Println(p.Stats().Summary(host, time.Since(start)))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
