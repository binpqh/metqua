//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryPath string

func TestMain(m *testing.M) {
	binaryPath = filepath.Join("..", "..", "dist", "simple-cli")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	os.Exit(m.Run())
}

func skipIfNoBinary(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("binary not found at %s — run 'make build' first", binaryPath)
	}
}

// run executes the binary with the given args and returns stdout, stderr, exit code.
func run(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = os.Environ()
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exit := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exit = exitErr.ExitCode()
		} else {
			exit = -1
		}
	}
	return outBuf.String(), errBuf.String(), exit
}

// runEnvJSON executes the binary with SIMPLE_CLI_OUTPUT=json set via environment (not flag).
func runEnvJSON(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "SIMPLE_CLI_OUTPUT=json")
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exit := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exit = exitErr.ExitCode()
		} else {
			exit = -1
		}
	}
	return outBuf.String(), errBuf.String(), exit
}

func TestVersionOutput(t *testing.T) {
	skipIfNoBinary(t)
	stdout, _, exit := run(t, "--version")
	require.Equal(t, 0, exit)
	assert.NotEmpty(t, stdout)
}

func TestHelpNoSessionMention(t *testing.T) {
	skipIfNoBinary(t)
	stdout, _, _ := run(t, "--help")
	assert.NotContains(t, strings.ToLower(stdout), "session",
		"help output must not mention 'session'")
}

func TestJSONOutputFlag(t *testing.T) {
	skipIfNoBinary(t)
	stdout, _, exit := run(t, "--output", "json", "example")
	require.Equal(t, 0, exit)

	var resp struct {
		Status string `json:"status"`
		Meta   struct {
			Version    string `json:"version"`
			DurationMs int64  `json:"duration_ms"`
			Command    string `json:"command"`
		} `json:"meta"`
	}
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp))
	assert.Equal(t, "ok", resp.Status)
	assert.NotEmpty(t, resp.Meta.Version)
	assert.GreaterOrEqual(t, resp.Meta.DurationMs, int64(0))
}

func TestEnvVarOutputJSON(t *testing.T) {
	skipIfNoBinary(t)
	stdout, _, exit := runEnvJSON(t, "example")
	require.Equal(t, 0, exit)

	var resp map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp))
	assert.Equal(t, "ok", resp["status"])
}

func TestInvalidOutputFlag(t *testing.T) {
	skipIfNoBinary(t)
	_, _, exit := run(t, "--output", "invalid", "example")
	assert.NotEqual(t, 0, exit, "invalid output flag should produce non-zero exit")
}
