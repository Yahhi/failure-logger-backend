package s3client

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/yourorg/failure-uploader/internal/logging"
)

// Presigner handles S3 presigned URL generation
type Presigner struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	ttl           time.Duration
}

// NewPresigner creates a new S3 presigner
func NewPresigner(ctx context.Context, bucket string, region string, ttl time.Duration) (*Presigner, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	return &Presigner{
		client:        client,
		presignClient: presignClient,
		bucket:        bucket,
		ttl:           ttl,
	}, nil
}

// PresignPut generates a presigned PUT URL for uploading
func (p *Presigner) PresignPut(ctx context.Context, key string, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	presignedReq, err := p.presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = p.ttl
	})
	if err != nil {
		logging.Error().Err(err).Str("key", key).Msg("failed to presign PUT URL")
		return "", err
	}

	return presignedReq.URL, nil
}

// PresignGet generates a presigned GET URL for downloading
func (p *Presigner) PresignGet(ctx context.Context, key string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}

	presignedReq, err := p.presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = p.ttl
	})
	if err != nil {
		logging.Error().Err(err).Str("key", key).Msg("failed to presign GET URL")
		return "", err
	}

	return presignedReq.URL, nil
}

// ObjectExists checks if an object exists in S3
func (p *Presigner) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}
	return true, nil
}

// VerifyObjectsExist checks if all specified keys exist in S3
func (p *Presigner) VerifyObjectsExist(ctx context.Context, keys []string) ([]string, error) {
	var missing []string
	for _, key := range keys {
		exists, err := p.ObjectExists(ctx, key)
		if err != nil {
			return nil, err
		}
		if !exists {
			missing = append(missing, key)
		}
	}
	return missing, nil
}

// Bucket returns the bucket name
func (p *Presigner) Bucket() string {
	return p.bucket
}
