package portscan

import (
	"fmt"
	"net"
	"time"
)

type (
	config struct {
		IP      string
		From    int
		To      int
		Proto   Proto
		Mode    Mode
		Workers int
		Timeout time.Duration
	}

	Option func(*config)
)

func (c config) validate() error {
	if c.From < 1 || c.From > 65535 {
		return fmt.Errorf("from out of range: %d", c.From)
	}
	if c.To < 1 || c.To > 65535 {
		return fmt.Errorf("to out of range: %d", c.To)
	}
	if c.From > c.To {
		return fmt.Errorf("from (%d) > to (%d)", c.From, c.To)
	}
	if net.ParseIP(c.IP) == nil {
		return fmt.Errorf("invalid ip %q", c.IP)
	}
	if c.Workers < 1 {
		return fmt.Errorf("workers must be >= 1")
	}
	return nil
}

func WithIP(ip string) Option {
	return func(c *config) {
		c.IP = ip
	}
}

func WithFrom(from int) Option {
	return func(c *config) {
		c.From = from
	}
}

func WithTo(to int) Option {
	return func(c *config) {
		c.To = to
	}
}

func WithProto(proto Proto) Option {
	return func(c *config) {
		c.Proto = proto
	}
}

func WithMode(mode Mode) Option {
	return func(c *config) {
		c.Mode = mode
	}
}

func WithWorkers(workers int) Option {
	return func(c *config) {
		c.Workers = workers
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.Timeout = timeout
	}
}
