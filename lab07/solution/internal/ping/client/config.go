package client

import "time"

type (
	config struct {
		Host    string
		Port    int
		Count   int
		Timeout time.Duration
		Stats   bool
	}

	Option func(*config)
)

func WithHost(host string) Option {
	return func(c *config) {
		c.Host = host
	}
}

func WithPort(port int) Option {
	return func(c *config) {
		c.Port = port
	}
}

func WithCount(count int) Option {
	return func(c *config) {
		c.Count = count
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.Timeout = timeout
	}
}

func WithStats(stats bool) Option {
	return func(c *config) {
		c.Stats = stats
	}
}
