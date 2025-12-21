package middleware

import (
	"net/http"

	"github.com/yourorg/failure-uploader/internal/logging"
)

const APIKeyHeader = "X-Api-Key"

// APIKeyAuth creates middleware that validates API key from header
func APIKeyAuth(apiKey string, enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth if disabled
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Get API key from header
			providedKey := r.Header.Get(APIKeyHeader)
			if providedKey == "" {
				logging.Warn().
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Msg("missing API key")
				http.Error(w, `{"error":"Missing API key","code":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Validate API key
			if providedKey != apiKey {
				logging.Warn().
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Msg("invalid API key")
				http.Error(w, `{"error":"Invalid API key","code":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs incoming requests
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote", r.RemoteAddr).
			Str("userAgent", r.UserAgent()).
			Msg("incoming request")

		next.ServeHTTP(w, r)
	})
}

// JSONContentType sets JSON content type for responses
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// CORS adds CORS headers for development
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Api-Key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
