package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/your-org/simple-cli/internal/config"
	"github.com/your-org/simple-cli/internal/output"
	sess "github.com/your-org/simple-cli/internal/session"
)

func newResumeCmd(store sess.SessionStore) *cobra.Command {
	var name, id string

	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume an existing session",
		Long: `Loads an existing session by name or ID and marks it as active.
At least one of --name or --id must be provided.`,
		Example: `  simple-cli session resume --name my-project
  simple-cli session resume --id 550e8400-e29b-41d4-a716-446655440000`,
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
				return f.FormatError("session resume", errorCode(err), err.Error(), hint(err), time.Since(start))
			}

			if s.Status == sess.StatusStopped {
				return f.FormatError("session resume", "SESSION_STOPPED",
					ErrStopped(s.Name).Error(), hint(sess.ErrSessionStopped), time.Since(start))
			}

			s.Status = sess.StatusActive
			s.UpdatedAt = time.Now().UTC()

			if err := store.Update(ctx, s); err != nil {
				return f.FormatError("session resume", errorCode(err), err.Error(), hint(err), time.Since(start))
			}

			return f.FormatSuccess("session resume", s, time.Since(start))
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Session name")
	cmd.Flags().StringVar(&id, "id", "", "Session UUID")
	return cmd
}

// ErrStopped builds a readable error for a stopped session.
func ErrStopped(name string) error {
	return fmt.Errorf("session %q is stopped and cannot be resumed directly", name)
}
