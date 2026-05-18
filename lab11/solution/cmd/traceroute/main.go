package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/networks-course/lab11/solution/internal/traceroute"
)

func main() {
	queries := flag.Int("queries", 3, "number of probes per hop")
	maxHops := flag.Int("max-hops", 30, "maximum TTL")
	timeout := flag.Duration("timeout", 2*time.Second, "per-probe reply timeout")
	resolve := flag.Bool("resolve", true, "reverse-DNS resolve hop names")
	verbose := flag.Bool("verbose", false, "log every sent/received packet to stderr")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <host> [flags]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	tr := traceroute.New(
		traceroute.WithHost(flag.Arg(0)),
		traceroute.WithQueries(*queries),
		traceroute.WithMaxHops(*maxHops),
		traceroute.WithTimeout(*timeout),
		traceroute.WithResolve(*resolve),
		traceroute.WithVerbose(*verbose),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := tr.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
