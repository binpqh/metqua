package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	authpkg "github.com/binpqh/simple-cli/internal/auth"
	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/provider"
	"github.com/binpqh/simple-cli/internal/tokenstore"
)

// providerRegistry maps provider names (or the fallback key "default") to adapter
// constructor functions. To add a new OAuth provider adapter, insert an entry here:
//
//	"myprovider": func(cfg *config.ProviderConfig) provider.ProviderAdapter { return myauth.New(cfg) },
//
// The "default" key is used for all providers not explicitly listed.
var providerRegistry = map[string]func(*config.ProviderConfig) provider.ProviderAdapter{
	"default": func(cfg *config.ProviderConfig) provider.ProviderAdapter {
		return authpkg.NewHTTPProviderAdapter(cfg)
	},
}

// resolveAdapter returns the registered adapter for the given provider name,
// falling back to "default" if no specific entry exists.
func resolveAdapter(name string, cfg *config.ProviderConfig) provider.ProviderAdapter {
	if fn, ok := providerRegistry[name]; ok {
		return fn(cfg)
	}
	return providerRegistry["default"](cfg)
}

func newLoginCmd() *cobra.Command {
	var providerName string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login using Device Authorization Flow",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ctx.Value(config.CtxKey{}).(*config.Config)
			// bind provider-specific env overrides to the (soon-to-be) active provider
			effectiveProvider := providerName
			if effectiveProvider == "" {
				effectiveProvider = cfg.DefaultProvider
			}
			_ = viper.BindEnv("providers."+effectiveProvider+".client_id", "SIMPLE_CLI_PROVIDER_CLIENT_ID")
			_ = viper.BindEnv("providers."+effectiveProvider+".device_endpoint", "SIMPLE_CLI_PROVIDER_DEVICE_ENDPOINT")
			_ = viper.BindEnv("providers."+effectiveProvider+".token_endpoint", "SIMPLE_CLI_PROVIDER_TOKEN_ENDPOINT")

			pc, err := cfg.ActiveProvider(providerName)
			if err != nil {
				return err
			}
			if err := config.ValidateProviderConfig(pc); err != nil {
				return err
			}
			// resolve adapter via registry (supports custom adapters)
			adapter := resolveAdapter(effectiveProvider, pc)
			store := tokenstore.NewFileTokenStore(tokenstore.PathForConfigDir(config.ConfigDir()))

			state, err := adapter.StartDeviceFlow(ctx)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Open this URL in your browser:\n\n  %s\n\nEnter code: %s\n\nWaiting for authorization...\n", state.VerificationURI, state.UserCode)

			tset, err := adapter.PollToken(ctx, state)
			if err != nil {
				return err
			}
			if tset == nil {
				return fmt.Errorf("no token received")
			}
			// set provider name and store
			tset.Provider = providerName
			if err := store.Set(ctx, providerName, tset); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "\n✓  Logged in to %s\n", providerName)
			return nil
		},
	}
	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name from config (default: default_provider)")
	return cmd
}
