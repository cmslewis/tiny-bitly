package main

import (
	"fmt"
	"log"
	"net/http"

	"tiny-bitly/internal/config"
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
	log.Printf("Server starting on port %s\n", port)
	err = server.ListenAndServe()
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
