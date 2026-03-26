// Package cmd — example_cmd demonstrates the sub-command extension pattern.
// Delete this file and replace it with your own commands when customising the template.
package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/output"
)

// newExampleCmd returns a minimal sub-command demonstrating how to add commands
// to the template. Delete this file and add your own in cmd/.
func newExampleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "example",
		Short: "Example sub-command — safe to delete when customising",
		Long: `A minimal example sub-command showing the extension pattern.

To create your own command:
  1. Copy this file to cmd/mycommand.go
  2. Rename newExampleCmd → newMyCmd and update Use/Short/Long
  3. Replace the RunE body with your logic
  4. Register it in cmd/root.go: rootCmd.AddCommand(newMyCmd())
  5. Delete this file`,
		Example: `  simple-cli example
  simple-cli --output json example`,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			cfg := cmd.Context().Value(config.CtxKey{}).(*config.Config)
			w := output.NewWriter(cfg.Quiet)
			f := output.NewFormatter(cfg.Output, w, cfg.NoColor)

			return f.FormatSuccess("example", map[string]any{
				"message": "replace this with your logic",
			}, time.Since(start))
		},
	}
}
