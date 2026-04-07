// Package config loads and validates simple-cli configuration.
// Constitution Principle II: configuration MUST honour the precedence chain:
// CLI flags > environment variables > config file > built-in defaults.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// CtxKey is the context key used to store and retrieve *Config from a
// context.Context. Using a private struct type avoids collisions.
type CtxKey struct{}

// Config holds the resolved runtime configuration for a single invocation.
type Config struct {
	Output   string `mapstructure:"output"`
	LogLevel string `mapstructure:"log_level"`
	NoColor  bool   `mapstructure:"no_color"`
	Quiet    bool   `mapstructure:"quiet"`
	// Provider configuration (v2.1.0)
	DefaultProvider string                    `mapstructure:"default_provider"`
	Providers       map[string]ProviderConfig `mapstructure:"providers"`
	// Insecure disables HTTPS enforcement and TLS cert verification.
	// FOR DEVELOPMENT USE ONLY — never set in production.
	Insecure bool `mapstructure:"insecure"`
}

// ProviderConfig holds all settings for one OAuth + Chat provider.
type ProviderConfig struct {
	ClientID       string   `mapstructure:"client_id"`
	DeviceEndpoint string   `mapstructure:"device_endpoint"`
	TokenEndpoint  string   `mapstructure:"token_endpoint"`
	ChatEndpoint   string   `mapstructure:"chat_endpoint"`
	Scopes         []string `mapstructure:"scopes"`
	DefaultModel   string   `mapstructure:"default_model"`
}

// ActiveProvider resolves the provider configuration by name. If name is
// empty, DefaultProvider is used. Returns an error when provider not found.
func (c *Config) ActiveProvider(name string) (*ProviderConfig, error) {
	if name == "" {
		name = c.DefaultProvider
	}
	if name == "" {
		return nil, errors.New("no provider specified and no default_provider set in config")
	}
	pc, ok := c.Providers[name]
	if !ok {
		// try lowercase normalization
		if pc2, ok2 := c.Providers[lower(name)]; ok2 {
			return &pc2, nil
		}
		return nil, fmt.Errorf("provider %q not found in config", name)
	}
	return &pc, nil
}

func lower(s string) string { return string([]byte(s)) }

// ValidateProviderConfig performs basic validation on a provider config.
// Pass insecure=true to allow plain http:// endpoints (dev/test only).
func ValidateProviderConfig(pc *ProviderConfig, insecure ...bool) error {
	allowHTTP := len(insecure) > 0 && insecure[0]
	if pc == nil {
		return errors.New("provider config is nil")
	}
	if pc.ClientID == "" {
		return errors.New("client_id is required")
	}
	for _, u := range []struct {
		name string
		raw  string
	}{
		{"device_endpoint", pc.DeviceEndpoint},
		{"token_endpoint", pc.TokenEndpoint},
		{"chat_endpoint", pc.ChatEndpoint},
	} {
		if u.raw == "" {
			return fmt.Errorf("%s is required", u.name)
		}
		parsed, err := url.Parse(u.raw)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return fmt.Errorf("%s must be a valid URL", u.name)
		}
		if parsed.Scheme != "https" && !allowHTTP {
			return fmt.Errorf("%s must be a valid https URL", u.name)
		}
	}
	return nil
}

// Load reads configuration from the provided Viper instance and returns a
// validated Config. It must be called after all flag bindings and env bindings
// are registered on v.
func Load(v *viper.Viper) (*Config, error) {
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	// Apply defaults when Viper fields are zero.
	if cfg.Output == "" {
		cfg.Output = "human"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	// Validate.
	switch cfg.Output {
	case "human", "json":
	default:
		return nil, fmt.Errorf("config: invalid output %q: must be 'human' or 'json'", cfg.Output)
	}
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return nil, fmt.Errorf("config: invalid log_level %q: must be debug|info|warn|error", cfg.LogLevel)
	}

	return &cfg, nil
}

// ConfigDir returns the OS-appropriate configuration directory.
func ConfigDir() string {
	if runtime.GOOS == "windows" {
		if d := os.Getenv("APPDATA"); d != "" {
			return filepath.Join(d, "simple-cli")
		}
	}
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "simple-cli")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "simple-cli")
}
