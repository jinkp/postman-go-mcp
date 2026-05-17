// Package security provides helpers for safe file operations and header redaction.
package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

const redacted = "[REDACTED]"

// RedactHeader returns a redacted value if the header key is sensitive.
// Sensitive headers: Authorization, X-API-Key, X-Auth-Token, Cookie, Set-Cookie.
func RedactHeader(key, value string) string {
	sensitive := []string{
		"authorization",
		"x-api-key",
		"x-auth-token",
		"cookie",
		"set-cookie",
		"proxy-authorization",
	}
	lower := strings.ToLower(key)
	for _, s := range sensitive {
		if lower == s {
			return redacted
		}
	}
	return value
}

// ValidatePath checks that outputPath is safe to write to:
//   - Must not contain ".." (path traversal)
//   - Must be an absolute or relative path without traversal sequences
//
// Returns an error if the path is considered unsafe.
func ValidatePath(outputPath string) error {
	cleaned := filepath.Clean(outputPath)
	// Reject any path that still contains ".." after cleaning
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("invalid output path: path traversal detected in %q", outputPath)
	}
	// Reject paths that look like they escape current context via ".."
	if strings.HasPrefix(outputPath, "..") {
		return fmt.Errorf("invalid output path: path must not start with '..'")
	}
	return nil
}

// IsDryRun returns true when DRY_RUN is enabled (passed from config).
// This helper allows domain code to check dry-run without importing config.
func IsDryRun(dryRun bool) bool {
	return dryRun
}
