package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yourorg/failure-uploader/internal/config"
	"github.com/yourorg/failure-uploader/internal/email"
	"github.com/yourorg/failure-uploader/internal/handlers"
	"github.com/yourorg/failure-uploader/internal/logging"
	"github.com/yourorg/failure-uploader/internal/router"
	"github.com/yourorg/failure-uploader/internal/s3client"
)

var httpHandler http.Handler

func init() {
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
		Msg("initializing failure-uploader")

	// Initialize S3 presigner
	presigner, err := s3client.NewPresigner(ctx, cfg.BucketName, cfg.AWSRegion, cfg.PresignTTL)
	if err != nil {
		logging.Error().Err(err).Msg("failed to initialize S3 presigner")
		panic(err)
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
	httpHandler = router.New(cfg, h)
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Convert API Gateway request to http.Request
	httpReq, err := convertRequest(ctx, req)
	if err != nil {
		logging.Error().Err(err).Msg("failed to convert request")
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       `{"error":"Internal server error"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	// Create response writer
	rw := &responseWriter{
		headers: make(http.Header),
		body:    make([]byte, 0),
		status:  200,
	}

	// Handle request
	httpHandler.ServeHTTP(rw, httpReq)

	// Convert response
	return events.APIGatewayV2HTTPResponse{
		StatusCode: rw.status,
		Body:       string(rw.body),
		Headers:    flattenHeaders(rw.headers),
	}, nil
}

func main() {
	lambda.Start(handler)
}

// convertRequest converts API Gateway request to http.Request
func convertRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		req.RequestContext.HTTP.Method,
		req.RawPath,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set body
	if req.Body != "" {
		httpReq.Body = &stringReader{s: req.Body, i: 0}
	}

	// Set query parameters
	q := httpReq.URL.Query()
	for k, v := range req.QueryStringParameters {
		q.Set(k, v)
	}
	httpReq.URL.RawQuery = q.Encode()

	return httpReq, nil
}

// responseWriter implements http.ResponseWriter for Lambda
type responseWriter struct {
	headers http.Header
	body    []byte
	status  int
}

func (rw *responseWriter) Header() http.Header {
	return rw.headers
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return len(b), nil
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
}

// stringReader implements io.ReadCloser for request body
type stringReader struct {
	s string
	i int
}

func (sr *stringReader) Read(p []byte) (n int, err error) {
	if sr.i >= len(sr.s) {
		return 0, nil
	}
	n = copy(p, sr.s[sr.i:])
	sr.i += n
	return n, nil
}

func (sr *stringReader) Close() error {
	return nil
}

// flattenHeaders converts http.Header to map[string]string
func flattenHeaders(h http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
