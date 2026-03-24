package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/your-org/simple-cli/internal/config"
	"github.com/your-org/simple-cli/internal/output"
	sess "github.com/your-org/simple-cli/internal/session"
)

func newStopCmd(store sess.SessionStore) *cobra.Command {
	var name, id string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a session",
		Long: `Sets the session status to stopped. State is retained on disk for inspection.
Use 'simple-cli session reset' to clear the state.`,
		Example: `  simple-cli session stop --name my-project
  simple-cli session stop --id 550e8400-e29b-41d4-a716-446655440000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			ctx := cmd.Context()

			cfg := ctx.Value(config.CtxKey{}).(*config.Config)
			w := output.NewWriter(cfg.Quiet)
			f := output.NewFormatter(cfg.Output, w, cfg.NoColor)

			if name == "" && id == "" {
				return fmt.Errorf("one of --name or --id is required")
			}

			var s *sess.Session
			var err error
			if name != "" {
				s, err = store.GetByName(ctx, name)
			} else {
				s, err = store.Get(ctx, id)
			}
			if err != nil {
				return f.FormatError("session stop", errorCode(err), err.Error(), hint(err), time.Since(start))
			}

			s.Status = sess.StatusStopped
			s.UpdatedAt = time.Now().UTC()
			if err := store.Update(ctx, s); err != nil {
				return f.FormatError("session stop", errorCode(err), err.Error(), hint(err), time.Since(start))
			}

			return f.FormatSuccess("session stop", s, time.Since(start))
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Session name")
	cmd.Flags().StringVar(&id, "id", "", "Session UUID")
	return cmd
}
