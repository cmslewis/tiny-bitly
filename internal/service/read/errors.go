package read

import (
	"context"
	"net/http"
	"tiny-bitly/internal/apperrors"
)

// Maps service errors to appropriate HTTP status codes and responses. Logs
// detailed error information while returning user-friendly messages.
func handleServiceError(ctx context.Context, w http.ResponseWriter, err error) {
	apperrors.HandleServiceError(ctx, w, err, map[error]apperrors.ErrorMapping{
		apperrors.ErrDataStoreUnavailable: {
			StatusCode:  http.StatusServiceUnavailable,
			UserMessage: "Service temporarily unavailable. Please try again later",
		},
	})
}
