package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/your-org/simple-cli/internal/config"
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
	assert.NotEmpty(t, cfg.StateDir)
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

func TestStateDirFromEnvXDG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG_STATE_HOME not used on Windows")
	}
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)
	t.Setenv("APPDATA", "") // ensure APPDATA doesn't win

	v := newViper()
	cfg, err := config.Load(v)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "simple-cli"), cfg.StateDir)
}

func TestStateDirFallbackHomeDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG fallback not used on Windows")
	}
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", dir)

	v := newViper()
	cfg, err := config.Load(v)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, ".local", "state", "simple-cli"), cfg.StateDir)
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

func TestStateDirFromEnvAPPDATA(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	// Clear XDG so APPDATA wins on any OS when we fake-test it.
	t.Setenv("XDG_STATE_HOME", "")

	v := newViper()
	_ = v.BindEnv("state_dir", "APPDATA") // real win path would be set by OS
	cfg, err := config.Load(v)
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.StateDir)
}

func TestStateDirOverride(t *testing.T) {
	dir := t.TempDir()
	v := newViper()
	v.Set("state_dir", dir)

	cfg, err := config.Load(v)
	require.NoError(t, err)
	assert.Equal(t, dir, cfg.StateDir)
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
