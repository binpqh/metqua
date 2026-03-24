// Package signals provides graceful-shutdown context helpers.
// Constitution Principle IX: long-running commands MUST implement graceful
// shutdown on SIGINT/SIGTERM and flush state within 5 seconds.
package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const drainTimeout = 5 * time.Second

// NotifyContext returns a context that is cancelled when SIGINT or SIGTERM is
// received. The returned cancel function MUST be deferred by the caller.
//
// After the signal fires the context has a hard 5-second drain deadline, after
// which the process may be forcefully terminated by the OS.
//
// Usage:
//
//	ctx, stop := signals.NotifyContext(context.Background())
//	defer stop()
func NotifyContext(parent context.Context) (context.Context, context.CancelFunc) {
	sigCtx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)

	drainCtx, cancel := context.WithCancelCause(sigCtx)

	go func() {
		// Wait for the signal context to be cancelled (signal received).
		<-sigCtx.Done()
		// Give the process up to drainTimeout to finish in-flight work.
		time.AfterFunc(drainTimeout, func() {
			cancel(context.DeadlineExceeded)
		})
	}()

	combined := func() {
		stop()
		cancel(nil)
	}
	return drainCtx, combined
}
