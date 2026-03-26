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

	authcmd "github.com/binpqh/simple-cli/cmd/auth"
	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/exitcode"
	"github.com/binpqh/simple-cli/pkg/version"
)

var (
	cfgFile string
	v       = viper.New()
	cfg     *config.Config
)

// rootCmd is the base command. Every sub-command is added as a child.
var rootCmd = &cobra.Command{
	Use:   "simple-cli",
	Short: "A cross-platform CLI template",
	Long: `simple-cli is a cross-platform CLI template that stays alive as a
long-running process until the device shuts down.

Customise it by adding sub-commands in cmd/ and implementing your logic
inside internal/.

Use --output json for machine-readable output suitable for AI agent workflows.`,
	Example: `  # Start the long-running process
  simple-cli run

  # Run with JSON output (for AI agents / scripts)
  simple-cli --output json run`,

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

	// To add a new sub-command, create cmd/mycommand.go with a newMyCmd() constructor
	// following the same pattern as cmd/run.go, then register it here:
	//   rootCmd.AddCommand(newMyCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newExampleCmd())
	rootCmd.AddCommand(authcmd.NewAuthCmd())
	rootCmd.AddCommand(newChatCmd())
}
