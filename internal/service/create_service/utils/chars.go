package utils

import "math/rand/v2"

// Contains all characters that can appear in a short code.
var AllowedChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// Returns a random base62 character.
func GetRandomChar() byte {
	charIndex := randInRange(0, len(AllowedChars))
	return AllowedChars[charIndex]
}

// Returns true if the provided character is a valid base62 character, or false
// otherwise.
func ValidateChar(char byte) bool {
	// O(N) is fine for only 62 elements
	for i := 0; i < len(AllowedChars); i += 1 {
		allowedChar := AllowedChars[i]
		if char == allowedChar {
			return true
		}
	}
	return false
}

// Returns a random integer in [min, max).
func randInRange(min, max int) int {
	return min + rand.IntN(max-min)
}
