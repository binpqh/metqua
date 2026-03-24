// Package session implements the `simple-cli session` sub-command group.
package session

import (
	"github.com/spf13/cobra"
	sess "github.com/your-org/simple-cli/internal/session"
)

// NewSessionCmd builds the `session` command group and attaches all sub-commands.
func NewSessionCmd(store sess.SessionStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage long-life sessions",
		Long: `Create, resume, list, stop, and reset persistent sessions.
Sessions survive terminal restarts and can be resumed at any time.`,
		Example: `  simple-cli session start --name my-project
  simple-cli session resume --name my-project
  simple-cli session list
  simple-cli session stop --name my-project
  simple-cli session reset --name my-project --force`,
	}

	cmd.AddCommand(
		newStartCmd(store),
		newResumeCmd(store),
		newListCmd(store),
		newStopCmd(store),
		newResetCmd(store),
	)

	return cmd
}
