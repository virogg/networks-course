package server

import "time"

type (
	config struct {
		Port       int
		DeadAfter  time.Duration
		CheckEvery time.Duration
	}

	Option func(*config)
)

func WithPort(port int) Option {
	return func(cfg *config) {
		cfg.Port = port
	}
}

func WithDeadAfter(deadAfter time.Duration) Option {
	return func(cfg *config) {
		cfg.DeadAfter = deadAfter
	}
}

func WithCheckEvery(checkEvery time.Duration) Option {
	return func(cfg *config) {
		cfg.CheckEvery = checkEvery
	}
}
