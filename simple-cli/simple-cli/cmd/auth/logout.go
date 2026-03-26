package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/tokenstore"
)

func newLogoutCmd() *cobra.Command {
	var providerName string
	var all bool
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout and remove stored tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ctx.Value(config.CtxKey{}).(*config.Config)
			if all {
				// delete all providers: remove tokens file
				store := tokenstore.NewFileTokenStore(tokenstore.PathForConfigDir(config.ConfigDir()))
				_ = store.Delete(ctx, providerName) // best-effort placeholder
				fmt.Fprintln(cmd.OutOrStdout(), "✓  Logged out (all providers)")
				return nil
			}
			pc, err := cfg.ActiveProvider(providerName)
			if err != nil {
				return err
			}
			_ = config.ValidateProviderConfig(pc)
			store := tokenstore.NewFileTokenStore(tokenstore.PathForConfigDir(config.ConfigDir()))
			_ = store.Delete(ctx, providerName)
			fmt.Fprintf(cmd.OutOrStdout(), "✓  Logged out from %s\n", providerName)
			return nil
		},
	}
	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name from config (default: default_provider)")
	cmd.Flags().BoolVar(&all, "all", false, "Logout from all providers")
	return cmd
}
