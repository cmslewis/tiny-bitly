package create

import (
	"encoding/json"
	"net/http"
	"tiny-bitly/internal/middleware"
)

type CreateURLRequest struct {
	URL string `json:"url"`

	// A specific user-provided alias to use in the short URL.
	// If not provided, a random short code will be created.
	Alias *string `json:"alias"`
}

type CreateURLResponse struct {
	ShortURL string `json:"shortUrl"`
}

// NewPostURLHandler creates an HTTP handler for POST /urls that uses the provided service.
// - 201 Created with a CreateUrlResponse on success
// - 400 Bad Request if the URL is invalid, exceeds length, or alias is invalid
// - 409 Conflict if the alias is already in use
// - 500 Internal Server Error for other errors
// - 503 Service Unavailable if the data store is unavailable
func NewPostURLHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Attempt to read the JSON request body.
		request, err := readRequestJson[CreateURLRequest](r)
		if err != nil {
			http.Error(w, "Malformatted request JSON", http.StatusBadRequest)
			return
		}

		// Log the inbound request.
		middleware.LogWithRequestID(r.Context(), "Request received", "requestURL", request.URL)

		// Create the short URL.
		shortURL, err := service.CreateShortURL(r.Context(), request.URL, request.Alias)
		if err != nil {
			handleServiceError(r.Context(), w, err)
			return
		}

		// Send the JSON response.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = writeResponseJson(w, CreateURLResponse{
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
// Note: Content-Type header should be set before calling this function.
func writeResponseJson[T any](w http.ResponseWriter, body T) error {
	// Send the JSON response - or an error.
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		// Note: We don't have request context here, so use regular log
		// In practice, this error is rare and would be caught by the handler
		return err
	}

	return nil
}
