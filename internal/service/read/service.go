package read

import (
	"context"
	"log"
	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/dao"
)

// Gets the original URL from a short code, or returns nil if one does not exist.
func getOriginalURL(ctx context.Context, dao dao.DAO, shortCode string) (*string, error) {
	// Validate the short code.
	if shortCode == "" {
		return nil, nil
	}

	// Lookup in the data store.
	urlRecord, err := dao.URLRecordDAO.GetByShortCode(ctx, shortCode)
	if err != nil {
		log.Printf("Failed to get URL record for short code %s: %v", shortCode, err)
		return nil, apperrors.ErrDataStoreUnavailable
	}

	if urlRecord == nil {
		log.Printf("URL record is nil for short code %s", shortCode)
		return nil, nil
	}

	return &urlRecord.OriginalURL, nil
}
