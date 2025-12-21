package validation

import (
	"testing"

	"github.com/yourorg/failure-uploader/internal/config"
	"github.com/yourorg/failure-uploader/internal/models"
)

func TestValidateUploadTicketRequest(t *testing.T) {
	cfg := &config.Config{
		MaxBodyBytes:  10 * 1024 * 1024,  // 10MB
		MaxFileBytes:  50 * 1024 * 1024,  // 50MB
		MaxTotalBytes: 100 * 1024 * 1024, // 100MB
	}

	tests := []struct {
		name       string
		req        models.UploadTicketRequest
		wantErrors int
	}{
		{
			name: "valid request",
			req: models.UploadTicketRequest{
				Project: "myapp",
				Env:     "prod",
				Request: models.RequestInfo{
					Method:      "POST",
					URL:         "https://api.example.com/v1/submit",
					ContentType: "application/json",
					BodyBytes:   1024,
				},
				Client: models.ClientInfo{
					AppVersion: "1.2.3",
					Platform:   "ios",
				},
			},
			wantErrors: 0,
		},
		{
			name: "missing project",
			req: models.UploadTicketRequest{
				Env: "prod",
				Request: models.RequestInfo{
					Method: "POST",
					URL:    "https://api.example.com/v1/submit",
				},
			},
			wantErrors: 1,
		},
		{
			name: "missing env",
			req: models.UploadTicketRequest{
				Project: "myapp",
				Request: models.RequestInfo{
					Method: "POST",
					URL:    "https://api.example.com/v1/submit",
				},
			},
			wantErrors: 1,
		},
		{
			name: "invalid project format",
			req: models.UploadTicketRequest{
				Project: "my app with spaces",
				Env:     "prod",
				Request: models.RequestInfo{
					Method: "POST",
					URL:    "https://api.example.com/v1/submit",
				},
			},
			wantErrors: 1,
		},
		{
			name: "invalid HTTP method",
			req: models.UploadTicketRequest{
				Project: "myapp",
				Env:     "prod",
				Request: models.RequestInfo{
					Method: "INVALID",
					URL:    "https://api.example.com/v1/submit",
				},
			},
			wantErrors: 1,
		},
		{
			name: "invalid URL",
			req: models.UploadTicketRequest{
				Project: "myapp",
				Env:     "prod",
				Request: models.RequestInfo{
					Method: "POST",
					URL:    "not-a-url",
				},
			},
			wantErrors: 1,
		},
		{
			name: "body too large",
			req: models.UploadTicketRequest{
				Project: "myapp",
				Env:     "prod",
				Request: models.RequestInfo{
					Method:    "POST",
					URL:       "https://api.example.com/v1/submit",
					BodyBytes: 20 * 1024 * 1024, // 20MB > 10MB limit
				},
			},
			wantErrors: 1,
		},
		{
			name: "file too large",
			req: models.UploadTicketRequest{
				Project: "myapp",
				Env:     "prod",
				Request: models.RequestInfo{
					Method: "POST",
					URL:    "https://api.example.com/v1/submit",
					Files: []models.FileInfo{
						{
							Name:     "photo",
							Filename: "a.jpg",
							Bytes:    60 * 1024 * 1024, // 60MB > 50MB limit
						},
					},
				},
			},
			wantErrors: 1,
		},
		{
			name: "total too large",
			req: models.UploadTicketRequest{
				Project: "myapp",
				Env:     "prod",
				Request: models.RequestInfo{
					Method:    "POST",
					URL:       "https://api.example.com/v1/submit",
					BodyBytes: 10 * 1024 * 1024,
					Files: []models.FileInfo{
						{Filename: "a.jpg", Bytes: 50 * 1024 * 1024},
						{Filename: "b.jpg", Bytes: 50 * 1024 * 1024},
					},
				},
			},
			wantErrors: 1,
		},
		{
			name: "invalid platform",
			req: models.UploadTicketRequest{
				Project: "myapp",
				Env:     "prod",
				Request: models.RequestInfo{
					Method: "POST",
					URL:    "https://api.example.com/v1/submit",
				},
				Client: models.ClientInfo{
					Platform: "windows",
				},
			},
			wantErrors: 1,
		},
		{
			name: "multiple errors",
			req: models.UploadTicketRequest{
				Project: "",
				Env:     "",
				Request: models.RequestInfo{
					Method: "",
					URL:    "",
				},
			},
			wantErrors: 4, // project, env, method, url
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateUploadTicketRequest(&tt.req, cfg)
			if len(errs) != tt.wantErrors {
				t.Errorf("ValidateUploadTicketRequest() returned %d errors, want %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  - %s", e.Error())
				}
			}
		})
	}
}

func TestValidateUploadCompleteRequest(t *testing.T) {
	tests := []struct {
		name       string
		req        models.UploadCompleteRequest
		wantErrors int
	}{
		{
			name: "valid request",
			req: models.UploadCompleteRequest{
				FailureID:    "abc-123",
				Project:      "myapp",
				Env:          "prod",
				UploadedKeys: []string{"key1", "key2"},
			},
			wantErrors: 0,
		},
		{
			name: "missing failureId",
			req: models.UploadCompleteRequest{
				Project:      "myapp",
				Env:          "prod",
				UploadedKeys: []string{"key1"},
			},
			wantErrors: 1,
		},
		{
			name: "missing uploadedKeys",
			req: models.UploadCompleteRequest{
				FailureID: "abc-123",
				Project:   "myapp",
				Env:       "prod",
			},
			wantErrors: 1,
		},
		{
			name:       "all missing",
			req:        models.UploadCompleteRequest{},
			wantErrors: 4, // failureId, project, env, uploadedKeys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateUploadCompleteRequest(&tt.req)
			if len(errs) != tt.wantErrors {
				t.Errorf("ValidateUploadCompleteRequest() returned %d errors, want %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  - %s", e.Error())
				}
			}
		})
	}
}
