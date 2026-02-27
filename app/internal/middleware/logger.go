package middleware

import (
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// StructuredLogger creates a middleware for structured logging with request context
func StructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			reqID := chimiddleware.GetReqID(r.Context())

			logger.WithFields(logrus.Fields{
				"request_id": reqID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"remote_ip":  r.RemoteAddr,
			}).Debug("Request started")

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			logger.WithFields(logrus.Fields{
				"request_id": reqID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     ww.Status(),
				"bytes":      ww.BytesWritten(),
				"duration":   duration.Milliseconds(),
			}).Info("Request completed")
		})
	}
}
