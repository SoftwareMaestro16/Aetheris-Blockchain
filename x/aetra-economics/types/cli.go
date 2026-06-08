package types

import "github.com/spf13/cobra"

func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Aetra economics transaction commands", RunE: cobra.NoArgs}
	for _, use := range []string{"update-params", "apply-epoch"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Build " + use + " transaction", RunE: cobra.NoArgs})
	}
	return cmd
}

func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Aetra economics query commands", RunE: cobra.NoArgs}
	for _, use := range []string{"current-inflation", "current-bonded-ratio", "estimated-apr", "fee-split-params", "burned-supply", "treasury-balance", "epoch-reward-summary"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Run " + use + " query", RunE: cobra.NoArgs})
	}
	return cmd
}
