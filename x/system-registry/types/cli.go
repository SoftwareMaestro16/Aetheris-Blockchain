package types

import "github.com/spf13/cobra"

func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "System registry transaction commands", DisableFlagParsing: true, RunE: cobra.NoArgs}
	for _, use := range []string{"register-entity", "update-entity", "pause-entity", "resume-entity", "deprecate-entity"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Build " + use + " transaction", RunE: cobra.NoArgs})
	}
	return cmd
}

func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{Use: ModuleName, Short: "System registry query commands", DisableFlagParsing: true, RunE: cobra.NoArgs}
	for _, use := range []string{"params", "entities", "entity", "reserved-addresses", "system-address", "dependency-graph"} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Run " + use + " query", RunE: cobra.NoArgs})
	}
	return cmd
}
