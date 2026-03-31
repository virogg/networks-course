package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Entry struct {
	URL          string      `json:"url"`
	StatusCode   int         `json:"status_code"`
	Headers      http.Header `json:"headers"`
	ETag         string      `json:"etag"`
	LastModified string      `json:"last_modified"`
	CachedAt     time.Time   `json:"cached_at"`
	BodyFile     string      `json:"body_file"`
}

type Cache struct {
	mu      sync.RWMutex
	dir     string
	entries map[string]*Entry
}

func New(dir string) (*Cache, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	c := &Cache{dir: dir, entries: make(map[string]*Entry)}
	c.load()
	return c, nil
}

func urlKey(url string) string { return fmt.Sprintf("%x", md5.Sum([]byte(url))) }

func (c *Cache) metaPath() string { return filepath.Join(c.dir, "meta.json") }

func (c *Cache) load() {
	data, err := os.ReadFile(c.metaPath())
	if err != nil {
		return
	}
	json.Unmarshal(data, &c.entries) //nolint:errcheck
}

func (c *Cache) save() {
	data, _ := json.MarshalIndent(c.entries, "", "  ")
	os.WriteFile(c.metaPath(), data, 0o644) //nolint:errcheck
}

func (c *Cache) Get(url string) (*Entry, []byte, bool) {
	k := urlKey(url)
	c.mu.RLock()
	e, ok := c.entries[k]
	c.mu.RUnlock()
	if !ok {
		return nil, nil, false
	}
	body, err := os.ReadFile(e.BodyFile)
	if err != nil {
		c.mu.Lock()
		delete(c.entries, k)
		c.save()
		c.mu.Unlock()
		return nil, nil, false
	}
	return e, body, true
}

func (c *Cache) Set(url string, resp *http.Response, body []byte) {
	k := urlKey(url)
	bf := filepath.Join(c.dir, k+".body")
	os.WriteFile(bf, body, 0o644) //nolint:errcheck
	e := &Entry{
		URL:          url,
		StatusCode:   resp.StatusCode,
		Headers:      resp.Header.Clone(),
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		CachedAt:     time.Now(),
		BodyFile:     bf,
	}
	c.mu.Lock()
	c.entries[k] = e
	c.save()
	c.mu.Unlock()
}
