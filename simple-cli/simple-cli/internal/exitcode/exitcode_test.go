package exitcode_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/binpqh/simple-cli/internal/exitcode"
)

func TestExitErrorMessage(t *testing.T) {
	err := exitcode.New(exitcode.NotFound, errors.New("not found"))
	assert.Equal(t, "not found", err.Error())
	assert.Equal(t, exitcode.NotFound, err.Code)
}

func TestExitErrorNilWrapped(t *testing.T) {
	err := exitcode.New(exitcode.GeneralError, nil)
	assert.Equal(t, "exit code 1", err.Error())
}

func TestExitErrorUnwrap(t *testing.T) {
	inner := errors.New("inner")
	err := exitcode.New(exitcode.InvalidArgument, inner)
	assert.Equal(t, inner, errors.Unwrap(err))
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 0, exitcode.Success)
	assert.Equal(t, 1, exitcode.GeneralError)
	assert.Equal(t, 2, exitcode.InvalidArgument)
	assert.Equal(t, 3, exitcode.NotFound)
	assert.Equal(t, 4, exitcode.PermissionDenied)
	assert.Equal(t, 5, exitcode.Timeout)
}
