package create

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestValidateAlias(t *testing.T) {
	type testCase struct {
		description string
		input       string
		expected    bool
	}

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
			require.Equal(tt, validateAlias(testCase.input), testCase.expected)
		})
	}
}

type ValidateURLSuite struct {
	suite.Suite
}

func TestValidateURLSuite(t *testing.T) {
	suite.Run(t, new(ValidateURLSuite))
}

func (suite *ValidateURLSuite) TestDetectsValid() {
	validatedURL, err := validateURL("https://www.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com")
}

func (suite *ValidateURLSuite) TestAddsHTTPSIfMissing() {
	validatedURL, err := validateURL("www.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com")
}

func (suite *ValidateURLSuite) TestKeepsSubdomain() {
	validatedURL, err := validateURL("http://a.b.c.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "http://a.b.c.example.com")
}

func (suite *ValidateURLSuite) TestKeepsHTTP() {
	validatedURL, err := validateURL("http://www.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "http://www.example.com")
}

func (suite *ValidateURLSuite) TestKeepsFTP() {
	validatedURL, err := validateURL("ftp://www.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "ftp://www.example.com")
}

func (suite *ValidateURLSuite) TestKeepsTrailingSlash() {
	validatedURL, err := validateURL("ftp://www.example.com/")
	suite.Nil(err)
	suite.Equal(*validatedURL, "ftp://www.example.com/")
}

func (suite *ValidateURLSuite) TestKeepsPath() {
	validatedURL, err := validateURL("https://www.example.com/a/b/c")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com/a/b/c")
}

func (suite *ValidateURLSuite) TestKeepsParams() {
	validatedURL, err := validateURL("https://www.example.com?a=1&b=2&c=3")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com?a=1&b=2&c=3")
}

func (suite *ValidateURLSuite) TestKeepsAnchor() {
	validatedURL, err := validateURL("https://www.example.com#a")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com#a")
}

func (suite *ValidateURLSuite) TestKeepsAll() {
	validatedURL, err := validateURL("https://a.b.c.example.com/a/b/c/?a=1&b=2&c=3#a")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://a.b.c.example.com/a/b/c/?a=1&b=2&c=3#a")
}

func (suite *ValidateURLSuite) TestLocalhostWithPort() {
	validatedURL, err := validateURL("http://localhost:3000/")
	suite.Nil(err)
	suite.Equal(*validatedURL, "http://localhost:3000/")
}

func (suite *ValidateURLSuite) TestInvalidEmpty() {
	validatedURL, err := validateURL("")
	suite.ErrorContains(err, "invalid URL")
	suite.Nil(validatedURL)
}

func (suite *ValidateURLSuite) TestInvalidProtocolOnly() {
	validatedURL, err := validateURL("https://")
	suite.ErrorContains(err, "invalid URL")
	suite.Nil(validatedURL)
}
