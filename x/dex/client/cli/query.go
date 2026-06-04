package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sovereign-l1/l1/x/dex/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "DEX queries",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(NewPoolCmd(), NewPoolByPairCmd(), NewPoolsCmd())
	return cmd
}

func NewPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [pool-id]",
		Short: "Query a pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			poolID, err := parseUint(args[0])
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).Pool(cmd.Context(), &types.QueryPoolRequest{PoolId: poolID})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func NewPoolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Query pools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			pageReq, err := client.ReadPageRequest(client.MustFlagSetWithPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).Pools(cmd.Context(), &types.QueryPoolsRequest{Pagination: pageReq})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "dex pools")
	return cmd
}

func NewPoolByPairCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool-by-pair [denom-a] [denom-b]",
		Short: "Query a pool by token pair",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).PoolByPair(cmd.Context(), &types.QueryPoolByPairRequest{
				DenomA: args[0],
				DenomB: args[1],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func parseUint(value string) (uint64, error) {
	return strconv.ParseUint(value, 10, 64)
}
