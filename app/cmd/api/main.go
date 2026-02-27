package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/novelcore/kubeapp-go-rest/internal/config"
	"github.com/novelcore/kubeapp-go-rest/internal/server"
	"github.com/novelcore/kubeapp-go-rest/internal/version"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	log := newLogger(cfg)
	log.WithField("version", version.VERSION).Info("Starting kubeapp-go-rest")

	srv, err := server.New(cfg, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to create server")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.WithField("signal", sig).Info("Received shutdown signal")
		cancel()
	}()

	if err := srv.Start(ctx); err != nil && err != context.Canceled {
		log.WithError(err).Error("Server stopped with error")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Error during server shutdown")
		os.Exit(1)
	}

	log.Info("Server shutdown complete")
}

func newLogger(cfg *config.Config) *logrus.Logger {
	log := logrus.New()

	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	if cfg.Log.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return log
}
