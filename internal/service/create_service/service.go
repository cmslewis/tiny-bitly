package create_service

import (
	"errors"
	"fmt"
	"net/url"
	"tiny-bitly/internal/dao"
	errorspkg "tiny-bitly/internal/errors"
	"tiny-bitly/internal/model"
	"tiny-bitly/internal/service/create_service/utils"
)

// The maximum number of times to try generating a unique short code before
// aborting and returning an error.
var maxTries = 10

// Creates and saves an alias for the provided long URL, then returns the alias.
func CreateShortURL(dao dao.DAO, originalURL string) (*string, error) {
	// Validate the URL.
	validatedURL, err := utils.ValidateURL(originalURL)
	if err != nil {
		return nil, errors.New("invalid URL")
	}

	// Retry until we find a short code not taken yet.
	var shortCode string
	numTries := 0
	for numTries < maxTries {
		numTries += 1

		shortCode = utils.GenerateShortCode()

		// Save a new URL record.
	_, err = dao.URLRecordDAO.Create(model.URLRecord{
		OriginalURL: *validatedURL,
		ShortCode:   shortCode,
	})
		if errorspkg.IsSystemError(err, errorspkg.SystemErrorShortCodeAlreadyInUse) {
			// Try again.
		} else if err != nil {
		return nil, errors.New("failed to save")
	}
	}

	fmt.Printf("Generated a new short code for URL %s: %s", *validatedURL, shortCode)


	// Build the short URL.
	shortURL, err := url.JoinPath(serviceUrlHostname, shortCode)
	if err != nil {
		return nil, errors.New("invalid URL path segments")
	}

	return &shortURL, nil
}
