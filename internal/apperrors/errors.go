package apperrors

import "errors"

// Sentinel errors for system-level issues.
// These can be checked using errors.Is(err, ErrShortCodeAlreadyInUse).
var (
	// Returned when attempting to create a URL record with a short code that is
	// already in use by an active (not deleted and not expired) entity.
	ErrShortCodeAlreadyInUse = errors.New("short code already in use")
)
