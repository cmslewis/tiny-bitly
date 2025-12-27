package main

import (
	"log"
	"net/http"
	"os"

	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/dao/daotypes"
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

	// Initialize dependencies.
	appDAO := dao.GetDAOOfType(dao.DAOTypeMemory)
	if appDAO == nil {
		log.Fatal("Failed to initialize DAO")
	}

	router := buildRouter(appDAO)

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

func buildRouter(appDAO *daotypes.DAO) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /urls", create_service.NewHandlePostURL(appDAO))
	mux.HandleFunc("GET /{shortCode}", read_service.NewHandleGetURL(appDAO))

	return mux
}
