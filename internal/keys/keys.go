package keys

import (
	"fmt"
	"path"
	"time"
)

// Builder constructs S3 keys for failure uploads
type Builder struct {
	project   string
	env       string
	failureID string
	date      time.Time
}

// NewBuilder creates a new key builder
func NewBuilder(project, env, failureID string) *Builder {
	return &Builder{
		project:   project,
		env:       env,
		failureID: failureID,
		date:      time.Now().UTC(),
	}
}

// WithDate sets a custom date (useful for testing)
func (b *Builder) WithDate(t time.Time) *Builder {
	b.date = t
	return b
}

// Prefix returns the S3 prefix for this failure
// Format: failures/{project}/{env}/YYYY/MM/DD/{failureId}/
func (b *Builder) Prefix() string {
	return fmt.Sprintf("failures/%s/%s/%s/%s/",
		b.project,
		b.env,
		b.date.Format("2006/01/02"),
		b.failureID,
	)
}

// Envelope returns the key for envelope.json
func (b *Builder) Envelope() string {
	return path.Join(b.Prefix(), "envelope.json")
}

// RequestRaw returns the key for request.raw
func (b *Builder) RequestRaw() string {
	return path.Join(b.Prefix(), "request.raw")
}

// RequestHeaders returns the key for request.headers.json
func (b *Builder) RequestHeaders() string {
	return path.Join(b.Prefix(), "request.headers.json")
}

// ResponseRaw returns the key for response.raw
func (b *Builder) ResponseRaw() string {
	return path.Join(b.Prefix(), "response.raw")
}

// Checksums returns the key for checksums.json
func (b *Builder) Checksums() string {
	return path.Join(b.Prefix(), "checksums.json")
}

// File returns the key for a file upload
func (b *Builder) File(filename string) string {
	return path.Join(b.Prefix(), "files", filename)
}

// RequiredKeys returns all required keys for a complete upload (excluding files)
func (b *Builder) RequiredKeys() []string {
	return []string{
		b.Envelope(),
		b.RequestRaw(),
		b.RequestHeaders(),
		b.Checksums(),
	}
}

// AllKeys returns all keys including files
func (b *Builder) AllKeys(filenames []string) []string {
	keys := b.RequiredKeys()
	keys = append(keys, b.ResponseRaw())
	for _, f := range filenames {
		keys = append(keys, b.File(f))
	}
	return keys
}
