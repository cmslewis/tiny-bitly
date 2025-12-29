package read

import (
	"context"
	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/middleware"
)

// Service handles URL lookup operations.
type Service struct {
	dao    dao.DAO
	config *config.Config
}

// NewService creates a new read service with the provided dependencies.
func NewService(dao dao.DAO, config *config.Config) *Service {
	return &Service{
		dao:    dao,
		config: config,
	}
}

// GetOriginalURL gets the original URL from a short code, or returns nil if one does not exist.
func (s *Service) GetOriginalURL(ctx context.Context, shortCode string) (*string, error) {
	// Validate the short code.
	err := validateShortCode(shortCode, s.config.MaxAliasLength)
	if err != nil {
		return nil, err
	}

	// Lookup in the data store.
	urlRecord, err := s.dao.URLRecordDAO.GetByShortCode(ctx, shortCode)
	if err != nil {
		middleware.LogErrorWithRequestID(ctx, err, "Failed to get URL record for short code", "shortCode", shortCode)
		return nil, apperrors.ErrDataStoreUnavailable
	}

	if urlRecord == nil {
		middleware.LogDebugWithRequestID(ctx, "URL record is nil for short code", "shortCode", shortCode)
		return nil, apperrors.ErrShortCodeNotFound
	}

	return &urlRecord.OriginalURL, nil
}

func validateShortCode(shortCode string, maxLength int) error {
	if shortCode == "" {
		return apperrors.ErrShortCodeNotFound
	}
	if len(shortCode) > maxLength {
		return apperrors.ErrShortCodeNotFound
	}
	return nil
}
