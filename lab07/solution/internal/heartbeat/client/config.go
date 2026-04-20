package client

import "time"

type (
	config struct {
		Host     string
		Port     int
		Interval time.Duration
	}

	Option func(*config)
)

func WithHost(host string) Option {
	return func(cfg *config) {
		cfg.Host = host
	}
}

func WithPort(port int) Option {
	return func(cfg *config) {
		cfg.Port = port
	}
}

func WithInterval(interval time.Duration) Option {
	return func(cfg *config) {
		cfg.Interval = interval
	}
}
