package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/virogg/networks-course/lab11/solution/internal/echo6"
)

func main() {
	addr := flag.String("addr", "[::1]:9006", "IPv6 listen address")
	flag.Parse()

	if err := echo6.NewServer().ListenAndServe(*addr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
