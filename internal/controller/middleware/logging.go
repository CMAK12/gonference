package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func WithLogging(innerHandler http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		innerHandler.ServeHTTP(w, r)
		duration := time.Since(start)

		logger.Info(
			fmt.Sprintf("%s %s request finished", r.Method, r.URL.Path),
			slog.Int64("elapsed", duration.Milliseconds()),
			slog.Int("status", http.StatusOK),
		)
	})
}
