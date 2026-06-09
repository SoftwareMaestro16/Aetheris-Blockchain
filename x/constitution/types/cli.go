package types

import "github.com/spf13/cobra"

func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Constitution transaction commands", DisableFlagParsing: true, RunE: cobra.NoArgs}
	for _, use := range []string{"propose-amendment", "vote-amendment", "execute-amendment", "cancel-amendment"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Build " + use + " transaction", RunE: cobra.NoArgs})
	}
	return cmd
}

func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Constitution query commands", DisableFlagParsing: true, RunE: cobra.NoArgs}
	for _, use := range []string{"params", "constitution", "pending-amendments", "amendment", "protected-limits"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Run " + use + " query", RunE: cobra.NoArgs})
	}
	return cmd
}
