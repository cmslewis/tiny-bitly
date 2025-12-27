package create_service

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"
	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	errorspkg "tiny-bitly/internal/errors"
	"tiny-bitly/internal/model"
	"tiny-bitly/internal/service/create_service/utils"
)

// Creates and saves an alias for the provided long URL, then returns the alias.
func CreateShortURL(dao dao.DAO, originalURL string, alias *string) (*string, error) {
	// Validate the URL.
	validatedURL, err := utils.ValidateURL(originalURL)
	if err != nil {
		return nil, errors.New("invalid URL")
	}

	// Read environment variables.
	maxTries := config.GetIntEnvOrDefault("MAX_TRIES_CREATE_SHORT_CODE", 10)
	shortCodeLength := config.GetIntEnvOrDefault("SHORT_CODE_LENGTH", 6)
	shortCodeTtlSeconds := config.GetIntEnvOrDefault("SHORT_CODE_TTL_SECONDS", 30)

	// If a custom alias was provided, validate it.
	if alias != nil && !utils.ValidateAlias(*alias) {
		return nil, errors.New("invalid alias, must be non-empty base62 string (A-Z, a-z, 0-9)")
	}

	// Retry until we find a short code not taken yet.
	var shortCode string
	numTries := 0
	for numTries < maxTries {
		numTries += 1

		if alias != nil {
			shortCode = *alias
		} else {
			shortCode = utils.GenerateShortCode(shortCodeLength)
		}

		// Set expiration time based on configured TTL.
		expiresAt := time.Now().Add(time.Duration(shortCodeTtlSeconds) * time.Second)

		log.Printf("Creating a new URL record shortCode=%s expiresAt=%v", shortCode, expiresAt)

		// Save a new URL record.
		_, err = dao.URLRecordDAO.Create(model.URLRecord{
			OriginalURL: *validatedURL,
			ShortCode:   shortCode,
			ExpiresAt:   expiresAt,
		})

		// If the short code is already in use:
		if errorspkg.IsSystemError(err, errorspkg.SystemErrorShortCodeAlreadyInUse) {
			// If we're using a custom alias, fail outright.
			if alias != nil {
				return nil, errors.New("custom alias already in use")
			}
			// Else, try again with a new randomly generated short code.
			continue
		}

		// Fail if another error occurred
		if err != nil {
			return nil, errors.New("failed to save")
		}

		// Break on success.
		break
	}

	fmt.Printf("Generated a new short code for URL %s: %s", *validatedURL, shortCode)

	// Get the URL of our client-facing service.
	hostname, isDefined := os.LookupEnv("API_HOSTNAME")
	if !isDefined {
		log.Fatal("environment variable API_HOSTNAME is not defined")
	}

	// Build the short URL using the short code.
	shortURL, err := url.JoinPath(hostname, shortCode)
	if err != nil {
		return nil, errors.New("invalid URL path segments")
	}

	return &shortURL, nil
}
