package auth

import (
	"encoding/json"
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
			store := tokenstore.NewFileTokenStore(tokenstore.PathForConfigDir(config.ConfigDir()))
			if all {
				_ = store.Delete(ctx, providerName)
				if cfg.Output == "json" {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]interface{}{
						"provider": "all", "success": true, "message": "Logged out (all providers)",
					})
				}
				fmt.Fprintln(cmd.OutOrStdout(), "✓  Logged out (all providers)")
				return nil
			}
			pc, err := cfg.ActiveProvider(providerName)
			if err != nil {
				return err
			}
			_ = config.ValidateProviderConfig(pc, cfg.Insecure)
			_ = store.Delete(ctx, providerName)
			if cfg.Output == "json" {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]interface{}{
					"provider": providerName, "success": true, "message": "Logged out successfully",
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓  Logged out from %s\n", providerName)
			return nil
		},
	}
	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name from config (default: default_provider)")
	cmd.Flags().BoolVar(&all, "all", false, "Logout from all providers")
	return cmd
}
