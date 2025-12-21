package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	BucketName    string
	AWSRegion     string
	SESFrom       string
	SESTo         string
	PresignTTL    time.Duration
	APIKey        string
	Stage         string
	MaxBodyBytes  int64
	MaxFileBytes  int64
	MaxTotalBytes int64
	AuthEnabled   bool
}

func Load() *Config {
	presignTTL := getEnvInt("PRESIGN_TTL_SECONDS", 900)
	apiKey := os.Getenv("API_KEY")

	return &Config{
		BucketName:    getEnv("BUCKET_NAME", "failure-uploads"),
		AWSRegion:     getEnv("AWS_REGION", "us-east-1"),
		SESFrom:       getEnv("SES_FROM", "noreply@example.com"),
		SESTo:         getEnv("SES_TO", "owner@example.com"),
		PresignTTL:    time.Duration(presignTTL) * time.Second,
		APIKey:        apiKey,
		Stage:         getEnv("STAGE", "dev"),
		MaxBodyBytes:  getEnvInt64("MAX_BODY_BYTES", 10*1024*1024),   // 10MB default
		MaxFileBytes:  getEnvInt64("MAX_FILE_BYTES", 50*1024*1024),   // 50MB default
		MaxTotalBytes: getEnvInt64("MAX_TOTAL_BYTES", 100*1024*1024), // 100MB default
		AuthEnabled:   apiKey != "" && getEnv("STAGE", "dev") != "dev",
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return defaultVal
}
