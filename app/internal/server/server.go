package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/novelcore/kubeapp-go-rest/internal/auth"
	"github.com/novelcore/kubeapp-go-rest/internal/config"
	"github.com/novelcore/kubeapp-go-rest/internal/router"
	"github.com/sirupsen/logrus"
)

// Server represents the HTTP server
type Server struct {
	config       *config.Config
	log          *logrus.Logger
	httpServer   *http.Server
	ready        atomic.Bool
	jwtValidator *auth.JWTValidator
}

// New creates a new Server instance
func New(cfg *config.Config, log *logrus.Logger) (*Server, error) {
	s := &Server{
		config: cfg,
		log:    log,
	}

	s.ready.Store(false)

	// Initialize JWT validator only when ZITADEL_OIDC is enabled
	if cfg.Auth.ZitadelOIDC {
		if cfg.Auth.JWT.Issuer != "" && cfg.Auth.JWT.Audience != "" && cfg.Auth.JWT.JWKSURL != "" {
			log.Info("Initializing JWT authentication")
			s.jwtValidator = auth.NewJWTValidator(
				cfg.Auth.JWT.Issuer,
				cfg.Auth.JWT.Audience,
				cfg.Auth.JWT.JWKSURL,
				cfg.Auth.JWT.CacheTTL,
			)
			log.WithFields(logrus.Fields{
				"issuer":   cfg.Auth.JWT.Issuer,
				"audience": cfg.Auth.JWT.Audience,
				"jwksURL":  cfg.Auth.JWT.JWKSURL,
				"cacheTTL": cfg.Auth.JWT.CacheTTL,
			}).Info("JWT validator initialized")
		} else {
			log.Warn("ZITADEL_OIDC=true but JWT config is incomplete (missing issuer, audience, or JWKS URL)")
		}
	} else {
		log.Info("Zitadel OIDC disabled (ZITADEL_OIDC=false) — API routes are unauthenticated")
	}

	return s, nil
}

// Start starts the HTTP server and blocks until ctx is cancelled or a fatal error occurs
func (s *Server) Start(ctx context.Context) error {
	s.log.Info("Starting kubeapp-go-rest server")

	s.ready.Store(true)
	s.log.Info("Server is ready")

	r := router.New(s.config, s.log, &s.ready, s.jwtValidator)

	addr := fmt.Sprintf("%s:%d", s.config.Server.Address, s.config.Server.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
	}

	errChan := make(chan error, 1)
	go func() {
		s.log.WithField("address", addr).Info("HTTP server listening")
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("http server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down HTTP server")
	s.ready.Store(false)

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("error shutting down http server: %w", err)
		}
	}

	return nil
}
