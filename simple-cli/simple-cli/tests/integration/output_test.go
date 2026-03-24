//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONOutputFlag(t *testing.T) {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("binary not found at %s — run 'make build' first", binaryPath)
	}

	stateDir := t.TempDir()
	stdout, _, exit := run(t, stateDir, "session", "list")
	require.Equal(t, 0, exit, "session list should succeed")

	// Validate JSON envelope structure.
	var resp struct {
		Status string `json:"status"`
		Data   any    `json:"data"`
		Meta   struct {
			Version    string `json:"version"`
			DurationMs int64  `json:"duration_ms"`
			Command    string `json:"command"`
		} `json:"meta"`
	}
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp))
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "session list", resp.Meta.Command)
	assert.NotEmpty(t, resp.Meta.Version)
	assert.GreaterOrEqual(t, resp.Meta.DurationMs, int64(0))
}

func TestEnvVarOutputJSON(t *testing.T) {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("binary not found at %s — run 'make build' first", binaryPath)
	}

	stateDir := t.TempDir()
	// Use env var instead of --output flag.
	stdout, _, exit := runEnvJSON(t, stateDir, "session", "list")
	require.Equal(t, 0, exit)

	var resp map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp))
	assert.Equal(t, "ok", resp["status"])
}

func TestErrorExitCodeJSONEnvelope(t *testing.T) {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("binary not found at %s — run 'make build' first", binaryPath)
	}

	stateDir := t.TempDir()
	_, stderr, exit := run(t, stateDir, "session", "stop", "--name", "nonexistent")
	assert.Equal(t, 3, exit, "not-found errors should exit 3")

	var errResp struct {
		Status string `json:"status"`
		Code   string `json:"code"`
		Meta   struct {
			Command string `json:"command"`
		} `json:"meta"`
	}
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stderr)), &errResp))
	assert.Equal(t, "error", errResp.Status)
	assert.Equal(t, "SESSION_NOT_FOUND", errResp.Code)
	assert.Equal(t, "session stop", errResp.Meta.Command)
}

// runEnvJSON executes the binary with SIMPLE_CLI_OUTPUT=json set via environment (not flag).
func runEnvJSON(t *testing.T, stateDir string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(),
		"SIMPLE_CLI_STATE_DIR="+stateDir,
		"SIMPLE_CLI_OUTPUT=json",
	)
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

