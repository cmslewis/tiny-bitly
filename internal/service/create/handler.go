package create

import (
	"encoding/json"
	"log"
	"net/http"
	"tiny-bitly/internal/dao"
)

type CreateUrlRequest struct {
	URL string `json:"url"`

	// A specific user-provided alias to use in the short URL.
	// If not provided, a random short code will be created.
	Alias *string `json:"alias"`
}

type CreateUrlResponse struct {
	ShortURL string `json:"shortUrl"`
}

// Creates an HTTP handler for POST /urls that uses the provided DAO.
// - 201 Created with a CreateUrlResponse on success
// - 400 Bad Request if the original URL is an invalid URL
// - 400 Bad Request if the original URL is longer than 1000 chars
// - 500 System Error if anything else fails
func NewHandlePostURL(dao *dao.DAO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		// Attempt to read the JSON request body.
		request, err := readRequestJson[CreateUrlRequest](r)
		if err != nil {
			http.Error(w, "Malformatted request JSON", http.StatusBadRequest)
			return
		}

		// Log the inbound request.
		log.Printf("Request: URL=%s\n", request.URL)

		// Create the short URL.
		shortURL, err := createShortURL(r.Context(), *dao, request.URL, request.Alias)
		if err != nil {
			log.Println("Internal server error:", err.Error())
			http.Error(w, "Failed to create URL", http.StatusInternalServerError)
			return
		}

		// Send the JSON response.
		w.WriteHeader(http.StatusCreated)
		err = writeResponseJson(w, CreateUrlResponse{
			ShortURL: *shortURL,
		})
		if err != nil {
			http.Error(w, "Failed to create URL", http.StatusInternalServerError)
			return
		}
	}
}

// Reads a JSON object of type T from the provided HTTP request.
// Returns an error if decoding fails.
func readRequestJson[T any](r *http.Request) (*T, error) {
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("Bad request:", err.Error())
		return nil, err
	}
	return &request, nil
}

// Writes a JSON object of type T to the provided HTTP response.
// Returns an error if encoding fails.
func writeResponseJson[T any](w http.ResponseWriter, body T) error {
	w.Header().Set("Content-Type", "application/json")

	// Send the JSON response - or an error.
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		log.Println("Internal server error:", err.Error())
		return err
	}

	return nil
}
