package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/failure-uploader/internal/config"
	"github.com/yourorg/failure-uploader/internal/email"
	"github.com/yourorg/failure-uploader/internal/handlers"
	"github.com/yourorg/failure-uploader/internal/logging"
	"github.com/yourorg/failure-uploader/internal/router"
	"github.com/yourorg/failure-uploader/internal/s3client"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg := config.Load()

	// Initialize logging
	logging.Init(cfg.Stage)

	logging.Info().
		Str("bucket", cfg.BucketName).
		Str("region", cfg.AWSRegion).
		Str("stage", cfg.Stage).
		Bool("authEnabled", cfg.AuthEnabled).
		Msg("starting failure-uploader server")

	// Initialize S3 presigner
	presigner, err := s3client.NewPresigner(ctx, cfg.BucketName, cfg.AWSRegion, cfg.PresignTTL)
	if err != nil {
		logging.Error().Err(err).Msg("failed to initialize S3 presigner")
		os.Exit(1)
	}

	// Initialize email sender (optional - may fail in dev)
	var emailer *email.Sender
	emailer, err = email.NewSender(ctx, cfg.AWSRegion, cfg.SESFrom, cfg.SESTo)
	if err != nil {
		logging.Warn().Err(err).Msg("failed to initialize email sender - notifications disabled")
		emailer = nil
	}

	// Create handler and router
	h := handlers.NewHandler(cfg, presigner, emailer)
	httpHandler := router.New(cfg, h)

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logging.Info().Str("addr", server.Addr).Msg("server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error().Err(err).Msg("server error")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Info().Msg("shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logging.Error().Err(err).Msg("server forced to shutdown")
		os.Exit(1)
	}

	logging.Info().Msg("server stopped")
}
