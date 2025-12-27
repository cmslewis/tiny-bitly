package create

import (
	"errors"
	"net/url"
	"strings"
)

// Returns true if the provided URL alias is a valid base62 string, or false
// otherwise.
func validateAlias(alias string) bool {
	if alias == "" {
		return false
	}

	for i := 0; i < len(alias); i += 1 {
		char := alias[i]
		isValidChar := validateChar(char)
		if !isValidChar {
			return false
		}
	}

	return true
}

// Ensures that the provided URL has a valid URL structure per url.Parse. Adds a
// protocol if one is missing, defaulting to HTTPS. Returns an error if the URL
// is otherwise invalid.
func validateURL(rawURL string) (*string, error) {
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

// Returns true if the provided character is a valid base62 character, or false
// otherwise.
func validateChar(char byte) bool {
	// O(N) is fine for only 62 elements
	for i := 0; i < len(allowedChars); i += 1 {
		allowedChar := allowedChars[i]
		if char == allowedChar {
			return true
		}
	}
	return false
}

func ensureProtocol(url string) string {
	// If the URL already contains a scheme separator, return it as is.
	if strings.Contains(url, "://") {
		return url
	}
	// Prepend the default scheme. Use "https://" as a common default.
	return "https://" + url
}
