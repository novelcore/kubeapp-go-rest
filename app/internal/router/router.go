package router

import (
	"net/http"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/novelcore/kubeapp-go-rest/internal/auth"
	"github.com/novelcore/kubeapp-go-rest/internal/config"
	"github.com/novelcore/kubeapp-go-rest/internal/handlers"
	"github.com/novelcore/kubeapp-go-rest/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

// New creates a new Chi router with all routes and middleware configured
func New(cfg *config.Config, log *logrus.Logger, ready *atomic.Bool, jwtValidator *auth.JWTValidator) http.Handler {
	r := chi.NewRouter()

	// CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "https://*.kubecore.eu"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-User-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(corsMiddleware.Handler)

	// Global middleware
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.StructuredLogger(log))
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Compress(5))

	// Health and readiness endpoints
	r.Get("/healthz", handlers.Health())
	r.Get("/readyz", handlers.Readiness(ready))

	// Metrics endpoint (opt-in)
	if cfg.Observability.Enabled {
		r.Handle("/metrics", promhttp.Handler())
	}

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		if cfg.Observability.Enabled {
			r.Use(middleware.Metrics())
		}

		if cfg.Auth.ZitadelOIDC && jwtValidator != nil {
			r.Use(middleware.JWTAuth(jwtValidator, log))
		}

		// Example endpoint — replace with real business logic
		r.Get("/hello", handlers.Hello())
	})

	return r
}
