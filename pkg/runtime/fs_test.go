package runtime

import (
	"path/filepath"
	"testing"
)

func TestCheckFSJail(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		allowedPaths []string
		expected     bool
	}{
		{
			name:         "Empty allowed paths denies all",
			path:         "/var/log/app.log",
			allowedPaths: []string{},
			expected:     false,
		},
		{
			name:         "Exact match",
			path:         "/var/data/ghost",
			allowedPaths: []string{"/var/data/ghost"},
			expected:     true,
		},
		{
			name:         "Subpath match",
			path:         "/tmp/sandbox/logs/output.txt",
			allowedPaths: []string{"/tmp/sandbox"},
			expected:     true,
		},
		{
			name:         "Partial directory name bypass attempt",
			path:         "/tmp/sandbox-hacked/data",
			allowedPaths: []string{"/tmp/sandbox"},
			expected:     false,
		},
		{
			name:         "Relative path traversal bypass attempt",
			path:         "/tmp/sandbox/../../etc/passwd",
			allowedPaths: []string{"/tmp/sandbox"},
			expected:     false, // cleanPath becomes /etc/passwd
		},
		{
			name:         "Not in allowed paths",
			path:         "/etc/shadow",
			allowedPaths: []string{"/opt/app/data"},
			expected:     false,
		},
		{
			name:         "Relative allowed path and relative target",
			path:         "data/output.txt",
			allowedPaths: []string{"data"},
			expected:     true,
		},
		{
			name:         "Relative allowed path trailing slash",
			path:         "data/output.txt",
			allowedPaths: []string{"data/"},
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var allowed []string
			for _, p := range tt.allowedPaths {
				allowed = append(allowed, filepath.FromSlash(p))
			}
			result := CheckFSJail(filepath.FromSlash(tt.path), allowed)

			if result != tt.expected {
				t.Errorf("CheckFSJail(%q, %v) = %v; want %v", tt.path, tt.allowedPaths, result, tt.expected)
			}
		})
	}
}
