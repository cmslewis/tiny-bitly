package apperrors

import "errors"

// Sentinel errors for system-level issues.
// These can be checked using errors.Is(err, ErrShortCodeAlreadyInUse).
var (
	// Returned when a custom alias is already in use.
	ErrAliasAlreadyInUse = errors.New("alias already in use")

	// Returned when a required configuration is missing.
	ErrConfigurationMissing = errors.New("configuration missing")

	// Returned when the data store is not accessible.
	ErrDataStoreUnavailable = errors.New("data store unavailable")

	// Returned when the provided alias is invalid.
	ErrInvalidAlias = errors.New("invalid alias")

	// Returned when the provided URL is invalid.
	ErrInvalidURL = errors.New("invalid URL")

	// Returned when unable to generate a unique short code after maximum
	// retries.
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")

	// Returned when attempting to create a URL record with a short code that is
	// already in use by an active (not deleted and not expired) entity.
	ErrShortCodeAlreadyInUse = errors.New("short code already in use")

	// Returned when attempting to get a short code that does not exist.
	ErrShortCodeNotFound = errors.New("short code not found")

	// Returned when the URL exceeds the maximum allowed length.
	ErrURLLengthExceeded = errors.New("URL length exceeded")
)
