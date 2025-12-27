package read

import (
	"log"
	"net/http"
	"tiny-bitly/internal/dao"
)

// NewHandleGetURL creates an HTTP handler for GET /{shortCode} that uses the provided DAO.
// - 302 Temporary Redirect if an original URL is found
// - 400 Not Found if an original URL is not found (or if the short URL is expired)
// - 500 System Error if any other error occurs
func NewHandleGetURL(dao *dao.DAO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		// Parse `shortCode` out of the URL.
		shortCode := r.PathValue("shortCode")
		if shortCode == "" {
			log.Print("Bad request: empty short code")
			http.Error(w, "Short code must be non-empty", http.StatusBadRequest)
			return
		}

		// Log the inbound request.
		log.Printf("Resolving short URL with code: %s\n", shortCode)

		// Get the original URL for this short code.
		originalURL, err := getOriginalURL(r.Context(), *dao, shortCode)
		if err != nil {
			log.Print("Internal server error: failed to lookup original URL")
			http.Error(w, "Failed to get URL", http.StatusInternalServerError)
			return
		}

		// Return 404 if original URL not found.
		if originalURL == nil {
			log.Printf("No URL found for short code '%s'", shortCode)
			http.Error(w, "No URL found for short code", http.StatusNotFound)
			return
		}

		// 302 Temporary Redirect to the original URL.
		http.Redirect(w, r, *originalURL, http.StatusFound)
	}
}
