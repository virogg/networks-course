package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/virogg/networks-course/lab11/solution/internal/draw"
)

func main() {
	addr := flag.String("addr", ":9007", "TCP listen address")
	flag.Parse()

	if err := draw.RunServer(*addr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
