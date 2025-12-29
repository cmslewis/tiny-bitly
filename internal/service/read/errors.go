package read

import (
	"context"
	"net/http"
	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/service"
)

// Maps service errors to appropriate HTTP status codes and responses. Logs
// detailed error information while returning user-friendly messages.
func handleServiceError(ctx context.Context, w http.ResponseWriter, err error) {
	service.HandleServiceError(ctx, w, err, map[error]service.ErrorMapping{
		apperrors.ErrDataStoreUnavailable: {
			StatusCode:  http.StatusServiceUnavailable,
			UserMessage: "Service temporarily unavailable. Please try again later",
		},
		apperrors.ErrShortCodeNotFound: {
			StatusCode:  http.StatusNotFound,
			UserMessage: "Short code does not exist",
		},
	})
}
