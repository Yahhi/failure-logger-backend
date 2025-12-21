# failure-uploader

A production-ready Go backend service for mobile apps to upload failed network request bundles to S3 using presigned URLs. After upload completion, the service validates the bundle and emails the project owner.

## Features

- **Presigned URL Generation**: Secure S3 uploads without exposing AWS credentials to clients
- **Email Notifications**: SES-based notifications when uploads complete
- **API Key Authentication**: Optional API key auth via `X-Api-Key` header
- **Size Validation**: Configurable limits for body, file, and total upload sizes
- **Structured Logging**: JSON logs for production, pretty logs for development
- **Lambda Ready**: Deployable to AWS Lambda behind API Gateway (HTTP API)
- **Local Development**: Standalone HTTP server for local testing

## Project Structure

```
failure-uploader/
├── api/
│   └── openapi.yaml     # OpenAPI 3.0 specification
├── cmd/
│   ├── lambda/          # Lambda entry point
│   │   └── main.go
│   └── server/          # Standalone HTTP server
│       └── main.go
├── internal/
│   ├── config/          # Environment configuration
│   ├── email/           # SES email sender
│   ├── handlers/        # HTTP handlers
│   ├── keys/            # S3 key builder
│   ├── logging/         # Structured logging
│   ├── middleware/      # Auth & request logging
│   ├── models/          # Request/response types
│   ├── router/          # HTTP routing
│   ├── s3client/        # S3 presigner
│   └── validation/      # Input validation
├── .env.example         # Environment variables template
├── Makefile
├── go.mod
└── README.md
```

## Configuration

Copy `.env.example` to `.env` and configure your environment:

```bash
cp .env.example .env
# Edit .env with your values
```

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `BUCKET_NAME` | S3 bucket for uploads | `failure-uploads` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `SES_FROM` | Sender email address | `noreply@example.com` |
| `SES_TO` | Recipient email address | `owner@example.com` |
| `PRESIGN_TTL_SECONDS` | Presigned URL expiration | `900` (15 min) |
| `API_KEY` | API key for authentication | (empty) |
| `STAGE` | Deployment stage (dev/staging/prod) | `dev` |
| `MAX_BODY_BYTES` | Max request body size | `10485760` (10MB) |
| `MAX_FILE_BYTES` | Max single file size | `52428800` (50MB) |
| `MAX_TOTAL_BYTES` | Max total upload size | `104857600` (100MB) |
| `PORT` | Server port (server mode only) | `8080` |

**Note**: Auth is disabled when `STAGE=dev` or `API_KEY` is empty.

## API Endpoints

### Health Check

```
GET /health
```

Response:
```json
{"status": "healthy", "time": "2024-03-15T10:30:00Z"}
```

### Create Upload Ticket

```
POST /v1/upload-ticket
```

Request:
```json
{
  "project": "myapp",
  "env": "prod",
  "request": {
    "method": "POST",
    "url": "https://api.example.com/v1/submit",
    "contentType": "application/json",
    "bodyBytes": 12345,
    "files": [
      {"name": "photo", "filename": "a.jpg", "contentType": "image/jpeg", "bytes": 345678}
    ]
  },
  "client": {"appVersion": "1.2.3", "platform": "ios"}
}
```

Response:
```json
{
  "failureId": "550e8400-e29b-41d4-a716-446655440000",
  "s3Prefix": "failures/myapp/prod/2024/03/15/550e8400-e29b-41d4-a716-446655440000/",
  "uploads": {
    "envelope": {"key": "failures/.../envelope.json", "putUrl": "https://..."},
    "requestRaw": {"key": "failures/.../request.raw", "putUrl": "https://..."},
    "requestHeaders": {"key": "failures/.../request.headers.json", "putUrl": "https://..."},
    "responseRaw": {"key": "failures/.../response.raw", "putUrl": "https://..."},
    "files": [{"key": "failures/.../files/a.jpg", "putUrl": "https://..."}],
    "checksums": {"key": "failures/.../checksums.json", "putUrl": "https://..."}
  },
  "expiresInSeconds": 900
}
```

### Complete Upload

```
POST /v1/upload-complete
```

Request:
```json
{
  "failureId": "550e8400-e29b-41d4-a716-446655440000",
  "project": "myapp",
  "env": "prod",
  "uploadedKeys": [
    "failures/myapp/prod/2024/03/15/550e8400.../envelope.json",
    "failures/myapp/prod/2024/03/15/550e8400.../request.raw",
    "failures/myapp/prod/2024/03/15/550e8400.../request.headers.json",
    "failures/myapp/prod/2024/03/15/550e8400.../checksums.json"
  ],
  "sha256": {
    "failures/.../envelope.json": "abc123...",
    "failures/.../request.raw": "def456..."
  }
}
```

Response:
```json
{"status": "ok"}
```

## Quick Start

### Prerequisites

- Go 1.22+
- AWS credentials configured (for S3 and SES access)
- S3 bucket created
- SES email addresses verified (in sandbox mode)

### Build

```bash
# Download dependencies
make deps

# Run tests
make test

# Build binaries
make build
```

### Run Locally

```bash
# Start development server
make run

# Or with custom port
PORT=3000 make run
```

### Deploy to Lambda

```bash
# Build and package for Lambda
make package-lambda

# Upload build/lambda/function.zip to AWS Lambda
```

## Example curl Requests

### Health Check

```bash
curl http://localhost:8080/health
```

### Create Upload Ticket

```bash
curl -X POST http://localhost:8080/v1/upload-ticket \
  -H "Content-Type: application/json" \
  -d '{
    "project": "myapp",
    "env": "prod",
    "request": {
      "method": "POST",
      "url": "https://api.example.com/v1/submit",
      "contentType": "application/json",
      "bodyBytes": 1024
    },
    "client": {"appVersion": "1.0.0", "platform": "ios"}
  }'
```

### Create Upload Ticket (with auth)

```bash
curl -X POST http://localhost:8080/v1/upload-ticket \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: your-secret-key" \
  -d '{
    "project": "myapp",
    "env": "prod",
    "request": {
      "method": "POST",
      "url": "https://api.example.com/v1/submit",
      "contentType": "application/json",
      "bodyBytes": 1024,
      "files": [
        {"name": "photo", "filename": "image.jpg", "contentType": "image/jpeg", "bytes": 50000}
      ]
    },
    "client": {"appVersion": "1.0.0", "platform": "android"}
  }'
```

### Upload File to Presigned URL

```bash
# Use the putUrl from the upload-ticket response
curl -X PUT "https://your-bucket.s3.amazonaws.com/..." \
  -H "Content-Type: application/json" \
  --data-binary @envelope.json
```

### Complete Upload

```bash
curl -X POST http://localhost:8080/v1/upload-complete \
  -H "Content-Type: application/json" \
  -d '{
    "failureId": "550e8400-e29b-41d4-a716-446655440000",
    "project": "myapp",
    "env": "prod",
    "uploadedKeys": [
      "failures/myapp/prod/2024/03/15/550e8400.../envelope.json",
      "failures/myapp/prod/2024/03/15/550e8400.../request.raw",
      "failures/myapp/prod/2024/03/15/550e8400.../request.headers.json",
      "failures/myapp/prod/2024/03/15/550e8400.../checksums.json"
    ]
  }'
```

## S3 Object Structure

```
failures/
└── {project}/
    └── {env}/
        └── YYYY/
            └── MM/
                └── DD/
                    └── {failureId}/
                        ├── envelope.json      # Metadata about the failure
                        ├── request.raw        # Raw request body
                        ├── request.headers.json # Request headers
                        ├── response.raw       # Raw response body (optional)
                        ├── checksums.json     # SHA256 checksums
                        └── files/
                            └── {filename}     # Attached files
```

## AWS IAM Policy

Minimum required permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:HeadObject"
      ],
      "Resource": "arn:aws:s3:::your-bucket-name/failures/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ses:SendEmail"
      ],
      "Resource": "*"
    }
  ]
}
```

## API Documentation

Full OpenAPI 3.0 specification is available at `api/openapi.yaml`.

### View with Swagger UI

You can view the API documentation using Swagger UI:

```bash
# Using Docker
docker run -p 8081:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/api:/api swaggerapi/swagger-ui

# Then open http://localhost:8081 in your browser
```

Or use online tools:
- [Swagger Editor](https://editor.swagger.io/) - paste the openapi.yaml content
- [Redocly](https://redocly.github.io/redoc/) - for a cleaner documentation view

## License

MIT
