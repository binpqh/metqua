//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunBlocksAndExitsOnSignal verifies the daemon blocks until a shutdown signal is sent
// and exits (cleanly on Unix; forcefully on Windows where Kill is used).
func TestRunBlocksAndExitsOnSignal(t *testing.T) {
	skipIfNoBinary(t)

	var outBuf, errBuf strings.Builder
	cmd := exec.Command(binaryPath, "run")
	cmd.Env = os.Environ()
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	require.NoError(t, cmd.Start())

	time.Sleep(150 * time.Millisecond)
	require.NoError(t, sendShutdown(cmd.Process))

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
		// On Unix: SIGINT triggers graceful exit → code 0.
		// On Windows: Kill() terminates immediately → code != 0. Both are acceptable.
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("process did not exit within 5 seconds after shutdown signal")
	}
}

// TestRunJSONEnvelope verifies that --output json produces a valid JSON envelope on shutdown.
// Requires graceful SIGINT — skipped on Windows where Kill() is used.
func TestRunJSONEnvelope(t *testing.T) {
	skipIfNoBinary(t)
	if runtime.GOOS == "windows" {
		t.Skip("graceful SIGINT not available for subprocess tests on Windows; Kill() does not trigger clean shutdown")
	}

	var outBuf, errBuf strings.Builder
	cmd := exec.Command(binaryPath, "--output", "json", "run")
	cmd.Env = os.Environ()
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	require.NoError(t, cmd.Start())

	time.Sleep(150 * time.Millisecond)
	require.NoError(t, sendShutdown(cmd.Process))

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		assert.NoError(t, err, "process should exit cleanly")
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("process did not exit within 5 seconds")
	}

	stdout := outBuf.String()
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Status   string `json:"status"`
			UptimeMs int64  `json:"uptime_ms"`
		} `json:"data"`
		Meta struct {
			Version    string `json:"version"`
			DurationMs int64  `json:"duration_ms"`
			Command    string `json:"command"`
		} `json:"meta"`
	}
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp),
		"stdout should be a valid JSON envelope, got: %q", stdout)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "stopped", resp.Data.Status)
	assert.GreaterOrEqual(t, resp.Data.UptimeMs, int64(0))
	assert.Equal(t, "run", resp.Meta.Command)
	assert.NotEmpty(t, resp.Meta.Version)
}

// TestRunHumanOutput verifies that human output writes "stopped" to stdout and
// shutdown info to stderr.
// Requires graceful SIGINT — skipped on Windows where Kill() is used.
func TestRunHumanOutput(t *testing.T) {
	skipIfNoBinary(t)
	if runtime.GOOS == "windows" {
		t.Skip("graceful SIGINT not available for subprocess tests on Windows; Kill() does not trigger clean shutdown")
	}

	var outBuf, errBuf strings.Builder
	cmd := exec.Command(binaryPath, "run")
	cmd.Env = os.Environ()
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	require.NoError(t, cmd.Start())

	time.Sleep(150 * time.Millisecond)
	require.NoError(t, sendShutdown(cmd.Process))

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		assert.NoError(t, err, "process should exit cleanly")
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("process did not exit within 5 seconds")
	}

	stdout := outBuf.String()
	stderr := errBuf.String()
	assert.NotEmpty(t, stdout, "human output should be written to stdout")
	assert.Contains(t, strings.ToLower(stdout), "stopped",
		"stdout should mention 'stopped'")
	assert.Contains(t, strings.ToLower(stderr), "shutdown",
		"stderr should mention shutdown signal")
}
