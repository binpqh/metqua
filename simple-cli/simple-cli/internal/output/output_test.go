package output_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/binpqh/simple-cli/internal/output"
)

func newTestWriter(quiet bool) (*output.Writer, *bytes.Buffer, *bytes.Buffer) {
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	w := &output.Writer{Out: outBuf, Err: errBuf, Quiet: quiet}
	return w, outBuf, errBuf
}

// --- JSONFormatter ---

func TestJSONFormatterSuccess(t *testing.T) {
	w, outBuf, _ := newTestWriter(false)
	f := output.NewJSONFormatter(w)

	data := map[string]string{"id": "123", "name": "test"}
	require.NoError(t, f.FormatSuccess("session start", data, 10*time.Millisecond))

	var resp output.SuccessResponse
	require.NoError(t, json.NewDecoder(outBuf).Decode(&resp))
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "session start", resp.Meta.Command)
	assert.GreaterOrEqual(t, resp.Meta.DurationMs, int64(0))
}

func TestJSONFormatterError(t *testing.T) {
	w, _, errBuf := newTestWriter(false)
	f := output.NewJSONFormatter(w)

	err := f.FormatError("session resume", "SESSION_NOT_FOUND", "session 'x' not found", "check list", 5*time.Millisecond)
	require.Error(t, err) // FormatError always returns an error with exit code.

	var resp output.ErrorResponse
	require.NoError(t, json.NewDecoder(errBuf).Decode(&resp))
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "SESSION_NOT_FOUND", resp.Code)
	assert.Equal(t, "check list", resp.Hint)
}

// --- HumanFormatter ---

func TestHumanFormatterSuccess(t *testing.T) {
	w, outBuf, _ := newTestWriter(false)
	f := output.NewHumanFormatter(w, false)

	require.NoError(t, f.FormatSuccess("test", "hello world", time.Millisecond))
	assert.Contains(t, outBuf.String(), "hello world")
}

func TestHumanFormatterError(t *testing.T) {
	w, _, errBuf := newTestWriter(false)
	f := output.NewHumanFormatter(w, false)

	err := f.FormatError("test", "INTERNAL_ERROR", "something broke", "try again", time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, errBuf.String(), "something broke")
	assert.Contains(t, errBuf.String(), "try again")
}

func TestHumanFormatterNoColor(t *testing.T) {
	w, outBuf, _ := newTestWriter(false)
	f := output.NewHumanFormatter(w, true)

	coloredMsg := "\x1b[32mgreen text\x1b[0m"
	require.NoError(t, f.FormatSuccess("test", coloredMsg, time.Millisecond))
	assert.NotContains(t, outBuf.String(), "\x1b[")
	assert.Contains(t, outBuf.String(), "green text")
}

func TestHumanFormatterQuiet(t *testing.T) {
	w, outBuf, _ := newTestWriter(true)
	f := output.NewHumanFormatter(w, false)

	require.NoError(t, f.FormatSuccess("test", "should be suppressed", time.Millisecond))
	assert.Empty(t, outBuf.String())
}

func TestNewFormatterReturnsJSON(t *testing.T) {
	w, outBuf, _ := newTestWriter(false)
	f := output.NewFormatter("json", w, false)
	require.NoError(t, f.FormatSuccess("cmd", "data", time.Millisecond))

	var resp output.SuccessResponse
	require.NoError(t, json.NewDecoder(outBuf).Decode(&resp))
	assert.Equal(t, "ok", resp.Status)
}

func TestNewFormatterReturnsHuman(t *testing.T) {
	w, outBuf, _ := newTestWriter(false)
	f := output.NewFormatter("human", w, false)
	require.NoError(t, f.FormatSuccess("cmd", "hello!", time.Millisecond))
	assert.Contains(t, outBuf.String(), "hello!")
}
