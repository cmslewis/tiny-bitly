package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"time"

	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/model"

	"github.com/joho/godotenv"
)

const (
	numRecords = 900000
	baseURL    = "https://example.com"
)

func main() {
	// Load environment variables from .env file in development only
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			slog.Warn("No .env file found, using environment variables", "error", err)
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize logging
	initLogging(cfg)

	// Initialize database DAO
	appDAO, err := dao.NewDatabaseDAO(
		cfg.PostgresPort,
		cfg.PostgresDB,
		cfg.PostgresUser,
		cfg.PostgresPassword,
	)
	if err != nil {
		slog.Error("Failed to initialize database DAO", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	slog.Info("Starting to load test data", "count", numRecords)

	successCount := 0
	failureCount := 0

	for i := 0; i < numRecords; i++ {
		// Generate a unique short code
		shortCode := generateShortCode(cfg.ShortCodeLength)

		// Generate a unique URL with a path
		originalURL := fmt.Sprintf("%s/page/%d?test=%d", baseURL, i, rand.IntN(10000))

		// Set expiration time (some records expire soon, some later)
		expiresAt := time.Now().Add(cfg.ShortCodeTTL)
		if i%10 == 0 {
			// Every 10th record expires in 1 hour (for testing expiration)
			expiresAt = time.Now().Add(1 * time.Hour)
		}

		urlRecord := model.URLRecord{
			OriginalURL: originalURL,
			ShortCode:   shortCode,
			ExpiresAt:   expiresAt,
		}

		_, err := appDAO.URLRecordDAO.Create(ctx, urlRecord)
		if err != nil {
			// If short code conflict, try again with a new code
			if errors.Is(err, apperrors.ErrShortCodeAlreadyInUse) {
				i-- // Retry this iteration
				failureCount++
				continue
			}
			slog.Error(
				"Failed to create URL record",
				"error", err,
				"index", i,
				"shortCode", shortCode,
			)
			failureCount++
			continue
		}

		successCount++
		if (i+1)%10 == 0 {
			slog.Info("Progress", "created", successCount, "failed", failureCount, "total", i+1)
		}
	}

	slog.Info("Finished loading test data",
		"success", successCount,
		"failures", failureCount,
		"total", numRecords,
	)
}

func initLogging(cfg *config.Config) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}
	handler = slog.NewTextHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}

// generateShortCode generates a random short code using base62 characters (A-Z, a-z, 0-9).
func generateShortCode(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.IntN(len(chars))]
	}
	return string(b)
}
