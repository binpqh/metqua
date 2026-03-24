//go:build integration

// Package integration runs black-box tests against the compiled simple-cli binary.
// Build: go test -tags integration ./tests/integration/... -binary /path/to/simple-cli
// The BINARY env var (or -binary flag) must point to the compiled binary.
package integration

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryPath string

func TestMain(m *testing.M) {
	flag.StringVar(&binaryPath, "binary", "", "Path to the simple-cli binary under test")
	flag.Parse()

	if binaryPath == "" {
		binaryPath = os.Getenv("SIMPLE_CLI_BINARY")
	}
	if binaryPath == "" {
		// Default to dist/ relative to the workspace root.
		_, thisFile, _, _ := runtime.Caller(0)
		root := filepath.Join(filepath.Dir(thisFile), "..", "..")
		if runtime.GOOS == "windows" {
			binaryPath = filepath.Join(root, "dist", "simple-cli.exe")
		} else {
			binaryPath = filepath.Join(root, "dist", "simple-cli")
		}
	}
	os.Exit(m.Run())
}

// run executes the binary with the given args plus a temporary state directory.
// Returns stdout, stderr, and the exit code.
func run(t *testing.T, stateDir string, args ...string) (string, string, int) {
	t.Helper()
	allArgs := append(args, "--output", "json")
	cmd := exec.Command(binaryPath, allArgs...)
	cmd.Env = append(os.Environ(), "SIMPLE_CLI_STATE_DIR="+stateDir)

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

func parseSuccessData(t *testing.T, stdout string) map[string]any {
	t.Helper()
	var resp struct {
		Status string         `json:"status"`
		Data   map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp))
	require.Equal(t, "ok", resp.Status)
	return resp.Data
}

func parseErrorCode(t *testing.T, stderr string) string {
	t.Helper()
	var resp struct {
		Status string `json:"status"`
		Code   string `json:"code"`
	}
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp))
	require.Equal(t, "error", resp.Status)
	return resp.Code
}

func TestSessionLifecycle(t *testing.T) {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("binary not found at %s — run 'make build' first", binaryPath)
	}

	stateDir := t.TempDir()
	name := "integration-test-" + time.Now().Format("150405")

	// Start.
	stdout, _, exit := run(t, stateDir, "session", "start", "--name", name)
	require.Equal(t, 0, exit)
	data := parseSuccessData(t, stdout)
	sessionID := data["id"].(string)
	assert.Equal(t, name, data["name"])
	assert.Equal(t, "active", data["status"])
	assert.NotEmpty(t, sessionID)

	// List — verify session exists.
	stdout, _, exit = run(t, stateDir, "session", "list")
	require.Equal(t, 0, exit)
	var listResp struct {
		Status string `json:"status"`
		Data   struct {
			Sessions []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"sessions"`
			Total int `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &listResp))
	require.Len(t, listResp.Data.Sessions, 1)
	assert.Equal(t, sessionID, listResp.Data.Sessions[0].ID)

	// Resume.
	stdout, _, exit = run(t, stateDir, "session", "resume", "--name", name)
	require.Equal(t, 0, exit)
	data = parseSuccessData(t, stdout)
	assert.Equal(t, "active", data["status"])

	// Stop.
	stdout, _, exit = run(t, stateDir, "session", "stop", "--name", name)
	require.Equal(t, 0, exit)
	data = parseSuccessData(t, stdout)
	assert.Equal(t, "stopped", data["status"])

	// Reset (requires --force to skip prompt in non-interactive mode).
	stdout, _, exit = run(t, stateDir, "session", "reset", "--name", name, "--force")
	require.Equal(t, 0, exit)
	data = parseSuccessData(t, stdout)
	assert.Equal(t, sessionID, data["old_id"])
	newID := data["id"].(string)
	assert.NotEqual(t, sessionID, newID, "reset should produce a new ID")
	assert.Equal(t, "active", data["status"])

	// Cleanup.
	run(t, stateDir, "session", "stop", "--name", name)
}

func TestSessionNotFound(t *testing.T) {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("binary not found at %s — run 'make build' first", binaryPath)
	}

	stateDir := t.TempDir()
	_, stderr, exit := run(t, stateDir, "session", "resume", "--name", "no-such-session")
	assert.Equal(t, 3, exit)
	code := parseErrorCode(t, stderr)
	assert.Equal(t, "SESSION_NOT_FOUND", code)
}

func TestSessionNameConflict(t *testing.T) {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("binary not found at %s — run 'make build' first", binaryPath)
	}

	stateDir := t.TempDir()
	name := "conflict-test"
	_, _, exit := run(t, stateDir, "session", "start", "--name", name)
	require.Equal(t, 0, exit)

	_, stderr, exit := run(t, stateDir, "session", "start", "--name", name)
	assert.NotEqual(t, 0, exit)
	code := parseErrorCode(t, stderr)
	assert.Equal(t, "SESSION_NAME_CONFLICT", code)
}
