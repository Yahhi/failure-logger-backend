package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/failure-uploader/internal/config"
	"github.com/yourorg/failure-uploader/internal/email"
	"github.com/yourorg/failure-uploader/internal/keys"
	"github.com/yourorg/failure-uploader/internal/logging"
	"github.com/yourorg/failure-uploader/internal/models"
	"github.com/yourorg/failure-uploader/internal/s3client"
	"github.com/yourorg/failure-uploader/internal/validation"
)

// Handler contains dependencies for HTTP handlers
type Handler struct {
	cfg       *config.Config
	presigner *s3client.Presigner
	emailer   *email.Sender
}

// NewHandler creates a new handler with dependencies
func NewHandler(cfg *config.Config, presigner *s3client.Presigner, emailer *email.Sender) *Handler {
	return &Handler{
		cfg:       cfg,
		presigner: presigner,
		emailer:   emailer,
	}
}

// UploadTicket handles POST /v1/upload-ticket
func (h *Handler) UploadTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.UploadTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_json", "Failed to parse request body", err.Error())
		return
	}

	// Validate request
	if errs := validation.ValidateUploadTicketRequest(&req, h.cfg); len(errs) > 0 {
		h.writeValidationErrors(w, errs)
		return
	}

	// Generate failure ID and build keys
	failureID := uuid.New().String()
	keyBuilder := keys.NewBuilder(req.Project, req.Env, failureID)

	logging.Info().
		Str("failureId", failureID).
		Str("project", req.Project).
		Str("env", req.Env).
		Msg("creating upload ticket")

	// Generate presigned URLs
	uploads, err := h.generatePresignedURLs(ctx, keyBuilder, &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "presign_failed", "Failed to generate presigned URLs", "")
		return
	}

	resp := models.UploadTicketResponse{
		FailureID:        failureID,
		S3Prefix:         keyBuilder.Prefix(),
		Uploads:          *uploads,
		ExpiresInSeconds: int(h.cfg.PresignTTL.Seconds()),
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// UploadComplete handles POST /v1/upload-complete
func (h *Handler) UploadComplete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.UploadCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_json", "Failed to parse request body", err.Error())
		return
	}

	// Validate request
	if errs := validation.ValidateUploadCompleteRequest(&req); len(errs) > 0 {
		h.writeValidationErrors(w, errs)
		return
	}

	logging.Info().
		Str("failureId", req.FailureID).
		Str("project", req.Project).
		Str("env", req.Env).
		Int("uploadedKeys", len(req.UploadedKeys)).
		Msg("processing upload complete")

	// Verify all uploaded keys exist in S3
	missing, err := h.presigner.VerifyObjectsExist(ctx, req.UploadedKeys)
	if err != nil {
		logging.Error().Err(err).Msg("failed to verify objects")
		h.writeError(w, http.StatusInternalServerError, "verification_failed", "Failed to verify uploaded objects", "")
		return
	}

	if len(missing) > 0 {
		logging.Warn().
			Str("failureId", req.FailureID).
			Strs("missing", missing).
			Msg("missing objects in S3")
		h.writeError(w, http.StatusBadRequest, "missing_objects", "Some objects were not found in S3", "")
		return
	}

	// Generate presigned GET URL for envelope
	keyBuilder := keys.NewBuilder(req.Project, req.Env, req.FailureID)
	envelopeURL, err := h.presigner.PresignGet(ctx, keyBuilder.Envelope())
	if err != nil {
		logging.Error().Err(err).Msg("failed to generate envelope URL")
		envelopeURL = "" // Continue without URL
	}

	// Send email notification
	if h.emailer != nil {
		notif := email.FailureNotification{
			FailureID:   req.FailureID,
			Project:     req.Project,
			Env:         req.Env,
			EnvelopeURL: envelopeURL,
		}

		if err := h.emailer.SendFailureNotification(ctx, notif); err != nil {
			logging.Error().Err(err).Msg("failed to send email notification")
			// Don't fail the request if email fails
		}
	}

	logging.Info().
		Str("failureId", req.FailureID).
		Msg("upload complete processed successfully")

	h.writeJSON(w, http.StatusOK, models.UploadCompleteResponse{Status: "ok"})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) generatePresignedURLs(ctx context.Context, kb *keys.Builder, req *models.UploadTicketRequest) (*models.UploadURLs, error) {
	uploads := &models.UploadURLs{}

	// Envelope
	url, err := h.presigner.PresignPut(ctx, kb.Envelope(), "application/json")
	if err != nil {
		return nil, err
	}
	uploads.Envelope = models.PresignedUpload{Key: kb.Envelope(), PutURL: url}

	// Request raw
	contentType := req.Request.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	url, err = h.presigner.PresignPut(ctx, kb.RequestRaw(), contentType)
	if err != nil {
		return nil, err
	}
	uploads.RequestRaw = models.PresignedUpload{Key: kb.RequestRaw(), PutURL: url}

	// Request headers
	url, err = h.presigner.PresignPut(ctx, kb.RequestHeaders(), "application/json")
	if err != nil {
		return nil, err
	}
	uploads.RequestHeaders = models.PresignedUpload{Key: kb.RequestHeaders(), PutURL: url}

	// Response raw
	url, err = h.presigner.PresignPut(ctx, kb.ResponseRaw(), "application/octet-stream")
	if err != nil {
		return nil, err
	}
	uploads.ResponseRaw = models.PresignedUpload{Key: kb.ResponseRaw(), PutURL: url}

	// Checksums
	url, err = h.presigner.PresignPut(ctx, kb.Checksums(), "application/json")
	if err != nil {
		return nil, err
	}
	uploads.Checksums = models.PresignedUpload{Key: kb.Checksums(), PutURL: url}

	// Files
	for _, file := range req.Request.Files {
		ct := file.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		url, err = h.presigner.PresignPut(ctx, kb.File(file.Filename), ct)
		if err != nil {
			return nil, err
		}
		uploads.Files = append(uploads.Files, models.PresignedUpload{
			Key:    kb.File(file.Filename),
			PutURL: url,
		})
	}

	return uploads, nil
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message, details string) {
	resp := models.ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	}
	h.writeJSON(w, status, resp)
}

func (h *Handler) writeValidationErrors(w http.ResponseWriter, errs []validation.ValidationError) {
	var messages []string
	for _, e := range errs {
		messages = append(messages, e.Error())
	}
	h.writeError(w, http.StatusBadRequest, "validation_error", "Validation failed", "")
}
