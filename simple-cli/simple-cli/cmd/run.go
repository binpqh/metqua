// Package cmd — run command starts the long-running daemon process.
package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/output"
	"github.com/binpqh/simple-cli/internal/signals"
)

// newRunCmd returns the "run" sub-command.
// It blocks until SIGINT or SIGTERM is received, then exits cleanly within 5 s.
func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Start the long-running process (stays alive until device shutdown)",
		Long: `Starts simple-cli as a long-running daemon process.

The process blocks until it receives SIGINT or SIGTERM (e.g., system shutdown,
service manager stop, or Ctrl+C). Add your application logic below the TODO
comment — it will run inside a context that is cancelled on shutdown.`,
		Example: `  # Start the daemon
  simple-cli run

  # Start with JSON output (machine-readable shutdown envelope)
  simple-cli --output json run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			ctx, stop := signals.NotifyContext(cmd.Context())
			defer stop()

			cfg := ctx.Value(config.CtxKey{}).(*config.Config)
			w := output.NewWriter(cfg.Quiet)
			f := output.NewFormatter(cfg.Output, w, cfg.NoColor)

			slog.Info("simple-cli started", "pid", os.Getpid())

			// TODO: Add your application logic here. This context is cancelled on SIGINT/SIGTERM.
			// Example:
			//   go myWorker(ctx)   // starts a goroutine that respects ctx.Done()
			//   <-ctx.Done()       // blocks until shutdown signal
			<-ctx.Done()

			slog.Info("shutdown signal received, exiting cleanly")
			return f.FormatSuccess("run", map[string]any{
				"status":    "stopped",
				"uptime_ms": time.Since(start).Milliseconds(),
			}, time.Since(start))
		},
	}
}
