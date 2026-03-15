package runtime

import (
	"net/url"
	"strings"
)

// CheckNetworkEgress checks if the requested URL string is allowed
// by the given network egress policies (allowed domains).
// It returns true if the request is allowed.
// If allowedDomains is empty, it returns false (deny by default if configured to check).
func CheckNetworkEgress(urlStr string, allowedDomains []string) bool {
	if len(allowedDomains) == 0 {
		return false
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	host := u.Hostname()

	for _, domain := range allowedDomains {
		// Exact match
		if host == domain {
			return true
		}
		// Subdomain match (e.g. host is "api.example.com" and domain is ".example.com" or "example.com")
		// To safely check subdomain, we check if host ends with "." + domain
		// For example, host: "api.example.com", domain: "example.com" -> host ends with ".example.com"
		if strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}
