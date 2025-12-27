package create

import (
	"math/rand/v2"
	"strings"
)

// Generates a random short code of the specified length using the characters
// A-Z, a-z, and 0-9.
func generateShortCode(length int) string {
	var builder strings.Builder
	for range length {
		char := getRandomChar()
		builder.WriteByte(char)
	}
	return builder.String()
}

// Returns a random base62 character.
func getRandomChar() byte {
	charIndex := randInRange(0, len(allowedChars))
	return allowedChars[charIndex]
}

// Returns a random integer in [min, max).
func randInRange(min, max int) int {
	return min + rand.IntN(max-min)
}
