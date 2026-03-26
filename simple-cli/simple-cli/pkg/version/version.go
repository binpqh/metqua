// Package version provides the build-time-injectable version information
// for simple-cli. Variables are set via -ldflags at build time:
//
//	-X github.com/binpqh/simple-cli/pkg/version.Version=$(git describe --tags)
//	-X github.com/binpqh/simple-cli/pkg/version.Commit=$(git rev-parse --short HEAD)
//	-X github.com/binpqh/simple-cli/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)
package version

import "fmt"

// Version is the semantic version string, injected at build time.
// Default is the current release; override via -ldflags at build time.
var Version = "v2.1.0"

// Commit is the short git commit hash, injected at build time.
var Commit = "unknown"

// BuildDate is the UTC build timestamp in RFC3339 format, injected at build time.
var BuildDate = "unknown"

// String returns the formatted version line printed by --version.
func String() string {
	return fmt.Sprintf("simple-cli version %s (commit %s, built %s)", Version, Commit, BuildDate)
}
