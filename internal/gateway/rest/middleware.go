package rest

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func withLogging(innerHandler http.Handler, logger *slog.Logger) http.Handler {
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

func withCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
