package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const requestIDKey contextType = "requestID"

// Generates a unique request ID for each request and adds it to:
// 1. The request context (for use in handlers/services)
// 2. The response header (X-Request-ID)
// 3. Log messages (via context)
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate or use existing request ID from header
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add request ID to response header.
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context.
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Extracts the request ID from the context. Returns empty string if not found.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// Generates a unique request ID using crypto/rand.
func generateRequestID() string {
	b := make([]byte, 8) // 16 hex characters
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return "fallback-id"
	}
	return hex.EncodeToString(b)
}
