package middleware

import (
	"context"
	"log"
)

// Logs a message with the request ID from context.
func LogWithRequestID(ctx context.Context, format string, v ...any) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		// Prepend requestID to the format string.
		// Build args slice: [requestID, v[0], v[1], ...]
		args := append([]any{requestID}, v...)
		log.Printf("[requestID=%s] "+format, args...)
	} else {
		log.Printf(format, v...)
	}
}

// Logs an error with the request ID from context.
func LogErrorWithRequestID(ctx context.Context, err error, message string) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		log.Printf("[requestID=%s] %s: %v", requestID, message, err)
	} else {
		log.Printf("%s: %v", message, err)
	}
}
