package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/virogg/networks-course/lab11/solution/internal/dvrouting"
)

func main() {
	mode := flag.String("mode", "sync", "simulation mode: sync (round-based) or async (goroutine-per-node)")
	flag.Parse()

	topo := dvrouting.ExampleNetwork()
	fmt.Println("Network (lab11.md): links 0-1=1, 0-2=3, 0-3=7, 1-2=1, 2-3=2")
	fmt.Printf("Mode: %s\n\n", *mode)

	switch *mode {
	case "sync":
		sim := dvrouting.NewSyncSimulator(topo)
		rounds := sim.Run()
		fmt.Printf("=== Converged routing tables (%d rounds) ===\n", rounds)
		fmt.Print(dvrouting.FormatTables(sim.Tables()))

		fmt.Println("\n--- Changing link 1-2 cost: 1 -> 8 ---")
		rounds = sim.UpdateLinkCost(1, 2, 8)
		fmt.Printf("\n=== Recomputed routing tables (%d rounds) ===\n", rounds)
		fmt.Print(dvrouting.FormatTables(sim.Tables()))

	case "async":
		sim := dvrouting.NewAsyncSimulator(topo)
		sim.Run()
		fmt.Println("=== Converged routing tables ===")
		fmt.Print(dvrouting.FormatTables(sim.Tables()))

		fmt.Println("\n--- Changing link 1-2 cost: 1 -> 8 ---")
		sim.UpdateLinkCost(1, 2, 8)
		fmt.Println("\n=== Recomputed routing tables ===")
		fmt.Print(dvrouting.FormatTables(sim.Tables()))

	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q (use sync or async)\n", *mode)
		os.Exit(2)
	}
}
