// Package output provides the Writer type that routes payload to stdout and
// diagnostics to stderr. Constitution Principle IV: stdout MUST contain only
// payload data; all errors and logs MUST go to stderr.
package output

import (
	"io"
	"os"
)

// Writer holds separated output destinations.
type Writer struct {
	Out   io.Writer
	Err   io.Writer
	Quiet bool
}

// NewWriter returns a Writer using os.Stdout and os.Stderr.
func NewWriter(quiet bool) *Writer {
	return &Writer{Out: os.Stdout, Err: os.Stderr, Quiet: quiet}
}

// WriteOut writes p to Out. When Quiet is true informational writes are
// suppressed; callers that must always write (e.g. primary data) should call
// Out.Write directly.
func (w *Writer) WriteOut(p []byte) (int, error) {
	if w.Quiet {
		return len(p), nil
	}
	return w.Out.Write(p)
}

// WriteErr writes p to Err regardless of Quiet mode — errors always surface.
func (w *Writer) WriteErr(p []byte) (int, error) {
	return w.Err.Write(p)
}
