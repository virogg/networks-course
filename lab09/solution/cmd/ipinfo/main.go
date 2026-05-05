package main

import (
	"flag"
	"log"
	"os"

	"github.com/virogg/networks-course/lab09/solution/internal/ipinfo"
)

func main() {
	all := flag.Bool("all", false, "include IPv6 and loopback interfaces")
	flag.Parse()

	if err := ipinfo.Print(os.Stdout, *all); err != nil {
		log.Fatal(err)
	}
}
