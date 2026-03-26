package version_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/binpqh/simple-cli/pkg/version"
)

func TestStringContainsVersion(t *testing.T) {
	s := version.String()
	assert.True(t, strings.HasPrefix(s, "simple-cli version "),
		"version string should start with 'simple-cli version ', got: %q", s)
}

func TestStringContainsCommit(t *testing.T) {
	s := version.String()
	assert.Contains(t, s, "commit", "version string should contain 'commit'")
}

func TestStringContainsBuildDate(t *testing.T) {
	s := version.String()
	assert.Contains(t, s, "built", "version string should contain 'built'")
}
