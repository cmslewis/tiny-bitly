package utils

import (
	"errors"
	"net/url"
	"strings"
)

// Ensures that the provided URL has a valid URL structure per url.Parse. Adds a
// protocol if one is missing, defaulting to HTTPS. Returns an error if the URL
// is otherwise invalid.
func ValidateURL(rawURL string) (*string, error) {
	rawURLWithProtocol := ensureProtocol(rawURL)
	parsedURL, err := url.Parse(rawURLWithProtocol)
	if err != nil {
		return nil, errors.New("invalid URL")
	}
	if parsedURL.Host == "" {
		return nil, errors.New("invalid URL")
	}
	return &rawURLWithProtocol, nil
}

func ensureProtocol(url string) string {
	// If the URL already contains a scheme separator, return it as is.
	if strings.Contains(url, "://") {
		return url
	}
	// Prepend the default scheme. Use "https://" as a common default.
	return "https://" + url
}
