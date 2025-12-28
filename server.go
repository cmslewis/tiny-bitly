package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/middleware"
	"tiny-bitly/internal/service/create"
	"tiny-bitly/internal/service/health"
	"tiny-bitly/internal/service/read"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize dependencies.
	appDAO := dao.NewMemoryDAO()

	// Initialize services.
	createService := create.NewService(*appDAO, config)
	readService := read.NewService(*appDAO)
	healthService := health.NewService(*appDAO)

	router := buildRouter(createService, readService, healthService)

	// Middleware: Generate a Request ID for each request (apply first so other
	// handlers can use it).
	handler := middleware.RequestIDMiddleware(router)

	// Configure HTTP server with timeouts.
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.APIPort),
		Handler:      http.TimeoutHandler(handler, config.RequestTimeout, "Request timeout"),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	// Start server in a goroutine so we can handle shutdown signals.
	errChannel := make(chan error, 1)
	go func() {
		log.Printf("Server starting on port %d\n", config.APIPort)
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
		handleQuitSignal(server, sig, config.ShutdownTimeout)
	}
}

// Kills the server when a fatal runtime error occurs.
func handleServerError(err error) {
	log.Fatalf("Server error: %v\n", err)
}

// Attempts to gracefully shut down the server.
func handleQuitSignal(server *http.Server, sig os.Signal, shutdownTimeout time.Duration) {
	log.Printf("Received signal: %v. Shutting down gracefully...\n", sig)

	// Create a context with timeout for graceful shutdown.
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

func buildRouter(createService *create.Service, readService *read.Service, healthService *health.Service) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("GET /health", health.NewGetHealthHandler(healthService))
	mux.HandleFunc("GET /ready", health.NewGetReadyHandler(healthService))

	// Application endpoints
	mux.HandleFunc("POST /urls", create.NewPostURLHandler(createService))
	mux.HandleFunc("GET /{shortCode}", read.NewGetURLHandler(readService))

	return mux
}
