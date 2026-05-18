package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/virogg/networks-course/lab11/solution/internal/draw"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9007", "draw server address")
	r := flag.Uint("r", 0, "stroke colour red component (0-255)")
	g := flag.Uint("g", 0, "stroke colour green component (0-255)")
	b := flag.Uint("b", 0, "stroke colour blue component (0-255)")
	flag.Parse()

	if err := draw.RunClient(*addr, uint8(*r), uint8(*g), uint8(*b)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
