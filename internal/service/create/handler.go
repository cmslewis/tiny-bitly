package create

import (
	"encoding/json"
	"net/http"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/middleware"
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
// - 400 Bad Request if the URL is invalid, exceeds length, or alias is invalid
// - 409 Conflict if the alias is already in use
// - 500 Internal Server Error for other errors
// - 503 Service Unavailable if the data store is unavailable
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
		middleware.LogWithRequestID(r.Context(), "Request: URL=%s", request.URL)

		// Create the short URL.
		shortURL, err := CreateShortURL(r.Context(), *dao, request.URL, request.Alias)
		if err != nil {
			handleServiceError(r.Context(), w, err)
			return
		}

		// Send the JSON response.
		w.WriteHeader(http.StatusCreated)
		err = writeResponseJson(w, CreateUrlResponse{
			ShortURL: *shortURL,
		})
		if err != nil {
			handleServiceError(r.Context(), w, err)
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
		middleware.LogErrorWithRequestID(r.Context(), err, "Bad request")
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
		// Note: We don't have request context here, so use regular log
		// In practice, this error is rare and would be caught by the handler
		return err
	}

	return nil
}
