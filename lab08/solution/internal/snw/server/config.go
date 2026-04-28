package server

import "time"

type (
	config struct {
		Addr        string
		Timeout     time.Duration
		ChunkSize   int
		LossProb    float64
		CorruptProb float64
		SendFile    string
		RecvFile    string
		Seed        int64
	}

	Option func(*config)
)

func WithAddr(addr string) Option        { return func(c *config) { c.Addr = addr } }
func WithTimeout(t time.Duration) Option { return func(c *config) { c.Timeout = t } }
func WithChunkSize(n int) Option         { return func(c *config) { c.ChunkSize = n } }
func WithLossProb(p float64) Option      { return func(c *config) { c.LossProb = p } }
func WithCorruptProb(p float64) Option   { return func(c *config) { c.CorruptProb = p } }
func WithSendFile(path string) Option    { return func(c *config) { c.SendFile = path } }
func WithRecvFile(path string) Option    { return func(c *config) { c.RecvFile = path } }
func WithSeed(seed int64) Option         { return func(c *config) { c.Seed = seed } }
