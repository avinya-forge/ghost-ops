package runtime

import (
	"path/filepath"
	"strings"
)

// CheckFSJail verifies if the target path is within or strictly under
// any of the directories listed in allowedPaths.
func CheckFSJail(path string, allowedPaths []string) bool {
	if len(allowedPaths) == 0 {
		return false
	}

	cleanPath := filepath.Clean(path)

	for _, jail := range allowedPaths {
		cleanJail := filepath.Clean(jail)

		// If exact match
		if cleanPath == cleanJail {
			return true
		}

		// If it's a subpath.
		// E.g., cleanJail: /var/log, cleanPath: /var/log/app.log
		// Prefix check needs to ensure it's not a partial directory name match
		// E.g., /var/logs shouldn't match /var/log as jail.
		// So we append the separator.
		jailPrefix := cleanJail
		if !strings.HasSuffix(jailPrefix, string(filepath.Separator)) {
			jailPrefix += string(filepath.Separator)
		}

		if strings.HasPrefix(cleanPath, jailPrefix) {
			return true
		}
	}

	return false
}
