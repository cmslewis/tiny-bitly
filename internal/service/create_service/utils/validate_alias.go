package utils

// Returns true if the provided URL alias is a valid base62 string, or false
// otherwise.
func ValidateAlias(alias string) bool {
	if alias == "" {
		return false
	}

	for i := 0; i < len(alias); i += 1 {
		char := alias[i]
		isValidChar := ValidateChar(char)
		if !isValidChar {
			return false
		}
	}

	return true
}
