// Package security provides helpers to prevent sensitive values from appearing
// in log output. Constitution Principle VI: sensitive values MUST NEVER appear
// in log output at any level.
package security

import "strings"

// sensitiveKeys lists substrings that mark a config/log key as sensitive.
// Match is case-insensitive.
var sensitiveKeys = []string{
	"token",
	"password",
	"passwd",
	"secret",
	"apikey",
	"api_key",
	"auth",
	"credential",
	"private_key",
	"privatekey",
}

// Redact returns "[REDACTED]" when the key name contains a sensitive substring,
// otherwise it returns the original value unchanged. Comparison is case-insensitive.
func Redact(key, value string) string {
	lk := strings.ToLower(key)
	for _, s := range sensitiveKeys {
		if strings.Contains(lk, s) {
			return "[REDACTED]"
		}
	}
	return value
}
