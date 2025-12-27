package utils

import (
	"strings"
)

// Generates a random short code of the specified length using the characters
// A-Z, a-z, and 0-9.
func GenerateShortCode(length int) string {
	var builder strings.Builder
	for range length {
		char := GetRandomChar()
		builder.WriteByte(char)
	}
	return builder.String()
}
