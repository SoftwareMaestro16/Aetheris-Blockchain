package types

import "github.com/spf13/cobra"

func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Aetra validator score transaction commands", RunE: cobra.NoArgs}
	for _, use := range []string{"update-params", "update-scores"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Build " + use + " transaction", RunE: cobra.NoArgs})
	}
	return cmd
}

func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Aetra validator score query commands", RunE: cobra.NoArgs}
	for _, use := range []string{"params", "validator-score", "public-validator-metrics", "all-validator-scores"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Run " + use + " query", RunE: cobra.NoArgs})
	}
	return cmd
}
