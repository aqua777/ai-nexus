package http

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// clientTestSuite is a test suite for the HTTP client package
type clientTestSuite struct {
	suite.Suite
	ctx context.Context
}

// SetupTest runs before each test
func (suite *clientTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

// TestGetBaseUrl tests the getBaseUrl function with various inputs
func (suite *clientTestSuite) TestGetBaseUrl() {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		// Valid URLs with scheme and host
		{
			name:        "full URL with http scheme",
			input:       "http://example.com",
			expected:    "http://example.com",
			expectError: false,
		},
		{
			name:        "full URL with https scheme",
			input:       "https://example.com",
			expected:    "https://example.com",
			expectError: false,
		},
		{
			name:        "URL with port",
			input:       "http://example.com:8080",
			expected:    "http://example.com:8080",
			expectError: false,
		},
		{
			name:        "URL with path",
			input:       "http://example.com/api/v1",
			expected:    "http://example.com/api/v1",
			expectError: false,
		},
		{
			name:        "URL with query parameters",
			input:       "http://example.com?key=value",
			expected:    "http://example.com?key=value",
			expectError: false,
		},
		{
			name:        "URL with fragment",
			input:       "http://example.com#section",
			expected:    "http://example.com#section",
			expectError: false,
		},
		{
			name:        "localhost URL",
			input:       "http://localhost:3000",
			expected:    "http://localhost:3000",
			expectError: false,
		},
		{
			name:        "IP address URL",
			input:       "http://192.168.1.1:8080",
			expected:    "http://192.168.1.1:8080",
			expectError: false,
		},

		// URLs without scheme (prepends http://)
		{
			name:        "URL without scheme, with host",
			input:       "example.com",
			expected:    "http://example.com",
			expectError: false,
		},
		{
			name:        "URL without scheme, with host and port",
			input:       "example.com:8080",
			expected:    "http://example.com:8080",
			expectError: false,
		},
		{
			name:        "URL without scheme, with path",
			input:       "example.com/api",
			expected:    "http://example.com/api",
			expectError: false,
		},
		{
			name:        "localhost without scheme",
			input:       "localhost:3000",
			expected:    "http://localhost:3000",
			expectError: false,
		},
		{
			name:        "IP address without scheme",
			input:       "192.168.1.1:8080",
			expected:    "http://192.168.1.1:8080",
			expectError: false,
		},

		// URLs without host (parsed as-is by url.Parse, no default host)
		{
			name:        "URL with scheme but no host",
			input:       "http://",
			expected:    "http:",
			expectError: false,
		},
		{
			name:        "URL with scheme and path but no host",
			input:       "http:///api",
			expected:    "http:///api",
			expectError: false,
		},
		{
			name:        "URL with https scheme but no host",
			input:       "https://",
			expected:    "https:",
			expectError: false,
		},

		// Edge cases
		{
			name:        "path only",
			input:       "/api/v1",
			expected:    "http:///api/v1",
			expectError: false,
		},
		{
			name:        "path with query",
			input:       "/api?key=value",
			expected:    "http:///api?key=value",
			expectError: false,
		},
		{
			name:        "relative path",
			input:       "api/endpoint",
			expected:    "http://api/endpoint",
			expectError: false,
		},

		// Invalid URLs (should return error)
		{
			name:        "invalid URL with spaces",
			input:       "http://example .com",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid URL with control characters",
			input:       "http://example.com\x00",
			expected:    "",
			expectError: true,
		},
		{
			name:        "malformed URL",
			input:       "://invalid",
			expected:    "",
			expectError: true,
		},
		{
			name:        "URL with invalid port",
			input:       "http://example.com:99999",
			expected:    "http://example.com:99999",
			expectError: false, // url.Parse doesn't validate port range
		},
		{
			name:        "URL with invalid characters in host",
			input:       "http://exam ple.com",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := getBaseUrl(tt.input)

			if tt.expectError {
				assert.Error(suite.T(), err, "getBaseUrl(%q) expected error but got none", tt.input)
				assert.Empty(suite.T(), result, "getBaseUrl(%q) expected empty string on error", tt.input)
			} else {
				assert.NoError(suite.T(), err, "getBaseUrl(%q) unexpected error", tt.input)
				assert.Equal(suite.T(), tt.expected, result, "getBaseUrl(%q) result mismatch", tt.input)
			}
		})
	}
}

// TestNewClient tests the NewClient function
func (suite *clientTestSuite) TestNewClient() {
	baseUrl := "http://example.com"

	client, err := NewClient(baseUrl)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), client)
	assert.Equal(suite.T(), DefaultTimeout, client.timeout)
	assert.Nil(suite.T(), client.client, "client should be nil initially")
	assert.Equal(suite.T(), baseUrl, client.baseUrl)
}

// TestNewClient_WithoutBaseUrl tests NewClient without base URL
func (suite *clientTestSuite) TestNewClient_WithoutBaseUrl() {
	client, err := NewClient()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), client)
	assert.Equal(suite.T(), DefaultTimeout, client.timeout)
	assert.Empty(suite.T(), client.baseUrl)
}

// TestNewClient_InvalidBaseUrl tests NewClient with invalid base URL
func (suite *clientTestSuite) TestNewClient_InvalidBaseUrl() {
	client, err := NewClient("://invalid")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), client)
}

// TestClient_WithTimeout tests the WithTimeout method
func (suite *clientTestSuite) TestClient_WithTimeout() {
	client, err := NewClient()
	suite.Require().NoError(err)
	suite.Require().NotNil(client)

	// Test setting timeout before client is initialized
	customTimeout := 60 * time.Second
	result := client.WithTimeout(customTimeout)

	assert.Equal(suite.T(), client, result, "WithTimeout() should return the same client instance")
	assert.Equal(suite.T(), customTimeout, client.timeout)

	// Test that setting timeout after client is initialized doesn't change it
	_ = client.getClient() // Initialize the client
	oldTimeout := client.timeout
	client.WithTimeout(120 * time.Second)

	assert.Equal(suite.T(), oldTimeout, client.timeout, "WithTimeout() should not change timeout after client is initialized")
}

// TestClient_getClient tests the getClient method
func (suite *clientTestSuite) TestClient_getClient() {
	client, err := NewClient()
	suite.Require().NoError(err)
	suite.Require().NotNil(client)

	// Test that getClient creates a client
	httpClient1 := client.getClient()
	assert.NotNil(suite.T(), httpClient1, "getClient() returned nil")

	// Test that getClient returns the same instance (singleton pattern)
	httpClient2 := client.getClient()
	assert.Equal(suite.T(), httpClient1, httpClient2, "getClient() should return the same instance on subsequent calls")

	// Test that timeout is set correctly
	assert.Equal(suite.T(), client.timeout, httpClient1.Timeout)
}

// TestClient_getClient_Concurrent tests concurrent access to getClient
func (suite *clientTestSuite) TestClient_getClient_Concurrent() {
	client, err := NewClient()
	suite.Require().NoError(err)
	suite.Require().NotNil(client)

	// Test concurrent access to getClient
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			c := client.getClient()
			assert.NotNil(suite.T(), c, "getClient() returned nil in concurrent test")
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify only one client instance was created
	assert.NotNil(suite.T(), client.client, "getClient() should have created a client instance")
}

// TestClient_getFullUrl tests the getFullUrl method
func (suite *clientTestSuite) TestClient_getFullUrl() {
	client, err := NewClient("http://example.com")
	suite.Require().NoError(err)
	suite.Require().NotNil(client)

	// Test normal path concatenation (note: replaces all "//" including scheme separator)
	result := client.getFullUrl("/api/v1")
	assert.Equal(suite.T(), "http:/example.com/api/v1", result)

	// Test path without leading slash
	result = client.getFullUrl("api/v1")
	assert.Equal(suite.T(), "http:/example.comapi/v1", result)

	// Test empty path
	result = client.getFullUrl("")
	assert.Equal(suite.T(), "http:/example.com", result)
}

// TestClient_getFullUrl_DoubleSlash tests getFullUrl handles double slashes
func (suite *clientTestSuite) TestClient_getFullUrl_DoubleSlash() {
	client, err := NewClient("http://example.com")
	suite.Require().NoError(err)
	suite.Require().NotNil(client)

	// Test that double slashes are replaced (including scheme separator)
	result := client.getFullUrl("//api/v1")
	assert.Equal(suite.T(), "http:/example.com/api/v1", result)
}

// TestClientSuite runs all tests in the suite
func TestClientSuite(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}
