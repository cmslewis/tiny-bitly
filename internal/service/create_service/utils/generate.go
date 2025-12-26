package utils

import (
	"math/rand/v2"
	"strings"
)

// length is the number of characters each short code should have.
var length = 6

// chars contains all characters that can appear in a short code.
var chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// Generates a random 6-character short code using the characters A-Z, a-z, and
// 0-9. This gives 56B+ unique codes.
func GenerateShortCode() string {
	var builder strings.Builder
	for range length {
		charIndex := randRange(0, len(chars))
		char := chars[charIndex]
		builder.WriteByte(char)
	}
	return builder.String()
}

// randRange returns a random integer in [min, max).
func randRange(min, max int) int {
	return min + rand.IntN(max-min)
}
