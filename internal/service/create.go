package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// TODO: Set up a real database.
var idCounter int64 = 1

// The URL of our client-facing service, constituting the first part of every
// short URL.
//
// TODO: Move this to environment variable.
var serviceUrlHostname = "https://www.tiny-bitly.com/"

// Short codes consist of 16 distinct characters. We need 8 characters to
// support 1B unique short codes.
var shortCodeLength int32 = 8

// CreateURL creates and saves an alias for the provided long URL, then returns
// the alias.
func CreateURL(originalURL string) (*string, error) {
	// Validate the URL.
	_, err := validateUrl(originalURL)
	if err != nil {
		return nil, errors.New("invalid URL")
	}

	// Create a short code based on the database row ID.
	shortCode, err := generateShortCode(idCounter)
	if err != nil {
		return nil, errors.New("invalid ID")
	}

	// Build the short URL.
	shortURL, err := url.JoinPath(serviceUrlHostname, *shortCode)
	if err != nil {
		return nil, errors.New("invalid URL path segments")
	}

	// Increment the ID counter for next time.
	idCounter += 1

	return &shortURL, nil
}

func validateUrl(rawURL string) (*string, error) {
	urlWithProtocol := ensureProtocol(rawURL)
	_, err := url.ParseRequestURI(urlWithProtocol)
	if err != nil {
		return nil, errors.New("invalid URL")
	}
	return &urlWithProtocol, nil
}

func ensureProtocol(url string) string {
	// If the URL already contains a scheme separator, return it as is.
	if strings.Contains(url, "://") {
		return url
	}
	// Prepend the default scheme. Use "https://" as a common default.
	return "https://" + url
}

func generateShortCode(id int64) (*string, error) {
	if id < 1 {
		return nil, errors.New("invalid ID: must be 1 or greater")
	}
	hexLowercase := fmt.Sprintf("%0*x", shortCodeLength, id)
	return &hexLowercase, nil
}
