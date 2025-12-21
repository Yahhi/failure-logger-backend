package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/yourorg/failure-uploader/internal/config"
	"github.com/yourorg/failure-uploader/internal/handlers"
	"github.com/yourorg/failure-uploader/internal/middleware"
)

// New creates a new HTTP router with all routes configured
func New(cfg *config.Config, h *handlers.Handler) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestLogger)
	r.Use(middleware.CORS)

	// Health check (no auth required)
	r.Get("/health", h.HealthCheck)

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// Apply API key auth to v1 routes
		r.Use(middleware.APIKeyAuth(cfg.APIKey, cfg.AuthEnabled))

		r.Post("/upload-ticket", h.UploadTicket)
		r.Post("/upload-complete", h.UploadComplete)
	})

	return r
}
