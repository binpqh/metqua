// Package cmd is the entry-point package for all simple-cli commands.
// Constitution Principle II: no init() except flag registration in cmd/ packages.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sessioncmd "github.com/your-org/simple-cli/cmd/session"
	"github.com/your-org/simple-cli/internal/config"
	"github.com/your-org/simple-cli/internal/exitcode"
	"github.com/your-org/simple-cli/internal/session"
	"github.com/your-org/simple-cli/pkg/version"
)

var (
	cfgFile string
	v       = viper.New()
	cfg     *config.Config

	// sessionStore is created once during Execute and shared across commands.
	sessionStore session.SessionStore
)

// rootCmd is the base command. Every sub-command is added as a child.
var rootCmd = &cobra.Command{
	Use:   "simple-cli",
	Short: "A cross-platform CLI for managing long-life sessions",
	Long: `simple-cli maintains persistent sessions across terminal restarts.
Sessions survive shell closes and can be resumed at any time.

Use --output json for machine-readable output suitable for AI agent workflows.`,
	Example: `  # Start a new session
  simple-cli session start --name my-project

  # Resume after terminal restart
  simple-cli session resume --name my-project

  # List all sessions in JSON mode (for AI agents)
  simple-cli --output json session list`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load config file if one was specified or auto-discovered.
		if cfgFile != "" {
			v.SetConfigFile(cfgFile)
		} else {
			v.AddConfigPath(config.ConfigDir())
			v.SetConfigName("config")
			v.SetConfigType("yaml")
		}
		// Silently ignore "file not found" — config file is optional.
		_ = v.ReadInConfig()

		// Bind env vars (Principle IV / Constitution §Config).
		v.SetEnvPrefix("SIMPLE_CLI")
		_ = v.BindEnv("output", "SIMPLE_CLI_OUTPUT")
		_ = v.BindEnv("log_level", "SIMPLE_CLI_LOG_LEVEL")
		_ = v.BindEnv("no_color", "NO_COLOR")
		_ = v.BindEnv("state_dir", "SIMPLE_CLI_STATE_DIR")
		v.AutomaticEnv()

		// Bind persistent flags into Viper.
		_ = v.BindPFlag("output", cmd.Root().PersistentFlags().Lookup("output"))
		_ = v.BindPFlag("log_level", cmd.Root().PersistentFlags().Lookup("log-level"))
		_ = v.BindPFlag("no_color", cmd.Root().PersistentFlags().Lookup("no-color"))
		_ = v.BindPFlag("quiet", cmd.Root().PersistentFlags().Lookup("quiet"))

		var err error
		cfg, err = config.Load(v)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Initialise slog — all output to stderr (Principle VI).
		var handler slog.Handler
		var levelVar slog.LevelVar
		switch cfg.LogLevel {
		case "debug":
			levelVar.Set(slog.LevelDebug)
		case "warn":
			levelVar.Set(slog.LevelWarn)
		case "error":
			levelVar.Set(slog.LevelError)
		default:
			levelVar.Set(slog.LevelInfo)
		}
		opts := &slog.HandlerOptions{Level: &levelVar}
		if cfg.Output == "json" {
			handler = slog.NewJSONHandler(os.Stderr, opts)
		} else {
			handler = slog.NewTextHandler(os.Stderr, opts)
		}
		slog.SetDefault(slog.New(handler))

		// Store resolved config in context so sub-commands can read it.
		cmd.SetContext(context.WithValue(cmd.Context(), config.CtxKey{}, cfg))

		// Initialise the session store once (idempotent on subsequent calls).
		if sessionStore == nil {
			fs, err := session.NewFileStore(cfg.StateDir)
			if err != nil {
				slog.Warn("falling back to in-memory session store", "err", err)
				sessionStore = session.NewMemStore()
			} else {
				sessionStore = fs
			}
		}

		return nil
	},
}

// Execute runs the root command. Called from main.
func Execute() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		var ee *exitcode.ExitError
		if errors.As(err, &ee) {
			os.Exit(ee.Code)
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(exitcode.GeneralError)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&cfgFile, "config", "", "Config file path (default: $XDG_CONFIG_HOME/simple-cli/config.yaml)")
	pf.String("output", "human", "Output format: human or json (env: SIMPLE_CLI_OUTPUT)")
	pf.String("log-level", "info", "Log verbosity: debug, info, warn, error (env: SIMPLE_CLI_LOG_LEVEL)")
	pf.Bool("no-color", false, "Suppress ANSI escape codes (env: NO_COLOR)")
	pf.BoolP("quiet", "q", false, "Suppress all informational output")

	rootCmd.Version = version.String()
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	// sessionStore is nil until PersistentPreRunE runs, so we pass a proxy
	// store that delegates to the real store at runtime.
	proxy := session.NewProxyStore(&sessionStore)
	rootCmd.AddCommand(sessioncmd.NewSessionCmd(proxy))
}
