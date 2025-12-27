package main

import (
	"log"
	"net/http"
	"os"

	"tiny-bitly/internal/routes"

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

	mux.HandleFunc("POST /urls", routes.HandlePostURL)
	mux.HandleFunc("GET /{shortCode}", routes.HandleGetURL)

	return mux
}
