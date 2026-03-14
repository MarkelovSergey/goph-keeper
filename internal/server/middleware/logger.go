// Package middleware содержит HTTP-мидлвары сервера.
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logger логирует метод, путь, статус и длительность каждого запроса.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		slog.Info("HTTP запрос",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", time.Since(start),
		)
	})
}

// responseWriter оборачивает http.ResponseWriter для перехвата кода статуса.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
