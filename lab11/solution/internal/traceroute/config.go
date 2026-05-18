package traceroute

import "time"

type (
	config struct {
		Host    string
		Queries int
		MaxHops int
		Timeout time.Duration
		Resolve bool // reverse-DNS each hop (task Б)
		Verbose bool
	}

	Option func(*config)
)

func defaultConfig() config {
	return config{
		Queries: 3,
		MaxHops: 30,
		Timeout: 2 * time.Second,
		Resolve: true,
	}
}

func WithHost(host string) Option        { return func(c *config) { c.Host = host } }
func WithQueries(n int) Option           { return func(c *config) { c.Queries = n } }
func WithMaxHops(n int) Option           { return func(c *config) { c.MaxHops = n } }
func WithTimeout(d time.Duration) Option { return func(c *config) { c.Timeout = d } }
func WithResolve(v bool) Option          { return func(c *config) { c.Resolve = v } }
func WithVerbose(v bool) Option          { return func(c *config) { c.Verbose = v } }
