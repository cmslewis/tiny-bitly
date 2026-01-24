package read

import (
	"net/http"
	"tiny-bitly/internal/middleware"
)

// NewGetURLHandler creates an HTTP handler for GET /{shortCode} that uses the provided service.
// - 302 Temporary Redirect if an original URL is found
// - 400 Bad Request if the short code is empty
// - 404 Not Found if an original URL is not found (or if the short URL is expired)
// - 500 Internal Server Error for other errors
// - 503 Service Unavailable if the data store is unavailable
func NewGetURLHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Parse `shortCode` out of the URL.
		shortCode := r.PathValue("shortCode")
		if shortCode == "" {
			middleware.LogDebugWithRequestID(r.Context(), "Bad request: empty short code")
			http.Error(w, "Short code must be non-empty", http.StatusBadRequest)
			return
		}

		// Log the inbound request.
		middleware.LogDebugWithRequestID(r.Context(), "Resolving short URL with code", "shortCode", shortCode)

		// Get the original URL for this short code.
		originalURL, err := service.GetOriginalURL(r.Context(), shortCode)
		if err != nil {
			handleServiceError(r.Context(), w, err)
			return
		}

		// Return 404 if original URL not found.
		if originalURL == nil {
			middleware.LogDebugWithRequestID(r.Context(), "No URL found for short code", "shortCode", shortCode)
			// Cache 404s for a short time to reduce load on invalid codes.
			w.Header().Set("Cache-Control", "public, max-age=60")
			http.Error(w, "No URL found for short code", http.StatusNotFound)
			return
		}

		// Set cache headers for CDN caching (302 redirects are cacheable).
		// Cache for 24 hours - short codes rarely change, and expired codes are filtered by DB query.
		w.Header().Set("Cache-Control", "public, max-age=86400, s-maxage=86400")
		w.Header().Set("Vary", "Accept-Encoding")

		// 302 Temporary Redirect to the original URL.
		http.Redirect(w, r, *originalURL, http.StatusFound)
	}
}
