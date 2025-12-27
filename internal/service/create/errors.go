package create

import (
	"net/http"
	"tiny-bitly/internal/apperrors"
)

// Maps service errors to appropriate HTTP status codes and responses. Logs
// detailed error information while returning user-friendly messages.
func handleServiceError(w http.ResponseWriter, err error) {
	apperrors.HandleServiceError(w, err, map[error]apperrors.ErrorMapping{
		apperrors.ErrInvalidURL: {
			StatusCode:  http.StatusBadRequest,
			UserMessage: "Invalid URL format",
		},
		apperrors.ErrURLLengthExceeded: {
			StatusCode:  http.StatusBadRequest,
			UserMessage: "URL exceeds maximum length",
		},
		apperrors.ErrInvalidAlias: {
			StatusCode:  http.StatusBadRequest,
			UserMessage: "Invalid alias format. Alias must contain only letters, numbers, and be non-empty",
		},
		apperrors.ErrAliasAlreadyInUse: {
			StatusCode:  http.StatusConflict,
			UserMessage: "Alias is already in use",
		},
		apperrors.ErrMaxRetriesExceeded: {
			StatusCode:  http.StatusInternalServerError,
			UserMessage: "Unable to generate unique short code. Please try again",
		},
		apperrors.ErrConfigurationMissing: {
			StatusCode:  http.StatusInternalServerError,
			UserMessage: "Service configuration error",
		},
		apperrors.ErrDataStoreUnavailable: {
			StatusCode:  http.StatusServiceUnavailable,
			UserMessage: "Service temporarily unavailable. Please try again later",
		},
	})
}
