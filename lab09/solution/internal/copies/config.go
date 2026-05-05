package copies

import "time"

type (
	config struct {
		Port           int
		Interval       time.Duration
		DeadMultiplier int
	}

	Option func(*config)
)

func WithPort(p int) Option               { return func(c *config) { c.Port = p } }
func WithInterval(d time.Duration) Option { return func(c *config) { c.Interval = d } }
func WithDeadMultiplier(m int) Option     { return func(c *config) { c.DeadMultiplier = m } }
