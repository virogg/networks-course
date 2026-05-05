package portscan

import (
	"context"
	"fmt"
	"net"
	"slices"
	"sync"
	"time"
)

type Result struct {
	Port int
	TCP  bool
	UDP  bool
}

func Scan(ctx context.Context, opts ...Option) ([]Result, Mode, error) {
	cfg := new(config)
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.Mode == "" {
		cfg.Mode = ModeAuto
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 1500 * time.Millisecond
	}

	if err := cfg.validate(); err != nil {
		return nil, "", err
	}

	mode := cfg.Mode
	if mode == ModeAuto {
		if isLocalIP(cfg.IP) {
			mode = ModeLocal
		} else {
			mode = ModeRemote
		}
	}
	if mode == ModeRemote && (cfg.Proto == ProtoUDP || cfg.Proto == ProtoBoth) {
		return nil, mode, fmt.Errorf("udp scanning is not supported in remote mode")
	}

	check := newChecker(mode, cfg.Timeout)

	ports := make(chan int)
	results := make(chan Result, cfg.To-cfg.From+1)

	var wg sync.WaitGroup
	for range cfg.Workers {
		wg.Go(func() {
			for p := range ports {
				r := Result{Port: p}
				if cfg.Proto == ProtoTCP || cfg.Proto == ProtoBoth {
					r.TCP = check.tcp(cfg.IP, p)
				}
				if cfg.Proto == ProtoUDP || cfg.Proto == ProtoBoth {
					r.UDP = check.udp(cfg.IP, p)
				}
				if r.TCP || r.UDP {
					results <- r
				}
			}
		})
	}

	go func() {
		defer close(ports)
		for p := cfg.From; p <= cfg.To; p++ {
			select {
			case <-ctx.Done():
				return
			case ports <- p:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]Result, 0, cfg.To-cfg.From+1)
	for r := range results {
		out = append(out, r)
	}
	slices.SortFunc(out, func(a, b Result) int { return a.Port - b.Port })
	return out, mode, nil
}

func isLocalIP(ip string) bool {
	target := net.ParseIP(ip)
	if target == nil {
		return false
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		if ipnet.IP.Equal(target) {
			return true
		}
	}
	return false
}
