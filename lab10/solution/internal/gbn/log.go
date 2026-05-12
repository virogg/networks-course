package gbn

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type EventLogger struct {
	mu    sync.Mutex
	out   io.Writer
	tag   string
	start time.Time
}

func NewLogger(out io.Writer, tag string) *EventLogger {
	return &EventLogger{out: out, tag: tag, start: time.Now()}
}

func (l *EventLogger) Event(event string, window string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.out, "[%s] t=%.3fs %s | %s\n", l.tag, time.Since(l.start).Seconds(), event, window)
}

func formatRange(lo, hi uint32) string {
	if hi <= lo {
		return "[]"
	}
	if hi == lo+1 {
		return fmt.Sprintf("[%d]", lo)
	}
	return fmt.Sprintf("[%d..%d]", lo, hi-1)
}
