package read

import (
	"net/http"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/middleware"
)

// NewHandleGetURL creates an HTTP handler for GET /{shortCode} that uses the provided DAO.
// - 302 Temporary Redirect if an original URL is found
// - 400 Bad Request if the short code is empty
// - 404 Not Found if an original URL is not found (or if the short URL is expired)
// - 500 Internal Server Error for other errors
// - 503 Service Unavailable if the data store is unavailable
func NewHandleGetURL(dao *dao.DAO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		// Parse `shortCode` out of the URL.
		shortCode := r.PathValue("shortCode")
		if shortCode == "" {
			middleware.LogWithRequestID(r.Context(), "Bad request: empty short code")
			http.Error(w, "Short code must be non-empty", http.StatusBadRequest)
			return
		}

		// Log the inbound request.
		middleware.LogWithRequestID(r.Context(), "Resolving short URL with code: %s", shortCode)

		// Get the original URL for this short code.
		originalURL, err := GetOriginalURL(r.Context(), *dao, shortCode)
		if err != nil {
			handleServiceError(r.Context(), w, err)
			return
		}

		// Return 404 if original URL not found.
		if originalURL == nil {
			middleware.LogWithRequestID(r.Context(), "No URL found for short code '%s'", shortCode)
			http.Error(w, "No URL found for short code", http.StatusNotFound)
			return
		}

		// 302 Temporary Redirect to the original URL.
		http.Redirect(w, r, *originalURL, http.StatusFound)
	}
}
