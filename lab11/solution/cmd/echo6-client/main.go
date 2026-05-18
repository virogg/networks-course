package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/virogg/networks-course/lab11/solution/internal/echo6"
)

func main() {
	addr := flag.String("addr", "[::1]:9006", "IPv6 server address")
	message := flag.String("message", "hello, ipv6", "message to send")
	flag.Parse()

	reply, remote, err := echo6.Echo(*addr, *message)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("connected to %s (IPv6)\n", remote)
	fmt.Printf("sent:     %q\n", *message)
	fmt.Printf("received: %q\n", reply)
}
