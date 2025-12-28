package read

import (
	"context"
	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/middleware"
)

// Service handles URL lookup operations.
type Service struct {
	dao dao.DAO
}

// NewService creates a new read service with the provided dependencies.
func NewService(dao dao.DAO) *Service {
	return &Service{
		dao: dao,
	}
}

// GetOriginalURL gets the original URL from a short code, or returns nil if one does not exist.
func (s *Service) GetOriginalURL(ctx context.Context, shortCode string) (*string, error) {
	// Validate the short code.
	if shortCode == "" {
		return nil, nil
	}

	// Lookup in the data store.
	urlRecord, err := s.dao.URLRecordDAO.GetByShortCode(ctx, shortCode)
	if err != nil {
		middleware.LogWithRequestID(ctx, "Failed to get URL record for short code %s: %v", shortCode, err)
		return nil, apperrors.ErrDataStoreUnavailable
	}

	if urlRecord == nil {
		middleware.LogWithRequestID(ctx, "URL record is nil for short code %s", shortCode)
		return nil, apperrors.ErrShortCodeNotFound
	}

	return &urlRecord.OriginalURL, nil
}
