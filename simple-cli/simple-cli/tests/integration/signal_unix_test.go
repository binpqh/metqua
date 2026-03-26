//go:build integration && !windows

package integration

import (
	"os"
	"syscall"
)

// sendShutdown sends SIGINT to the process to trigger a clean shutdown.
func sendShutdown(p *os.Process) error {
	return p.Signal(syscall.SIGINT)
}
