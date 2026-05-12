package ping

import "time"

type (
	config struct {
		Host     string
		Interval time.Duration
		Timeout  time.Duration
		Count    int // 0 = unlimited
		Size     int // bytes of ICMP payload (>= 8 to fit timestamp)
	}

	Option func(*config)
)

func defaultConfig() config {
	return config{
		Interval: time.Second,
		Timeout:  time.Second,
		Size:     56,
	}
}

func WithHost(host string) Option         { return func(c *config) { c.Host = host } }
func WithInterval(d time.Duration) Option { return func(c *config) { c.Interval = d } }
func WithTimeout(d time.Duration) Option  { return func(c *config) { c.Timeout = d } }
func WithCount(n int) Option              { return func(c *config) { c.Count = n } }
func WithSize(n int) Option               { return func(c *config) { c.Size = n } }
