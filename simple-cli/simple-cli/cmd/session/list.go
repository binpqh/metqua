package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/your-org/simple-cli/internal/config"
	"github.com/your-org/simple-cli/internal/output"
	sess "github.com/your-org/simple-cli/internal/session"
)

func newListCmd(store sess.SessionStore) *cobra.Command {
	var statusFilter string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Long:  `Lists all sessions. Use --status to filter by lifecycle status.`,
		Example: `  simple-cli session list
  simple-cli session list --status active
  simple-cli session list --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			ctx := cmd.Context()

			cfg := ctx.Value(config.CtxKey{}).(*config.Config)
			w := output.NewWriter(cfg.Quiet)
			f := output.NewFormatter(cfg.Output, w, cfg.NoColor)

			var filter *sess.SessionStatus
			if statusFilter != "" {
				s := sess.SessionStatus(statusFilter)
				switch s {
				case sess.StatusActive, sess.StatusPaused, sess.StatusStopped:
					filter = &s
				default:
					return fmt.Errorf("invalid status %q: must be active, paused, or stopped", statusFilter)
				}
			}

			sessions, err := store.List(ctx, filter)
			if err != nil {
				return f.FormatError("session list", errorCode(err), err.Error(), hint(err), time.Since(start))
			}

			// For JSON mode, pass the raw slice.
			// For human mode, format as a table.
			if cfg.Output == "json" {
				type listData struct {
					Sessions []*sess.Session `json:"sessions"`
					Total    int             `json:"total"`
				}
				return f.FormatSuccess("session list",
					listData{Sessions: sessions, Total: len(sessions)}, time.Since(start))
			}

			// Human table.
			if len(sessions) == 0 {
				return f.FormatSuccess("session list", "No sessions found.", time.Since(start))
			}
			header := fmt.Sprintf("%-10s %-24s %-10s %s", "ID", "NAME", "STATUS", "CREATED")
			rows := header + "\n"
			for _, s := range sessions {
				shortID := s.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				rows += fmt.Sprintf("%-10s %-24s %-10s %s\n",
					shortID, s.Name, s.Status, s.CreatedAt.Format(time.RFC3339))
			}
			return f.FormatSuccess("session list", rows, time.Since(start))
		},
	}

	cmd.Flags().StringVarP(&statusFilter, "status", "s", "", "Filter by status: active, paused, stopped")
	return cmd
}
