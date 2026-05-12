package gbn

import "time"

type (
	clientConfig struct {
		RemoteAddr string
		File       string
		ChunkSize  int
		WindowSize uint32
		Timeout    time.Duration
		LossRate   float64
		Logger     *EventLogger
	}

	ClientOption func(*clientConfig)

	serverConfig struct {
		ListenAddr string
		OutPath    string
		LossRate   float64
		Logger     *EventLogger
	}

	ServerOption func(*serverConfig)
)

func defaultClientConfig() clientConfig {
	return clientConfig{
		ChunkSize:  1024,
		WindowSize: 4,
		Timeout:    500 * time.Millisecond,
	}
}

func defaultServerConfig() serverConfig {
	return serverConfig{}
}

func WithClientRemoteAddr(addr string) ClientOption {
	return func(c *clientConfig) { c.RemoteAddr = addr }
}
func WithClientFile(f string) ClientOption       { return func(c *clientConfig) { c.File = f } }
func WithClientChunkSize(n int) ClientOption     { return func(c *clientConfig) { c.ChunkSize = n } }
func WithClientWindowSize(n uint32) ClientOption { return func(c *clientConfig) { c.WindowSize = n } }
func WithClientTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) { c.Timeout = d }
}
func WithClientLossRate(r float64) ClientOption { return func(c *clientConfig) { c.LossRate = r } }

func WithServerListenAddr(addr string) ServerOption {
	return func(c *serverConfig) { c.ListenAddr = addr }
}
func WithServerOutPath(p string) ServerOption   { return func(c *serverConfig) { c.OutPath = p } }
func WithServerLossRate(r float64) ServerOption { return func(c *serverConfig) { c.LossRate = r } }
