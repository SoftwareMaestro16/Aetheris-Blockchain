package types

import "github.com/spf13/cobra"

func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Aetra staking policy transaction commands", RunE: cobra.NoArgs}
	for _, use := range []string{"update-params", "register-validator-identity", "acknowledge-concentration-warning"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Build " + use + " transaction", RunE: cobra.NoArgs})
	}
	return cmd
}

func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Aetra staking policy query commands", RunE: cobra.NoArgs}
	for _, use := range []string{"params", "validator-effective-power", "validator-stake", "top-n-concentration", "validator-reward-multiplier", "delegation-warning-status"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Run " + use + " query", RunE: cobra.NoArgs})
	}
	return cmd
}
