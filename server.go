package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	// Start server in a goroutine so we can handle shutdown signals.
	errChannel := make(chan error, 1)
	go func() {
		log.Printf("Server starting on port %d\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChannel <- err
		}
	}()

	// Wait for interrupt signal or server error.
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChannel:
		handleServerError(err)
	case sig := <-quitChannel:
		handleQuitSignal(server, sig)
	}
}

// Kills the server when a fatal runtime error occurs.
func handleServerError(err error) {
	log.Fatalf("Server error: %v\n", err)
}

// Attempts to gracefully shut down the server.
func handleQuitSignal(server *http.Server, sig os.Signal) {
	log.Printf("Received signal: %v. Shutting down gracefully...\n", sig)

	// Create a context with timeout for graceful shutdown.
	shutdownTimeout := config.GetDurationEnvOrDefault("TIMEOUT_SHUTDOWN_MILLIS", 30000)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown.
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
		// Force close if graceful shutdown fails.
		if closeErr := server.Close(); closeErr != nil {
			log.Fatalf("Error forcing server close: %v\n", closeErr)
		}
		log.Fatal("Server forced to close")
	}

	log.Println("Server shutdown complete")
}

func buildRouter(appDAO *dao.DAO) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /urls", create.NewHandlePostURL(appDAO))
	mux.HandleFunc("GET /{shortCode}", read.NewHandleGetURL(appDAO))

	return mux
}
