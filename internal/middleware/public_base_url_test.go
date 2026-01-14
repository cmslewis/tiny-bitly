package middleware

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PublicBaseURLSuite struct {
	suite.Suite
}

func TestPublicBaseURLSuite(t *testing.T) {
	suite.Run(t, new(PublicBaseURLSuite))
}

func (suite *PublicBaseURLSuite) TestPrefersXForwardedHeaders() {
	r := httptest.NewRequest("GET", "http://internal.svc/urls", nil)
	r.Host = "internal.svc"
	r.Header.Set("X-Forwarded-Host", "example.com")
	r.Header.Set("X-Forwarded-Proto", "https")

	got := PublicBaseURL(r, "http://fallback.invalid")
	want := "https://example.com"
	suite.Equal(want, got)
}

func (suite *PublicBaseURLSuite) TestUsesFirstXForwardedValueWhenCommaSeparated() {
	r := httptest.NewRequest("GET", "http://internal.svc/urls", nil)
	r.Header.Set("X-Forwarded-Host", "example.com, proxy.local")
	r.Header.Set("X-Forwarded-Proto", "https, http")

	got := PublicBaseURL(r, "http://fallback.invalid")
	want := "http://proxy.local"
	suite.Equal(want, got)
}

func (suite *PublicBaseURLSuite) TestParsesForwardedHeader() {
	r := httptest.NewRequest("GET", "http://internal.svc/urls", nil)
	r.Header.Set("Forwarded", `for=192.0.2.43;proto=https;host=example.com:8443`)

	got := PublicBaseURL(r, "http://fallback.invalid")
	want := "https://example.com:8443"
	suite.Equal(want, got)
}

func (suite *PublicBaseURLSuite) TestParsesForwardedHeader_UsesLastEntryWhenCommaSeparated() {
	r := httptest.NewRequest("GET", "http://internal.svc/urls", nil)
	r.Header.Set("Forwarded", `for=192.0.2.1;proto=http;host=evil.example, for=192.0.2.2;proto=https;host=good.example`)

	got := PublicBaseURL(r, "http://fallback.invalid")
	want := "https://good.example"
	suite.Equal(want, got)
}

func (suite *PublicBaseURLSuite) TestUsesHostAndTLSWhenNoForwardedHeaders() {
	r := httptest.NewRequest("GET", "http://internal/urls", nil)
	r.Host = "api.example.com"
	r.TLS = &tls.ConnectionState{}

	got := PublicBaseURL(r, "http://fallback.invalid")
	want := "https://api.example.com"
	suite.Equal(want, got)
}

func (suite *PublicBaseURLSuite) TestFallsBackToConfiguredBaseURLWhenHostMissing() {
	r := httptest.NewRequest("GET", "http://internal/urls", nil)
	r.Host = ""

	got := PublicBaseURL(r, "http://fallback.example.com/")
	want := "http://fallback.example.com"
	suite.Equal(want, got)
}
