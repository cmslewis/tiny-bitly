package middleware

import (
	"context"
	"log/slog"
)

// Logs a message with the request ID from context.
// The variadic args should be key-value pairs for structured logging.
func LogWithRequestID(ctx context.Context, message string, args ...any) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		// Prepend requestID to the args.
		allArgs := make([]any, 0, len(args)+2)
		allArgs = append(allArgs, "requestID", requestID)
		allArgs = append(allArgs, args...)
		slog.Info(message, allArgs...)
	} else {
		slog.Info(message, args...)
	}
}

// Logs an error with the request ID from context.
// The variadic args should be key-value pairs for structured logging.
func LogErrorWithRequestID(ctx context.Context, err error, message string, args ...any) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		// Prepend requestID and error to the args
		allArgs := make([]any, 0, len(args)+4)
		allArgs = append(allArgs, "requestID", requestID, "error", err)
		allArgs = append(allArgs, args...)
		slog.Error(message, allArgs...)
	} else {
		// Prepend error to the args
		allArgs := make([]any, 0, len(args)+2)
		allArgs = append(allArgs, "error", err)
		allArgs = append(allArgs, args...)
		slog.Error(message, allArgs...)
	}
}
