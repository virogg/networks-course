package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/virogg/networks-course/solution/pkg/blacklist"
	"github.com/virogg/networks-course/solution/pkg/cache"
	"github.com/virogg/networks-course/solution/pkg/logger"
	"github.com/virogg/networks-course/solution/pkg/proxy"
)

type Config struct {
	Logger    logger.Logger
	Cache     *cache.Cache        // nil = no caching
	Blacklist blacklist.Blacklist // nil = no blacklist
}

func Handler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		target := proxy.ExtractTarget(r)

		// blacklist
		if cfg.Blacklist != nil && cfg.Blacklist.IsBlocked(target) {
			cfg.Logger.Info(target+" [BLOCKED]", logger.NewField("code", http.StatusForbidden))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			resp, err := blacklist.Respond(target)
			if err != nil {
				cfg.Logger.Info(fmt.Sprintf("%s [ERROR]\t%s", target, err), logger.NewField("code", http.StatusInternalServerError))
				return
			}
			fmt.Fprint(w, resp) //nolint:errcheck
			return
		}

		// cache
		if r.Method == http.MethodGet && cfg.Cache != nil {
			if entry, body, hit := cfg.Cache.Get(target); hit {
				if entry.ETag != "" || entry.LastModified != "" {
					extra := http.Header{}
					if entry.ETag != "" {
						extra.Set("If-None-Match", entry.ETag)
					}
					if entry.LastModified != "" {
						extra.Set("If-Modified-Since", entry.LastModified)
					}

					resp, freshBody, err := proxy.Forward(r, target, extra)
					if err != nil {
						log.Printf("upstream error, serving stale cache: %v", err)
						cfg.Logger.Info(target+" [cache-stale]", logger.NewField("code", entry.StatusCode))
						writeCached(w, entry, body)
						return
					}
					defer resp.Body.Close() //nolint:errcheck

					if resp.StatusCode == http.StatusNotModified {
						cfg.Logger.Info(target+" [cache-hit]", logger.NewField("code", entry.StatusCode))
						writeCached(w, entry, body)
						return
					}

					cfg.Cache.Set(target, resp, freshBody)
					cfg.Logger.Info(target+" [cache-updated]", logger.NewField("code", resp.StatusCode))
					proxy.WriteResp(w, resp, freshBody)
					return
				}

				cfg.Logger.Info(target+" [cache-hit]", logger.NewField("code", entry.StatusCode))
				writeCached(w, entry, body)
				return
			}
		}

		resp, body, err := proxy.Forward(r, target, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusBadGateway)
			cfg.Logger.Info(target, logger.NewField("code", http.StatusBadGateway))
			return
		}
		defer resp.Body.Close() //nolint:errcheck

		if r.Method == http.MethodGet && cfg.Cache != nil && resp.StatusCode == http.StatusOK {
			cfg.Cache.Set(target, resp, body)
		}

		cfg.Logger.Info(target, logger.NewField("code", resp.StatusCode))
		proxy.WriteResp(w, resp, body)
	}
}

func writeCached(w http.ResponseWriter, e *cache.Entry, body []byte) {
	for k, vs := range e.Headers {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(e.StatusCode)
	w.Write(body) //nolint:errcheck
}
