// Package config loads and validates simple-cli configuration.
// Constitution Principle II: configuration MUST honour the precedence chain:
// CLI flags > environment variables > config file > built-in defaults.
package config

import (
	"fmt"
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
	StateDir string `mapstructure:"state_dir"`
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

	// Resolve StateDir when not overridden.
	if cfg.StateDir == "" {
		cfg.StateDir = defaultStateDir()
	}

	return &cfg, nil
}

// StateDir returns the OS-appropriate session state directory.
// Callers should use the value from Config.StateDir instead of calling this
// directly; it is exported for testing.
func defaultStateDir() string {
	if runtime.GOOS == "windows" {
		if d := os.Getenv("APPDATA"); d != "" {
			return filepath.Join(d, "simple-cli")
		}
		return filepath.Join(os.TempDir(), "simple-cli")
	}
	// XDG Base Directory Spec: $XDG_STATE_HOME defaults to ~/.local/state
	if d := os.Getenv("XDG_STATE_HOME"); d != "" {
		return filepath.Join(d, "simple-cli")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "simple-cli")
	}
	return filepath.Join(home, ".local", "state", "simple-cli")
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
