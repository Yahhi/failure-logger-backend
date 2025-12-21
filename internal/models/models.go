package models

import "time"

// UploadTicketRequest is the input for POST /v1/upload-ticket
type UploadTicketRequest struct {
	Project string      `json:"project"`
	Env     string      `json:"env"`
	Request RequestInfo `json:"request"`
	Client  ClientInfo  `json:"client"`
}

type RequestInfo struct {
	Method      string     `json:"method"`
	URL         string     `json:"url"`
	ContentType string     `json:"contentType"`
	BodyBytes   int64      `json:"bodyBytes"`
	Files       []FileInfo `json:"files,omitempty"`
}

type FileInfo struct {
	Name        string `json:"name"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Bytes       int64  `json:"bytes"`
}

type ClientInfo struct {
	AppVersion string `json:"appVersion"`
	Platform   string `json:"platform"`
}

// UploadTicketResponse is the output for POST /v1/upload-ticket
type UploadTicketResponse struct {
	FailureID        string     `json:"failureId"`
	S3Prefix         string     `json:"s3Prefix"`
	Uploads          UploadURLs `json:"uploads"`
	ExpiresInSeconds int        `json:"expiresInSeconds"`
}

type UploadURLs struct {
	Envelope       PresignedUpload   `json:"envelope"`
	RequestRaw     PresignedUpload   `json:"requestRaw"`
	RequestHeaders PresignedUpload   `json:"requestHeaders"`
	ResponseRaw    PresignedUpload   `json:"responseRaw"`
	Files          []PresignedUpload `json:"files,omitempty"`
	Checksums      PresignedUpload   `json:"checksums"`
}

type PresignedUpload struct {
	Key    string `json:"key"`
	PutURL string `json:"putUrl"`
}

// UploadCompleteRequest is the input for POST /v1/upload-complete
type UploadCompleteRequest struct {
	FailureID    string            `json:"failureId"`
	Project      string            `json:"project"`
	Env          string            `json:"env"`
	UploadedKeys []string          `json:"uploadedKeys"`
	SHA256       map[string]string `json:"sha256,omitempty"`
}

// UploadCompleteResponse is the output for POST /v1/upload-complete
type UploadCompleteResponse struct {
	Status string `json:"status"`
}

// Envelope is the metadata stored in envelope.json
type Envelope struct {
	FailureID string      `json:"failureId"`
	Project   string      `json:"project"`
	Env       string      `json:"env"`
	Request   RequestInfo `json:"request"`
	Client    ClientInfo  `json:"client"`
	CreatedAt time.Time   `json:"createdAt"`
	S3Prefix  string      `json:"s3Prefix"`
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
