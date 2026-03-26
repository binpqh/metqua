package security_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/binpqh/simple-cli/internal/security"
)

func TestRedactSensitiveKeys(t *testing.T) {
	sensitiveKeys := []string{
		"token", "TOKEN", "access_token",
		"password", "PASSWORD", "db_password",
		"secret", "SECRET", "client_secret",
		"api_key", "API_KEY", "private_key",
	}
	for _, key := range sensitiveKeys {
		t.Run(key, func(t *testing.T) {
			got := security.Redact(key, "super-secret-value")
			assert.Equal(t, "[REDACTED]", got)
		})
	}
}

func TestRedactSafeKeys(t *testing.T) {
	safeKeys := []string{
		"output", "log_level", "state_dir", "name", "id", "status",
	}
	for _, key := range safeKeys {
		t.Run(key, func(t *testing.T) {
			got := security.Redact(key, "visible-value")
			assert.Equal(t, "visible-value", got)
		})
	}
}
