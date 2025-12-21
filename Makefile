.PHONY: build build-lambda build-server test clean run deps lint

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Binary names
LAMBDA_BINARY=bootstrap
SERVER_BINARY=failure-uploader

# Build directories
BUILD_DIR=build
LAMBDA_DIR=$(BUILD_DIR)/lambda
SERVER_DIR=$(BUILD_DIR)/server

# Default target
all: deps test build

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
test:
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Build Lambda binary (for Amazon Linux 2)
build-lambda:
	mkdir -p $(LAMBDA_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(LAMBDA_DIR)/$(LAMBDA_BINARY) ./cmd/lambda

# Build server binary
build-server:
	mkdir -p $(SERVER_DIR)
	$(GOBUILD) -ldflags="-s -w" -o $(SERVER_DIR)/$(SERVER_BINARY) ./cmd/server

# Build both
build: build-lambda build-server

# Create Lambda deployment package
package-lambda: build-lambda
	cd $(LAMBDA_DIR) && zip -j function.zip $(LAMBDA_BINARY)

# Run local server
run:
	STAGE=dev \
	BUCKET_NAME=failure-uploads-dev \
	AWS_REGION=us-east-1 \
	SES_FROM=noreply@example.com \
	SES_TO=owner@example.com \
	$(GOCMD) run ./cmd/server

# Run with custom port
run-port:
	PORT=$(PORT) $(MAKE) run

# Format code
fmt:
	$(GOFMT) ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Show help
help:
	@echo "Available targets:"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  build          - Build both Lambda and server binaries"
	@echo "  build-lambda   - Build Lambda binary only"
	@echo "  build-server   - Build server binary only"
	@echo "  package-lambda - Create Lambda deployment ZIP"
	@echo "  run            - Run local development server"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  clean          - Remove build artifacts"
