//go:build integration && windows

package integration

import "os"

// sendShutdown kills the process on Windows (no SIGINT support for child processes).
func sendShutdown(p *os.Process) error {
	return p.Kill()
}
