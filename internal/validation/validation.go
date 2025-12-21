package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yourorg/failure-uploader/internal/config"
	"github.com/yourorg/failure-uploader/internal/models"
)

var (
	projectRegex  = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
	envRegex      = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
	platformRegex = regexp.MustCompile(`^(ios|android|web|desktop)$`)
	methodRegex   = regexp.MustCompile(`^(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateUploadTicketRequest validates the upload ticket request
func ValidateUploadTicketRequest(req *models.UploadTicketRequest, cfg *config.Config) []ValidationError {
	var errors []ValidationError

	// Project validation
	if req.Project == "" {
		errors = append(errors, ValidationError{Field: "project", Message: "required"})
	} else if !projectRegex.MatchString(req.Project) {
		errors = append(errors, ValidationError{Field: "project", Message: "invalid format (alphanumeric, underscore, hyphen, max 64 chars)"})
	}

	// Env validation
	if req.Env == "" {
		errors = append(errors, ValidationError{Field: "env", Message: "required"})
	} else if !envRegex.MatchString(req.Env) {
		errors = append(errors, ValidationError{Field: "env", Message: "invalid format (alphanumeric, underscore, hyphen, max 32 chars)"})
	}

	// Request validation
	if req.Request.Method == "" {
		errors = append(errors, ValidationError{Field: "request.method", Message: "required"})
	} else if !methodRegex.MatchString(strings.ToUpper(req.Request.Method)) {
		errors = append(errors, ValidationError{Field: "request.method", Message: "invalid HTTP method"})
	}

	if req.Request.URL == "" {
		errors = append(errors, ValidationError{Field: "request.url", Message: "required"})
	} else if !strings.HasPrefix(req.Request.URL, "http://") && !strings.HasPrefix(req.Request.URL, "https://") {
		errors = append(errors, ValidationError{Field: "request.url", Message: "must be a valid HTTP(S) URL"})
	}

	// Size validation
	if req.Request.BodyBytes < 0 {
		errors = append(errors, ValidationError{Field: "request.bodyBytes", Message: "cannot be negative"})
	} else if req.Request.BodyBytes > cfg.MaxBodyBytes {
		errors = append(errors, ValidationError{Field: "request.bodyBytes", Message: fmt.Sprintf("exceeds maximum allowed size (%d bytes)", cfg.MaxBodyBytes)})
	}

	// Files validation
	var totalFileBytes int64
	for i, file := range req.Request.Files {
		if file.Filename == "" {
			errors = append(errors, ValidationError{Field: fmt.Sprintf("request.files[%d].filename", i), Message: "required"})
		}
		if file.Bytes < 0 {
			errors = append(errors, ValidationError{Field: fmt.Sprintf("request.files[%d].bytes", i), Message: "cannot be negative"})
		} else if file.Bytes > cfg.MaxFileBytes {
			errors = append(errors, ValidationError{Field: fmt.Sprintf("request.files[%d].bytes", i), Message: fmt.Sprintf("exceeds maximum allowed size (%d bytes)", cfg.MaxFileBytes)})
		}
		totalFileBytes += file.Bytes
	}

	// Total size validation
	totalBytes := req.Request.BodyBytes + totalFileBytes
	if totalBytes > cfg.MaxTotalBytes {
		errors = append(errors, ValidationError{Field: "totalBytes", Message: fmt.Sprintf("total upload size exceeds maximum (%d bytes)", cfg.MaxTotalBytes)})
	}

	// Client validation
	if req.Client.Platform != "" && !platformRegex.MatchString(strings.ToLower(req.Client.Platform)) {
		errors = append(errors, ValidationError{Field: "client.platform", Message: "must be one of: ios, android, web, desktop"})
	}

	return errors
}

// ValidateUploadCompleteRequest validates the upload complete request
func ValidateUploadCompleteRequest(req *models.UploadCompleteRequest) []ValidationError {
	var errors []ValidationError

	if req.FailureID == "" {
		errors = append(errors, ValidationError{Field: "failureId", Message: "required"})
	}

	if req.Project == "" {
		errors = append(errors, ValidationError{Field: "project", Message: "required"})
	} else if !projectRegex.MatchString(req.Project) {
		errors = append(errors, ValidationError{Field: "project", Message: "invalid format"})
	}

	if req.Env == "" {
		errors = append(errors, ValidationError{Field: "env", Message: "required"})
	} else if !envRegex.MatchString(req.Env) {
		errors = append(errors, ValidationError{Field: "env", Message: "invalid format"})
	}

	if len(req.UploadedKeys) == 0 {
		errors = append(errors, ValidationError{Field: "uploadedKeys", Message: "required"})
	}

	return errors
}
