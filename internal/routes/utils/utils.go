package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// Reads a JSON object of type T from the provided HTTP request.
// Returns an error if decoding fails.
func ReadRequestJson[T any](r *http.Request) (*T, error) {
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
func WriteResponseJson[T any](w http.ResponseWriter, body T) error {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	// Send the JSON response - or an error.
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		log.Println("Internal server error:", err.Error())
		return err
	}

	return nil
}
