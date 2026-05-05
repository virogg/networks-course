package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/networks-course/lab09/solution/internal/portscan"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "ip address to scan")
	from := flag.Int("from", 8000, "first port (inclusive)")
	to := flag.Int("to", 8100, "last port (inclusive)")
	protoStr := flag.String("proto", "tcp", "tcp|udp|both")
	modeStr := flag.String("mode", "auto", "auto|local|remote (auto = local for own IP, remote TCP-connect otherwise)")
	workers := flag.Int("workers", 256, "concurrent workers")
	timeout := flag.Duration("timeout", 1500*time.Millisecond, "TCP-connect timeout (remote mode)")
	flag.Parse()

	proto, err := portscan.ParseProto(*protoStr)
	if err != nil {
		log.Fatal(err)
	}
	mode, err := portscan.ParseMode(*modeStr)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	results, resolvedMode, err := portscan.Scan(ctx,
		portscan.WithIP(*ip),
		portscan.WithFrom(*from),
		portscan.WithTo(*to),
		portscan.WithProto(proto),
		portscan.WithMode(mode),
		portscan.WithWorkers(*workers),
		portscan.WithTimeout(*timeout),
	)
	if err != nil {
		log.Fatal(err)
	}

	label := "free"
	if resolvedMode == portscan.ModeRemote {
		label = "open"
	}
	total := *to - *from + 1
	_, _ = fmt.Fprintf(os.Stdout, "scanned %s ports %d-%d (proto=%s mode=%s): %d/%d %s\n",
		*ip, *from, *to, proto, resolvedMode, len(results), total, label)

	for _, r := range results {
		switch proto {
		case portscan.ProtoBoth:
			_, _ = fmt.Fprintf(os.Stdout, "%-6d tcp=%v udp=%v\n", r.Port, r.TCP, r.UDP)
		default:
			_, _ = fmt.Fprintln(os.Stdout, r.Port)
		}
	}
}
