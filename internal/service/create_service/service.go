package create_service

import (
	"errors"
	"net/url"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/model"
	"tiny-bitly/internal/service/create_service/utils"
)

// The URL of our client-facing service, constituting the first part of every
// short URL.
//
// TODO: Move this to environment variable.
var serviceUrlHostname = "https://www.tiny-bitly.com/"

// Creates and saves an alias for the provided long URL, then returns the alias.
func CreateShortURL(dao dao.DAO, originalURL string) (*string, error) {
	// Validate the URL.
	validatedURL, err := utils.ValidateURL(originalURL)
	if err != nil {
		return nil, errors.New("invalid URL")
	}

	// Create a short code based on the database row ID.
	shortCode := utils.GenerateShortCode()

	// Persist the URL entry.
	_, err = dao.URLRecordDAO.Create(model.URLRecord{
		OriginalURL: *validatedURL,
		ShortCode:   shortCode,
	})
	if err != nil {
		return nil, errors.New("failed to save")
	}

	// Build the short URL.
	shortURL, err := url.JoinPath(serviceUrlHostname, shortCode)
	if err != nil {
		return nil, errors.New("invalid URL path segments")
	}

	return &shortURL, nil
}
