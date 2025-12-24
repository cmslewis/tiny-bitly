package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"tiny-bitly/internal/service"
)

func main() {
	// Register a handler function for the default route ("/").

	router := buildRouter()

	// Start the HTTP server.
	port := ":8080"
	log.Printf("Server starting on port %s\n", port)
	err := http.ListenAndServe(port, router)
	if err != nil {
		log.Fatal(err)
	}
}

func buildRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/urls", handleCreateURL)

	return router
}

type CreateUrlRequest struct {
	URL string `json:"url"`
}

type CreateUrlResponse struct {
	ShortURL string `json:"shortUrl"`
}

// handleCreateURL
func handleCreateURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Verify this is a POST request.
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Attempt to read the JSON request body.
	request, err := readRequest[CreateUrlRequest](r)
	if err != nil {
		http.Error(w, "Malformatted request JSON", http.StatusBadRequest)
		return
	}

	// Log the inbound request.
	fmt.Printf("Request: URL=%s\n", request.URL)

	// Create the short URL.
	shortURL, err := service.CreateURL(request.URL)
	if err != nil {
		fmt.Println("[Create URL] Internal server error:", err.Error())
		http.Error(w, "Failed to create URL", http.StatusInternalServerError)
		return
	}

	// Send the JSON response.
	err = writeResponse(w, CreateUrlResponse{
		ShortURL: *shortURL,
	})
	if err != nil {
		http.Error(w, "Failed to create URL", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("URL created successfully"))
}

func readRequest[T any](r *http.Request) (*T, error) {
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("[Create URL] Bad request:", err.Error())
		return nil, err
	}
	return &request, nil
}

func writeResponse[T any](w http.ResponseWriter, body T) error {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	// Send the JSON response - or an error.
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		fmt.Println("[Create URL] Internal server error:", err.Error())
		return err
	}

	return nil
}
