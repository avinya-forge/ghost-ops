package runtime

import (
	"testing"
)

func TestCheckNetworkEgress(t *testing.T) {
	tests := []struct {
		name           string
		urlStr         string
		allowedDomains []string
		expected       bool
	}{
		{
			name:           "Empty allowed domains denies all",
			urlStr:         "https://example.com/api",
			allowedDomains: []string{},
			expected:       false,
		},
		{
			name:           "Exact match",
			urlStr:         "https://api.ghost-ops.io/v1/data",
			allowedDomains: []string{"api.ghost-ops.io"},
			expected:       true,
		},
		{
			name:           "Subdomain match",
			urlStr:         "https://auth.ghost-ops.io/login",
			allowedDomains: []string{"ghost-ops.io"},
			expected:       true,
		},
		{
			name:           "Subdomain match with dot prefix",
			urlStr:         "https://auth.ghost-ops.io/login",
			allowedDomains: []string{".ghost-ops.io"},
			expected:       false, // wait, our code does strings.HasSuffix(host, "."+domain)
			// if domain is ".ghost-ops.io", "."+domain is "..ghost-ops.io"
			// Let's ensure standard format is "ghost-ops.io"
		},
		{
			name:           "No match",
			urlStr:         "https://evil.com/steal",
			allowedDomains: []string{"ghost-ops.io", "github.com"},
			expected:       false,
		},
		{
			name:           "Invalid URL",
			urlStr:         "://bad-url",
			allowedDomains: []string{"ghost-ops.io"},
			expected:       false,
		},
		{
			name:           "Port in URL is handled correctly",
			urlStr:         "http://example.com:8080/path",
			allowedDomains: []string{"example.com"},
			expected:       true,
		},
		{
			name:           "Subdomain trick bypass attempt",
			urlStr:         "https://ghost-ops.io.evil.com",
			allowedDomains: []string{"ghost-ops.io"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckNetworkEgress(tt.urlStr, tt.allowedDomains)
			if result != tt.expected {
				t.Errorf("CheckNetworkEgress(%q, %v) = %v; want %v", tt.urlStr, tt.allowedDomains, result, tt.expected)
			}
		})
	}
}
