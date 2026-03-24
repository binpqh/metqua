package session

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/your-org/simple-cli/internal/config"
	"github.com/your-org/simple-cli/internal/output"
	sess "github.com/your-org/simple-cli/internal/session"
)

func newResetCmd(store sess.SessionStore) *cobra.Command {
	var name, id string
	var force bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset (delete + recreate) a session",
		Long: `Deletes the session and creates a new one with the same name and a fresh state.
Requires --force to skip the confirmation prompt in interactive mode.
In --output json mode, --force is assumed.`,
		Example: `  simple-cli session reset --name my-project --force
  simple-cli session reset --id 550e8400-... --force`,
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
				return f.FormatError("session reset", errorCode(err), err.Error(), hint(err), time.Since(start))
			}

			// In human mode require --force or prompt confirmation.
			if cfg.Output != "json" && !force {
				fmt.Fprintf(os.Stderr, "Reset session %q? All state will be lost. [y/N] ", s.Name)
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					fmt.Fprintln(os.Stderr, "Aborted.")
					return nil
				}
			}

			oldID := s.ID
			newName := s.Name

			if err := store.Delete(ctx, oldID); err != nil {
				return f.FormatError("session reset", errorCode(err), err.Error(), hint(err), time.Since(start))
			}

			newSession := &sess.Session{
				ID:        uuid.New().String(),
				Name:      newName,
				Status:    sess.StatusActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				State:     map[string]any{},
			}
			if err := store.Create(ctx, newSession); err != nil {
				return f.FormatError("session reset", errorCode(err), err.Error(), hint(err), time.Since(start))
			}

			type resetData struct {
				OldID     string    `json:"old_id"`
				ID        string    `json:"id"`
				Name      string    `json:"name"`
				Status    string    `json:"status"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
				State     map[string]any `json:"state"`
			}
			data := resetData{
				OldID:     oldID,
				ID:        newSession.ID,
				Name:      newSession.Name,
				Status:    string(newSession.Status),
				CreatedAt: newSession.CreatedAt,
				UpdatedAt: newSession.UpdatedAt,
				State:     newSession.State,
			}
			return f.FormatSuccess("session reset", data, time.Since(start))
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Session name")
	cmd.Flags().StringVar(&id, "id", "", "Session UUID")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
