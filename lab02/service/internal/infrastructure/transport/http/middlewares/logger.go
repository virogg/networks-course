package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/virogg/networks-course/service/pkg/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func Logger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			path := r.URL.Path
			statusCode := strconv.Itoa(rw.statusCode)
			duration := time.Since(start).Seconds()

			log.Info(
				"request completed",
				logger.NewField("method", r.Method),
				logger.NewField("path", path),
				logger.NewField("status", statusCode),
				logger.NewField("duration", duration),
			)
		})
	}
}
