package middleware

import (
	"net/http"

	"github.com/pajtaand/dmap-zero/internal/controller/metrics"
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.RESTHTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()

		next.ServeHTTP(w, r)
	})
}
