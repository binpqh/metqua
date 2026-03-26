package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/tokenstore"
)

func newStatusCmd() *cobra.Command {
	var providerName string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status for provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ctx.Value(config.CtxKey{}).(*config.Config)
			pc, err := cfg.ActiveProvider(providerName)
			if err != nil {
				return err
			}
			_ = config.ValidateProviderConfig(pc)
			store := tokenstore.NewFileTokenStore(tokenstore.PathForConfigDir(config.ConfigDir()))
			ts, err := store.Get(ctx, providerName)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Provider : %s\nStatus   : Not logged in\n", providerName)
				return nil
			}
			expired := ts.IsExpired()
			when := ts.Expiry.UTC().Format(time.RFC3339)
			if expired {
				fmt.Fprintf(cmd.OutOrStdout(), "Provider : %s\nStatus   : Token expired — run 'auth login' to refresh\nUser     : %s\nExpired  : %s\n", providerName, ts.UserID, when)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Provider : %s\nStatus   : Logged in\nUser     : %s\nExpires  : %s\n", providerName, ts.UserID, when)
			return nil
		},
	}
	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name from config (default: default_provider)")
	return cmd
}
