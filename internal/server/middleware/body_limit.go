package middleware

import "net/http"

// maxBodyBytes — максимальный размер тела запроса (1 МБ).
const maxBodyBytes = 1 << 20

// MaxBodySize ограничивает размер тела входящего запроса, защищая от OOM-атак.
func MaxBodySize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		next.ServeHTTP(w, r)
	})
}
