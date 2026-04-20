package main

import (
	"flag"
	"log"
	"time"

	"github.com/virogg/networks-course/lab07/solution/internal/ping/client"
)

func main() {
	host := flag.String("host", "127.0.0.1", "ping server host")
	port := flag.Int("port", 8082, "ping server port")
	count := flag.Int("count", 10, "number of echo requests")
	timeout := flag.Duration("timeout", time.Second, "timeout for each reply")
	stats := flag.Bool("stats", false, "print ping-style summary (min/avg/max, loss %%)")
	flag.Parse()

	c := client.NewClient(
		client.WithHost(*host),
		client.WithPort(*port),
		client.WithCount(*count),
		client.WithTimeout(*timeout),
		client.WithStats(*stats),
	)

	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
