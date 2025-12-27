package main

import (
	"fmt"
	"log"
	"net/http"

	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/service/create"
	"tiny-bitly/internal/service/read"

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

	port := config.GetIntEnvOrDefault("API_PORT", 8080)
	idleTimeout := config.GetDurationEnvOrDefault("TIMEOUT_IDLE_MILLIS", 60000)
	requestTimeout := config.GetDurationEnvOrDefault("TIMEOUT_REQUEST_MILLIS", 30000)
	readTimeout := config.GetDurationEnvOrDefault("TIMEOUT_READ_MILLIS", 30000)
	writeTimeout := config.GetDurationEnvOrDefault("TIMEOUT_WRITE_MILLIS", 30000)

	// Configure HTTP server with timeouts.
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      http.TimeoutHandler(router, requestTimeout, "Request timeout"),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Start the HTTP server.
	log.Printf("Server starting on port %d\n", port)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func buildRouter(appDAO *dao.DAO) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /urls", create.NewHandlePostURL(appDAO))
	mux.HandleFunc("GET /{shortCode}", read.NewHandleGetURL(appDAO))

	return mux
}
