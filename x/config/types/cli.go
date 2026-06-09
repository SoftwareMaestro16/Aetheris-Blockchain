package types

import "github.com/spf13/cobra"

func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Config transaction commands", DisableFlagParsing: true, RunE: cobra.NoArgs}
	for _, use := range []string{"submit-change", "approve-change", "reject-change", "execute-change", "cancel-change"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Build " + use + " transaction", RunE: cobra.NoArgs})
	}
	return cmd
}

func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "Config query commands", DisableFlagParsing: true, RunE: cobra.NoArgs}
	for _, use := range []string{"params", "entries", "entry", "pending-changes", "change", "effective-params"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Run " + use + " query", RunE: cobra.NoArgs})
	}
	return cmd
}
