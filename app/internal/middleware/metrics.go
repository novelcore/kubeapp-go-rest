package middleware

import (
	"net/http"
	"strconv"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kubeapp_go_rest_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubeapp_go_rest_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
)

// Metrics creates a middleware for Prometheus metrics collection
func Metrics() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(ww.Status())
			endpoint := simplifyEndpoint(r.URL.Path)

			httpDuration.WithLabelValues(r.Method, endpoint, status).Observe(duration)
			httpRequests.WithLabelValues(r.Method, endpoint, status).Inc()
		})
	}
}

func simplifyEndpoint(path string) string {
	switch {
	case path == "/healthz":
		return "health_liveness"
	case path == "/readyz":
		return "health_readiness"
	case path == "/metrics":
		return "metrics"
	default:
		return "api"
	}
}
