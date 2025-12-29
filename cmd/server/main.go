package main

import (
	"context"
	"fmt"
	"log/slog"
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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load environment variables.
	err := godotenv.Load()
	if err != nil {
		logFatal("Failed to load .env file", "error", err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		logFatal("Failed to load configuration", "error", err)
	}

	initLogging(cfg)

	// Initialize services.
	appDAO := dao.NewMemoryDAO()
	createService := create.NewService(*appDAO, cfg)
	readService := read.NewService(*appDAO, cfg)
	healthService := health.NewService(*appDAO)

	router := buildRouter(createService, readService, healthService)
	handler := middleware.RequestIDMiddleware(router)
	handler = middleware.RateLimitMiddleware(handler, cfg.RateLimitRequestsPerSecond, cfg.RateLimitBurst)
	handler = middleware.MetricsMiddleware(handler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.APIPort),
		Handler:      http.TimeoutHandler(handler, cfg.RequestTimeout, "Request timeout"),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Start server in a goroutine so we can handle shutdown signals.
	errChannel := make(chan error, 1)
	go func() {
		slog.Info("Server starting", "port", cfg.APIPort)
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
		handleQuitSignal(server, sig, cfg.ShutdownTimeout)
	}
}

// Kills the server when a fatal runtime error occurs.
func handleServerError(err error) {
	logFatal("Server error", "error", err)
}

// Attempts to gracefully shut down the server.
func handleQuitSignal(server *http.Server, sig os.Signal, shutdownTimeout time.Duration) {
	slog.Info("Received quit signal. Shutting down gracefully...", "signal", sig)

	// Create a context with timeout for graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown.
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error during server shutdown", "error", err)
		// Force close if graceful shutdown fails.
		if closeErr := server.Close(); closeErr != nil {
			logFatal("Error forcing server close", "error", closeErr)
		}
		logFatal("Server forced to close")
	}

	slog.Info("Server shutdown complete")
}

func initLogging(cfg *config.Config) {
	// Emit structured logs as JSON at the configured log level.
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})
	slog.SetDefault(slog.New(logHandler))
}

func buildRouter(createService *create.Service, readService *read.Service, healthService *health.Service) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("GET /health", health.NewGetHealthHandler(healthService))
	mux.HandleFunc("GET /ready", health.NewGetReadyHandler(healthService))

	// Metrics endpoints
	mux.Handle("GET /metrics", promhttp.Handler())

	// Application endpoints
	mux.HandleFunc("POST /urls", create.NewPostURLHandler(createService))
	mux.HandleFunc("GET /{shortCode}", read.NewGetURLHandler(readService))

	return mux
}

// Logs an error using structured logging and exits the program with code 1.
func logFatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
