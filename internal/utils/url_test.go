package utils

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestURLPathFullClean(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldReturnFullPathSingleSlash", "https://example.com/", "/"},
		{"ShouldReturnFullPathSingleSlashWithQuery", "https://example.com/?query=1&alt=2", "/?query=1&alt=2"},
		{"ShouldReturnFullPathNormal", "https://example.com/test", "/test"},
		{"ShouldReturnFullPathNormalWithSlashSuffix", "https://example.com/test/", "/test/"},
		{"ShouldReturnFullPathNormalWithSlashSuffixAndQuery", "https://example.com/test/?query=1&alt=2", "/test/?query=1&alt=2"},
		{"ShouldReturnFullPathWithQuery", "https://example.com/test?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPath", "https://example.com/five/../test?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscaped", "https://example.com/five/..%2ftest?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscapedExtra", "https://example.com/five/..%2ftest?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscapedExtraSurrounding", "https://example.com/five/%2f..%2f/test?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscapedPeriods", "https://example.com/five/%2f%2e%2e%2f/test?query=1&alt=2", "/test?query=1&alt=2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.have)
			require.NoError(t, err)

			actual := URLPathFullClean(u)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func isURLSafe(requestURI string, domains []schema.SessionDomainConfiguration) bool {
	url, _ := url.ParseRequestURI(requestURI)
	return IsRedirectionSafe(*url, domains)
}

func TestIsRedirectionSafe_ShouldReturnTrueOnExactDomain(t *testing.T) {
	assert.True(t, isURLSafe("https://example.com", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
}

func TestIsRedirectionSafe_ShouldReturnFalseOnBadScheme(t *testing.T) {
	assert.False(t, isURLSafe("http://secure.example.com", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
	assert.False(t, isURLSafe("ftp://secure.example.com", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
	assert.True(t, isURLSafe("https://secure.example.com", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
}

func TestIsRedirectionSafe_ShouldReturnFalseOnBadDomain(t *testing.T) {
	assert.False(t, isURLSafe("https://secure.example.com.c", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
	assert.False(t, isURLSafe("https://secure.example.comc", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
	assert.False(t, isURLSafe("https://secure.example.co", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
}

func TestIsRedirectionURISafe_CannotParseURI(t *testing.T) {
	_, err := IsRedirectionURISafe("http//invalid", []schema.SessionDomainConfiguration{{Domain: "example.com"}})
	assert.EqualError(t, err, "Unable to parse redirection URI http//invalid: parse \"http//invalid\": invalid URI for request")
}

func TestIsRedirectionURISafe_InvalidRedirectionURI(t *testing.T) {
	valid, err := IsRedirectionURISafe("http://myurl.com/myresource", []schema.SessionDomainConfiguration{{Domain: "example.com"}})
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestIsRedirectionURISafe_ValidRedirectionURI(t *testing.T) {
	valid, err := IsRedirectionURISafe("http://myurl.example.com/myresource", []schema.SessionDomainConfiguration{{Domain: "example.com"}})
	assert.NoError(t, err)
	assert.False(t, valid)

	valid, err = IsRedirectionURISafe("http://example.com/myresource", []schema.SessionDomainConfiguration{{Domain: "example.com"}})
	assert.NoError(t, err)
	assert.False(t, valid)
}
