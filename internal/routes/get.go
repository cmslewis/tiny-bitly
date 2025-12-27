package routes

import (
	"log"
	"net/http"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/service/read_service"
)

// Handles a GET request for an original URL based on its short URL.
// - 302 Temporary Redirect if an original URL is found
// - 400 Not Found if an original URL is not found (or if the short URL is expired)
// - 500 System Error if any other error occurs
func HandleGetURL(w http.ResponseWriter, r *http.Request) {
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

	// Create a DAO.
	dao := dao.GetDAOOfType(dao.DAOTypeMemory)
	if dao == nil {
		log.Print("Internal server error: failed to get DAO")
		http.Error(w, "Failed to get URL", http.StatusInternalServerError)
		return
	}

	// Get the original URL for this short code.
	originalURL, err := read_service.GetOriginalURL(*dao, shortCode)
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
