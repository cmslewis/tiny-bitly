package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	description string
	input       string
	expected    bool
}

func TestValidateAlias(t *testing.T) {
	testCases := []testCase{
		{description: "InvalidEmpty", input: "", expected: false},
		{description: "InvalidCharUnderscore", input: "ABCabc123_", expected: false},
		{description: "ValidLength1", input: "A", expected: true},
		{description: "ValidLength2", input: "Aa", expected: true},
		{description: "ValidLength3", input: "Aa1", expected: true},
		{description: "ValidLength4", input: "Aa1B", expected: true},
		{description: "ValidLength5", input: "Aa1Bb", expected: true},
		{description: "ValidLength6", input: "Aa1Bb2", expected: true},
		{description: "ValidLength7", input: "Aa1Bb2C", expected: true},
		{description: "ValidChars", input: "ABCabc123", expected: true},
		{description: "ValidWord", input: "myalias", expected: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(tt *testing.T) {
			require.Equal(tt, ValidateAlias(testCase.input), testCase.expected)
		})
	}
}
