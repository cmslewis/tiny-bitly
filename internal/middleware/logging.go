package middleware

import (
	"context"
	"log/slog"
)

// LogDebugWithRequestID logs a debug message with the request ID from context.
// Use this for verbose debugging information that's useful during development
// but may be too noisy for production.
// The variadic args should be key-value pairs for structured logging.
func LogDebugWithRequestID(ctx context.Context, message string, args ...any) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		// Prepend requestID to the args.
		allArgs := make([]any, 0, len(args)+2)
		allArgs = append(allArgs, "requestID", requestID)
		allArgs = append(allArgs, args...)
		slog.Debug(message, allArgs...)
	} else {
		slog.Debug(message, args...)
	}
}

// LogWithRequestID logs an info message with the request ID from context.
// Use this for normal operational events that are useful to track in production.
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

// LogErrorWithRequestID logs an error with the request ID from context.
// Use this for error conditions that need attention.
// The variadic args should be key-value pairs for structured logging.
func LogErrorWithRequestID(ctx context.Context, err error, message string, args ...any) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		// Prepend requestID and error to the args.
		allArgs := make([]any, 0, len(args)+4)
		allArgs = append(allArgs, "requestID", requestID, "error", err)
		allArgs = append(allArgs, args...)
		slog.Error(message, allArgs...)
	} else {
		// Prepend error to the args.
		allArgs := make([]any, 0, len(args)+2)
		allArgs = append(allArgs, "error", err)
		allArgs = append(allArgs, args...)
		slog.Error(message, allArgs...)
	}
}
