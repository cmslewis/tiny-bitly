package utils

import (
	"testing"
)

func TestGenerateShortCodeLength(t *testing.T) {
	expectedLength := 6
	result := GenerateShortCode(expectedLength)
	if len(result) != expectedLength {
		t.Errorf("Expected short code length %d, got %d", expectedLength, len(result))
	}
}
