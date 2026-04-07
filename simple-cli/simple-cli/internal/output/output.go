// Package output provides Formatter implementations for human-readable and
// machine-readable (JSON) output. Constitution Principle IV: every command
// MUST support --output json for AI agent interoperability.
package output

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/binpqh/simple-cli/internal/exitcode"
	"github.com/binpqh/simple-cli/pkg/version"
)

// ansiRe matches ANSI escape sequences for stripping in --no-color mode.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Meta holds per-response metadata included in every JSON envelope.
type Meta struct {
	Version    string `json:"version"`
	DurationMs int64  `json:"duration_ms"`
	Command    string `json:"command"`
}

// SuccessResponse is the JSON envelope for a successful command execution.
type SuccessResponse struct {
	Status string `json:"status"` // always "ok"
	Data   any    `json:"data"`
	Meta   Meta   `json:"meta"`
}

// ErrorResponse is the JSON envelope for a failed command execution.
type ErrorResponse struct {
	Status  string `json:"status"` // always "error"
	Code    string `json:"code"`   // SCREAMING_SNAKE_CASE stable string
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
	Meta    Meta   `json:"meta"`
}

// Formatter is the interface both HumanFormatter and JSONFormatter satisfy.
type Formatter interface {
	// FormatSuccess writes a success response containing data to the writer's Out.
	FormatSuccess(command string, data any, elapsed time.Duration) error
	// FormatError writes an error response to the writer's Err.
	FormatError(command, code, message, hint string, elapsed time.Duration) error
}

// HumanFormatter prints plain human-readable text.
type HumanFormatter struct {
	w       *Writer
	noColor bool
}

// NewHumanFormatter creates a HumanFormatter.
func NewHumanFormatter(w *Writer, noColor bool) *HumanFormatter {
	return &HumanFormatter{w: w, noColor: noColor}
}

// FormatSuccess writes msg as a plain text line to stdout.
// The data is expected to already be formatted as a human-readable string
// by the caller; this method just flushes it.
func (f *HumanFormatter) FormatSuccess(_ string, data any, _ time.Duration) error {
	s := fmt.Sprintf("%v\n", data)
	if f.noColor {
		s = ansiRe.ReplaceAllString(s, "")
	}
	_, err := f.w.WriteOut([]byte(s))
	return err
}

// FormatError writes a plain text error line to stderr.
func (f *HumanFormatter) FormatError(_, code, message, hint string, _ time.Duration) error {
	s := "Error: " + message
	if hint != "" {
		s += "\nHint: " + hint
	}
	f.w.WriteErr([]byte(s + "\n")) //nolint:errcheck // stderr write error is unrecoverable
	return exitcode.New(exitCodeFor(code), fmt.Errorf("%s", message))
}

// JSONFormatter writes structured JSON envelopes.
type JSONFormatter struct {
	w *Writer
}

// NewJSONFormatter creates a JSONFormatter.
func NewJSONFormatter(w *Writer) *JSONFormatter {
	return &JSONFormatter{w: w}
}

// FormatSuccess marshals a SuccessResponse to stdout.
func (f *JSONFormatter) FormatSuccess(command string, data any, elapsed time.Duration) error {
	resp := SuccessResponse{
		Status: "ok",
		Data:   data,
		Meta: Meta{
			Version:    version.Version,
			DurationMs: elapsed.Milliseconds(),
			Command:    command,
		},
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("output: marshal success: %w", err)
	}
	_, err = f.w.Out.Write(append(b, '\n'))
	return err
}

// FormatError marshals an ErrorResponse to stderr.
func (f *JSONFormatter) FormatError(command, code, message, hint string, elapsed time.Duration) error {
	resp := ErrorResponse{
		Status:  "error",
		Code:    code,
		Message: message,
		Hint:    hint,
		Meta: Meta{
			Version:    version.Version,
			DurationMs: elapsed.Milliseconds(),
			Command:    command,
		},
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("output: marshal error: %w", err)
	}
	f.w.Err.Write(append(b, '\n')) //nolint:errcheck // stderr write error is unrecoverable
	return exitcode.New(exitCodeFor(code), fmt.Errorf("%s", message))
}

// exitCodeFor maps a stable error code string to a process exit code per
// contracts/exit-codes.md.
func exitCodeFor(code string) int {
	switch code {
	case "SESSION_NOT_FOUND":
		return exitcode.NotFound
	case "INVALID_ARGUMENT":
		return exitcode.InvalidArgument
	case "SESSION_LOCK_TIMEOUT", "CONTEXT_DEADLINE_EXCEEDED":
		return exitcode.Timeout
	case "STORE_READ_ONLY":
		return exitcode.PermissionDenied
	default:
		return exitcode.GeneralError
	}
}

// NewFormatter returns the correct Formatter based on the output mode setting.
func NewFormatter(mode string, w *Writer, noColor bool) Formatter {
	if mode == "json" {
		return NewJSONFormatter(w)
	}
	return NewHumanFormatter(w, noColor)
}
