package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/binpqh/simple-cli/internal/config"
)

func newViper() *viper.Viper {
	return viper.New()
}

func TestLoadDefaults(t *testing.T) {
	v := newViper()
	cfg, err := config.Load(v)
	require.NoError(t, err)
	assert.Equal(t, "human", cfg.Output)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.False(t, cfg.NoColor)
	assert.False(t, cfg.Quiet)
}

func TestLoadFromViper(t *testing.T) {
	v := newViper()
	v.Set("output", "json")
	v.Set("log_level", "debug")
	v.Set("no_color", true)
	v.Set("quiet", true)

	cfg, err := config.Load(v)
	require.NoError(t, err)
	assert.Equal(t, "json", cfg.Output)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.True(t, cfg.NoColor)
	assert.True(t, cfg.Quiet)
}

func TestLoadInvalidOutput(t *testing.T) {
	v := newViper()
	v.Set("output", "xml")
	_, err := config.Load(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output")
}

func TestLoadInvalidLogLevel(t *testing.T) {
	v := newViper()
	v.Set("log_level", "verbose")
	_, err := config.Load(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log_level")
}

func TestConfigDirXDG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG config not used on Windows")
	}
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	result := config.ConfigDir()
	assert.Equal(t, filepath.Join(dir, "simple-cli"), result)
}

func TestConfigDirXDGFallback(t *testing.T) {
	// Clear APPDATA so the Windows branch falls through to XDG check.
	t.Setenv("APPDATA", "")
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	result := config.ConfigDir()
	assert.Equal(t, filepath.Join(dir, "simple-cli"), result)
}

func TestConfigDirHomeFallback(t *testing.T) {
	// Clear both APPDATA and XDG_CONFIG_HOME to exercise the home directory fallback.
	t.Setenv("APPDATA", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	result := config.ConfigDir()
	assert.NotEmpty(t, result)
	assert.True(t, filepath.IsAbs(result), "ConfigDir fallback must return absolute path")
}

func TestConfigDirNotEmpty(t *testing.T) {
	dir := config.ConfigDir()
	assert.NotEmpty(t, dir)
}

func TestLoadFromConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("output: json\nlog_level: warn\n"), 0o600))

	v := newViper()
	v.SetConfigFile(cfgFile)
	require.NoError(t, v.ReadInConfig())

	cfg, err := config.Load(v)
	require.NoError(t, err)
	assert.Equal(t, "json", cfg.Output)
	assert.Equal(t, "warn", cfg.LogLevel)
}

func TestProviderConfigUnmarshalAndActiveProvider(t *testing.T) {
	v := newViper()
	// Build providers map programmatically to avoid YAML parsing issues on CI
	providers := map[string]interface{}{
		"my-api": map[string]interface{}{
			"client_id":       "id-123",
			"device_endpoint": "https://auth.example.com/device",
			"token_endpoint":  "https://auth.example.com/token",
			"chat_endpoint":   "https://api.example.com/v1/chat/completions",
			"scopes":          []string{"chat"},
			"default_model":   "gpt-4o",
		},
		"other": map[string]interface{}{
			"client_id":       "id-456",
			"device_endpoint": "https://auth2.example.com/device",
			"token_endpoint":  "https://auth2.example.com/token",
			"chat_endpoint":   "https://api2.example.com/v1/chat/completions",
			"scopes":          []string{},
			"default_model":   "gpt-4o-mini",
		},
	}
	v.Set("output", "json")
	v.Set("default_provider", "my-api")
	v.Set("providers", providers)

	cfg, err := config.Load(v)
	require.NoError(t, err)

	// Providers unmarshalled
	require.NotNil(t, cfg.Providers)
	assert.Contains(t, cfg.Providers, "my-api")
	assert.Contains(t, cfg.Providers, "other")

	// ActiveProvider by name
	pc, err := cfg.ActiveProvider("other")
	require.NoError(t, err)
	assert.Equal(t, "id-456", pc.ClientID)

	// ActiveProvider fallback to default when name empty
	def, err := cfg.ActiveProvider("")
	require.NoError(t, err)
	assert.Equal(t, "id-123", def.ClientID)
}

func TestActiveProviderErrors(t *testing.T) {
	// Config with providers but no default_provider
	v := newViper()
	v.Set("providers.some", map[string]interface{}{"client_id": "x", "device_endpoint": "https://a", "token_endpoint": "https://b", "chat_endpoint": "https://c"})
	cfg, err := config.Load(v)
	require.NoError(t, err)

	_, err = cfg.ActiveProvider("")
	require.Error(t, err)

	_, err = cfg.ActiveProvider("nonexistent")
	require.Error(t, err)
}
