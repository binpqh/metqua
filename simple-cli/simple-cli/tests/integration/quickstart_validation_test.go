//go:build integration

// Package integration validates the quickstart.md end-to-end against the built binary.
package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuickstartValidation executes every command from docs/quickstart.md in
// sequence and asserts expected exit codes and output patterns.
// Run with: go test -tags integration ./tests/integration/... -run TestQuickstartValidation
func TestQuickstartValidation(t *testing.T) {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("binary not found at %s — run 'make build' first", binaryPath)
	}

	stateDir := t.TempDir()

	// Step 1: simple-cli --version
	t.Run("version", func(t *testing.T) {
		stdout, _, exit := runBare(t, "--version")
		require.Equal(t, 0, exit)
		assert.Contains(t, stdout, "simple-cli")
	})

	// Step 2: session start
	var sessionID string
	t.Run("session_start", func(t *testing.T) {
		stdout, _, exit := run(t, stateDir, "session", "start", "--name", "qs-demo")
		require.Equal(t, 0, exit)
		data := parseSuccessData(t, stdout)
		assert.Equal(t, "qs-demo", data["name"])
		assert.Equal(t, "active", data["status"])
		sessionID = data["id"].(string)
		assert.NotEmpty(t, sessionID)
	})

	// Step 3: session list
	t.Run("session_list", func(t *testing.T) {
		_, _, exit := run(t, stateDir, "session", "list")
		assert.Equal(t, 0, exit)
	})

	// Step 4: session resume (simulates terminal restart by simply calling resume)
	t.Run("session_resume", func(t *testing.T) {
		stdout, _, exit := run(t, stateDir, "session", "resume", "--name", "qs-demo")
		require.Equal(t, 0, exit)
		data := parseSuccessData(t, stdout)
		assert.Equal(t, "active", data["status"])
	})

	// Step 5: session stop
	t.Run("session_stop", func(t *testing.T) {
		stdout, _, exit := run(t, stateDir, "session", "stop", "--name", "qs-demo")
		require.Equal(t, 0, exit)
		data := parseSuccessData(t, stdout)
		assert.Equal(t, "stopped", data["status"])
	})

	// Step 6: session reset --force
	t.Run("session_reset", func(t *testing.T) {
		stdout, _, exit := run(t, stateDir, "session", "reset", "--name", "qs-demo", "--force")
		require.Equal(t, 0, exit)
		data := parseSuccessData(t, stdout)
		assert.Equal(t, sessionID, data["old_id"])
		assert.NotEqual(t, sessionID, data["id"])
		assert.Equal(t, "active", data["status"])
	})
}

// runBare executes the binary with the given args only (no --output json, no state dir).
func runBare(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
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

