package session

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/your-org/simple-cli/internal/config"
	"github.com/your-org/simple-cli/internal/exitcode"
	"github.com/your-org/simple-cli/internal/output"
	sess "github.com/your-org/simple-cli/internal/session"
)

// nameRe validates session names: alphanumeric start, then alphanumeric/hyphen/underscore, max 64 chars.
var nameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

// adjectives and nouns for auto-generated session names.
var adjectives = []string{"bold", "swift", "quiet", "bright", "calm", "eager", "fierce"}
var nouns = []string{"river", "peak", "storm", "meadow", "canyon", "forge", "harbor"}

func generateName() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // non-crypto name gen
	return adjectives[r.Intn(len(adjectives))] + "-" + nouns[r.Intn(len(nouns))]
}

func newStartCmd(store sess.SessionStore) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new session",
		Long: `Creates a new persistent session that survives terminal restarts.
If --name is not provided, a unique adjective-noun name is auto-generated.`,
		Example: `  simple-cli session start
  simple-cli session start --name my-project
  simple-cli session start --name my-project --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			ctx := cmd.Context()

			cfg := ctx.Value(config.CtxKey{}).(*config.Config)
			w := output.NewWriter(cfg.Quiet)
			f := output.NewFormatter(cfg.Output, w, cfg.NoColor)

			if name == "" {
				name = generateName()
			}
			if !nameRe.MatchString(name) {
				return exitcode.New(exitcode.InvalidArgument,
					fmt.Errorf("invalid session name %q: must match ^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$", name))
			}

			s := &sess.Session{
				ID:        uuid.New().String(),
				Name:      name,
				Status:    sess.StatusActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				State:     map[string]any{},
			}

			if err := store.Create(ctx, s); err != nil {
				return f.FormatError("session start", errorCode(err), err.Error(), hint(err), time.Since(start))
			}

			return f.FormatSuccess("session start", s, time.Since(start))
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Session name (auto-generated if omitted)")
	return cmd
}

// errorCode converts a sentinel error to a stable SCREAMING_SNAKE_CASE code.
func errorCode(err error) string {
	switch {
	case isErr(err, sess.ErrNotFound):
		return "SESSION_NOT_FOUND"
	case isErr(err, sess.ErrNameConflict):
		return "SESSION_NAME_CONFLICT"
	case isErr(err, sess.ErrLockTimeout):
		return "SESSION_LOCK_TIMEOUT"
	case isErr(err, sess.ErrStoreReadOnly):
		return "STORE_READ_ONLY"
	case isErr(err, sess.ErrSessionStopped):
		return "SESSION_STOPPED"
	default:
		return "INTERNAL_ERROR"
	}
}

func hint(err error) string {
	switch {
	case isErr(err, sess.ErrNotFound):
		return "Use 'simple-cli session list' to see available sessions."
	case isErr(err, sess.ErrNameConflict):
		return "Choose a different name or use 'simple-cli session list' to see existing sessions."
	case isErr(err, sess.ErrSessionStopped):
		return "Use 'simple-cli session reset --name <name> --force' to start a fresh session with the same name."
	default:
		return ""
	}
}

func isErr(err, target error) bool {
	return err != nil && (err == target || unwrapContains(err, target))
}

func unwrapContains(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
