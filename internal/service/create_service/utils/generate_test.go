package utils

import (
	"testing"
)

func TestGenerateShortCodeLength(t *testing.T) {
	result := GenerateShortCode()
	expectedLength := 8
	if len(result) != expectedLength {
		t.Errorf("Expected short code length %d, got %d", expectedLength, len(result))
	}
}
