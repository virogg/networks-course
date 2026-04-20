package ping

import (
	"fmt"
	"time"
)

const (
	headerFmt = "\n--- %s ping statistics ---\n"
	statsFmt  = "%d packets transmitted, %d packets received, %.1f%% packet loss\n"
	rttFmt    = "round-trip min/avg/max = %v/%v/%v\n"
)

func PrintStats(addr string, sent, received int, rtts []time.Duration) {
	fmt.Printf(headerFmt, addr)
	loss := float64(sent-received) / float64(sent) * 100
	fmt.Printf(statsFmt, sent, received, loss)

	if len(rtts) == 0 {
		return
	}

	mn, mx, sum := rtts[0], rtts[0], time.Duration(0)
	for _, rtt := range rtts {
		mn = min(mn, rtt)
		mx = max(mx, rtt)
		sum += rtt
	}
	avg := sum / time.Duration(len(rtts))
	fmt.Printf(rttFmt, mn, avg, mx)
}
