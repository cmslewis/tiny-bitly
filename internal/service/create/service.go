package create

import (
	"context"
	"errors"
	"net/url"
	"time"
	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/middleware"
	"tiny-bitly/internal/model"
)

// Service handles URL shortening operations.
type Service struct {
	dao    dao.DAO
	config *config.Config
}

// NewService creates a new create service with the provided dependencies.
func NewService(dao dao.DAO, config *config.Config) *Service {
	return &Service{
		dao:    dao,
		config: config,
	}
}

// CreateShortURL creates and saves an alias for the provided long URL, then returns the alias.
func (s *Service) CreateShortURL(
	ctx context.Context,
	originalURL string,
	alias *string,
) (*string, error) {
	// Validate the URL.
	validatedURL, err := validateURL(originalURL)
	if err != nil {
		return nil, apperrors.ErrInvalidURL
	}

	hostname := s.config.APIHostname
	if hostname == "" {
		return nil, apperrors.ErrConfigurationMissing
	}
	maxTries := s.config.MaxTriesCreateShortCode
	maxURLLength := s.config.MaxURLLength
	shortCodeLength := s.config.ShortCodeLength
	shortCodeTTL := s.config.ShortCodeTTL

	if len(originalURL) > maxURLLength {
		return nil, apperrors.ErrURLLengthExceeded
	}

	// If a custom alias was provided, validate it.
	if alias != nil && !validateAlias(*alias) {
		return nil, apperrors.ErrInvalidAlias
	}

	// Retry until we find a short code not taken yet.
	var shortCode string
	var hasCreated bool
	numTries := 0
	for numTries < maxTries {
		numTries += 1

		if alias != nil {
			shortCode = *alias
		} else {
			shortCode = generateShortCode(shortCodeLength)
		}

		// Set expiration time based on configured TTL.
		expiresAt := time.Now().Add(shortCodeTTL)

		middleware.LogWithRequestID(ctx, "Creating a new URL record shortCode=%s expiresAt=%v", shortCode, expiresAt)

		// Save a new URL record.
		_, err = s.dao.URLRecordDAO.Create(ctx, model.URLRecord{
			OriginalURL: *validatedURL,
			ShortCode:   shortCode,
			ExpiresAt:   expiresAt,
		})

		// If the short code is already in use:
		if errors.Is(err, apperrors.ErrShortCodeAlreadyInUse) {
			// If we're using a custom alias, fail outright.
			if alias != nil {
				return nil, apperrors.ErrAliasAlreadyInUse
			}
			// Else, try again with a new randomly generated short code.
			continue
		}

		// Fail if another error occurred
		if err != nil {
			middleware.LogErrorWithRequestID(ctx, err, "Failed to save URL record")
			return nil, apperrors.ErrDataStoreUnavailable
		}

		// Break on success.
		hasCreated = true
		break
	}

	// Check if we exceeded max retries without success.
	if !hasCreated {
		return nil, apperrors.ErrMaxRetriesExceeded
	}

	middleware.LogWithRequestID(ctx, "Generated a new short code for URL %s: %s", *validatedURL, shortCode)

	// Build the short URL using the short code.
	shortURL, err := url.JoinPath(hostname, shortCode)
	if err != nil {
		middleware.LogErrorWithRequestID(ctx, err, "Failed to build short URL")
		return nil, apperrors.ErrConfigurationMissing
	}

	return &shortURL, nil
}
