package apperrors

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"tiny-bitly/internal/middleware"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ErrorMapping struct {
	StatusCode  int
	UserMessage string
}

// Maps service errors to appropriate HTTP status codes and responses. Logs
// detailed error information while returning user-friendly messages.
func HandleServiceError(ctx context.Context, w http.ResponseWriter, err error, mappings map[error]ErrorMapping) {
	// Log the detailed error for debugging with request ID.
	middleware.LogErrorWithRequestID(ctx, err, "Service error")

	w.Header().Set("Content-Type", "application/json")

	found := false
	for entryError, entry := range mappings {
		if errors.Is(err, entryError) {
			writeResponse(w, entry.StatusCode, entry.UserMessage)
			found = true
		}
	}
	if !found {
		writeResponse(w,
			http.StatusInternalServerError,
			"An unexpected error occurred",
		)
	}
}

func writeResponse(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: errorMessage,
	})
}
