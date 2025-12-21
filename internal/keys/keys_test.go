package keys

import (
	"testing"
	"time"
)

func TestBuilder_Prefix(t *testing.T) {
	tests := []struct {
		name      string
		project   string
		env       string
		failureID string
		date      time.Time
		want      string
	}{
		{
			name:      "basic prefix",
			project:   "myapp",
			env:       "prod",
			failureID: "abc-123",
			date:      time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
			want:      "failures/myapp/prod/2024/03/15/abc-123/",
		},
		{
			name:      "different date",
			project:   "testapp",
			env:       "staging",
			failureID: "xyz-789",
			date:      time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:      "failures/testapp/staging/2024/12/01/xyz-789/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(tt.project, tt.env, tt.failureID).WithDate(tt.date)
			got := b.Prefix()
			if got != tt.want {
				t.Errorf("Prefix() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuilder_Keys(t *testing.T) {
	date := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	b := NewBuilder("myapp", "prod", "abc-123").WithDate(date)

	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{
			name: "envelope",
			fn:   b.Envelope,
			want: "failures/myapp/prod/2024/03/15/abc-123/envelope.json",
		},
		{
			name: "request raw",
			fn:   b.RequestRaw,
			want: "failures/myapp/prod/2024/03/15/abc-123/request.raw",
		},
		{
			name: "request headers",
			fn:   b.RequestHeaders,
			want: "failures/myapp/prod/2024/03/15/abc-123/request.headers.json",
		},
		{
			name: "response raw",
			fn:   b.ResponseRaw,
			want: "failures/myapp/prod/2024/03/15/abc-123/response.raw",
		},
		{
			name: "checksums",
			fn:   b.Checksums,
			want: "failures/myapp/prod/2024/03/15/abc-123/checksums.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestBuilder_File(t *testing.T) {
	date := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	b := NewBuilder("myapp", "prod", "abc-123").WithDate(date)

	tests := []struct {
		filename string
		want     string
	}{
		{
			filename: "photo.jpg",
			want:     "failures/myapp/prod/2024/03/15/abc-123/files/photo.jpg",
		},
		{
			filename: "document.pdf",
			want:     "failures/myapp/prod/2024/03/15/abc-123/files/document.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := b.File(tt.filename)
			if got != tt.want {
				t.Errorf("File(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestBuilder_RequiredKeys(t *testing.T) {
	date := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	b := NewBuilder("myapp", "prod", "abc-123").WithDate(date)

	keys := b.RequiredKeys()

	if len(keys) != 4 {
		t.Errorf("RequiredKeys() returned %d keys, want 4", len(keys))
	}

	expected := map[string]bool{
		"failures/myapp/prod/2024/03/15/abc-123/envelope.json":        true,
		"failures/myapp/prod/2024/03/15/abc-123/request.raw":          true,
		"failures/myapp/prod/2024/03/15/abc-123/request.headers.json": true,
		"failures/myapp/prod/2024/03/15/abc-123/checksums.json":       true,
	}

	for _, k := range keys {
		if !expected[k] {
			t.Errorf("RequiredKeys() contains unexpected key %q", k)
		}
	}
}

func TestBuilder_AllKeys(t *testing.T) {
	date := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	b := NewBuilder("myapp", "prod", "abc-123").WithDate(date)

	filenames := []string{"a.jpg", "b.png"}
	keys := b.AllKeys(filenames)

	// 4 required + 1 response.raw + 2 files = 7
	if len(keys) != 7 {
		t.Errorf("AllKeys() returned %d keys, want 7", len(keys))
	}
}
