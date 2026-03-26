package output_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/binpqh/simple-cli/internal/output"
)

func TestWriterWriteOut(t *testing.T) {
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	w := &output.Writer{Out: outBuf, Err: errBuf, Quiet: false}

	n, err := w.WriteOut([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", outBuf.String())
	assert.Empty(t, errBuf.String())
}

func TestWriterWriteOutQuiet(t *testing.T) {
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	w := &output.Writer{Out: outBuf, Err: errBuf, Quiet: true}

	n, err := w.WriteOut([]byte("suppressed"))
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Empty(t, outBuf.String())
}

func TestWriterWriteErrAlwaysWrites(t *testing.T) {
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	w := &output.Writer{Out: outBuf, Err: errBuf, Quiet: true}

	_, err := w.WriteErr([]byte("error message"))
	assert.NoError(t, err)
	assert.Equal(t, "error message", errBuf.String())
}

func TestNewWriter(t *testing.T) {
	w := output.NewWriter(false)
	assert.NotNil(t, w.Out)
	assert.NotNil(t, w.Err)
	assert.False(t, w.Quiet)
}
