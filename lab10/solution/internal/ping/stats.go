package ping

import (
	"fmt"
	"math"
	"slices"
	"time"
)

const (
	headerFmt = "--- %s ping statistics ---\n%d packets transmitted, %d received, %.0f%% packet loss, time %s"
	statsFmt  = "%s\nrtt mn/avg/mx/mdev = %.3f/%.3f/%.3f/%.3f ms"
)

type Stats struct {
	Sent     int
	Received int
	rtts     []time.Duration
}

func (s *Stats) Record(rtt time.Duration) {
	s.Received++
	s.rtts = append(s.rtts, rtt)
}

func (s *Stats) Summary(host string, elapsed time.Duration) string {
	loss := 0.0
	if s.Sent > 0 {
		loss = float64(s.Sent-s.Received) * 100.0 / float64(s.Sent)
	}
	header := fmt.Sprintf(headerFmt, host, s.Sent, s.Received, loss, elapsed.Round(time.Millisecond))
	if len(s.rtts) == 0 {
		return header
	}
	mn := slices.Min(s.rtts)
	mx := slices.Max(s.rtts)
	var sum time.Duration
	for _, r := range s.rtts {
		sum += r
	}
	avg := sum / time.Duration(len(s.rtts))
	var sq float64
	for _, r := range s.rtts {
		d := float64(r-avg) / float64(time.Millisecond)
		sq += d * d
	}
	mdev := math.Sqrt(sq / float64(len(s.rtts)))
	return fmt.Sprintf(statsFmt,
		header,
		float64(mn)/float64(time.Millisecond),
		float64(avg)/float64(time.Millisecond),
		float64(mx)/float64(time.Millisecond),
		mdev,
	)
}
