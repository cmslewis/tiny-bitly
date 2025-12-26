package read_service

import (
	"errors"
	"log"
	"tiny-bitly/internal/dao"
)

// Gets the original URL from a short code, or returns nil if one does not exist.
func GetOriginalURL(dao dao.DAO, shortCode string) (*string, error) {
	// Validate the short code.
	if shortCode == "" {
		return nil, nil
	}

	// Lookup in the data store.
	urlRecord, err := dao.URLRecordDAO.GetByShortCode(shortCode)
	if err != nil {
		log.Print(err)
		return nil, errors.New("failed to get original URL by short code")
	}

	if urlRecord == nil {
		log.Printf("URL record is nil for short code %s", shortCode)
		return nil, nil
	}

	return &urlRecord.OriginalURL, nil
}
