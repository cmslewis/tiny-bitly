package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/service/create_service"
	"tiny-bitly/internal/service/read_service"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := buildRouter()

	// Lookup the PORT to use.
	port, isDefined := os.LookupEnv("API_PORT")
	if !isDefined {
		log.Fatal("environment variable API_PORT is not defined")
	}

	// Start the HTTP server.
	log.Printf("Server starting on port %s\n", port)
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatal(err)
	}
}

func buildRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /urls", handlePostURL)
	mux.HandleFunc("GET /{shortCode}", handleGetURL)

	return mux
}

type CreateUrlRequest struct {
	URL string `json:"url"`
}

type CreateUrlResponse struct {
	ShortURL string `json:"shortUrl"`
}

func handlePostURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Attempt to read the JSON request body.
	request, err := readRequestJson[CreateUrlRequest](r)
	if err != nil {
		http.Error(w, "Malformatted request JSON", http.StatusBadRequest)
		return
	}

	maxURLLength := config.GetIntEnvOrDefault("MAX_URL_LENGTH", 6)

	if len(request.URL) > maxURLLength {
		log.Print("Bad request: empty short code")
		http.Error(w, "URL must be no longer than 1000 chars", http.StatusBadRequest)
		return
	}

	// Log the inbound request.
	log.Printf("Request: URL=%s\n", request.URL)

	// Create a DAO.
	dao := dao.GetDAOOfType(dao.DAOTypeMemory)
	if dao == nil {
		log.Println("Internal server error: failed to get DAO")
		http.Error(w, "Failed to create URL", http.StatusInternalServerError)
		return
	}

	// Create the short URL.
	shortURL, err := create_service.CreateShortURL(*dao, request.URL)
	if err != nil {
		log.Println("Internal server error:", err.Error())
		http.Error(w, "Failed to create URL", http.StatusInternalServerError)
		return
	}

	// Send the JSON response.
	err = writeResponseJson(w, CreateUrlResponse{
		ShortURL: *shortURL,
	})
	if err != nil {
		http.Error(w, "Failed to create URL", http.StatusInternalServerError)
		return
	}
}

func handleGetURL(w http.ResponseWriter, r *http.Request) {
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

func readRequestJson[T any](r *http.Request) (*T, error) {
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("Bad request:", err.Error())
		return nil, err
	}
	return &request, nil
}

func writeResponseJson[T any](w http.ResponseWriter, body T) error {
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
