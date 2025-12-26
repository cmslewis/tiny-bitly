package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ValidateURLSuite struct {
	suite.Suite
}

func TestValidateURLSuite(t *testing.T) {
	suite.Run(t, new(ValidateURLSuite))
}

func (suite *ValidateURLSuite) TestDetectsValid() {
	validatedURL, err := ValidateURL("https://www.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com")
}

func (suite *ValidateURLSuite) TestAddsHTTPSIfMissing() {
	validatedURL, err := ValidateURL("www.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com")
}

func (suite *ValidateURLSuite) TestKeepsSubdomain() {
	validatedURL, err := ValidateURL("http://a.b.c.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "http://a.b.c.example.com")
}

func (suite *ValidateURLSuite) TestKeepsHTTP() {
	validatedURL, err := ValidateURL("http://www.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "http://www.example.com")
}

func (suite *ValidateURLSuite) TestKeepsFTP() {
	validatedURL, err := ValidateURL("ftp://www.example.com")
	suite.Nil(err)
	suite.Equal(*validatedURL, "ftp://www.example.com")
}

func (suite *ValidateURLSuite) TestKeepsTrailingSlash() {
	validatedURL, err := ValidateURL("ftp://www.example.com/")
	suite.Nil(err)
	suite.Equal(*validatedURL, "ftp://www.example.com/")
}

func (suite *ValidateURLSuite) TestKeepsPath() {
	validatedURL, err := ValidateURL("https://www.example.com/a/b/c")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com/a/b/c")
}

func (suite *ValidateURLSuite) TestKeepsParams() {
	validatedURL, err := ValidateURL("https://www.example.com?a=1&b=2&c=3")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com?a=1&b=2&c=3")
}

func (suite *ValidateURLSuite) TestKeepsAnchor() {
	validatedURL, err := ValidateURL("https://www.example.com#a")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://www.example.com#a")
}

func (suite *ValidateURLSuite) TestKeepsAll() {
	validatedURL, err := ValidateURL("https://a.b.c.example.com/a/b/c/?a=1&b=2&c=3#a")
	suite.Nil(err)
	suite.Equal(*validatedURL, "https://a.b.c.example.com/a/b/c/?a=1&b=2&c=3#a")
}

func (suite *ValidateURLSuite) TestLocalhostWithPort() {
	validatedURL, err := ValidateURL("http://localhost:3000/")
	suite.Nil(err)
	suite.Equal(*validatedURL, "http://localhost:3000/")
}

func (suite *ValidateURLSuite) TestInvalidEmpty() {
	validatedURL, err := ValidateURL("")
	suite.ErrorContains(err, "invalid URL")
	suite.Nil(validatedURL)
}

func (suite *ValidateURLSuite) TestInvalidProtocolOnly() {
	validatedURL, err := ValidateURL("https://")
	suite.ErrorContains(err, "invalid URL")
	suite.Nil(validatedURL)
}
