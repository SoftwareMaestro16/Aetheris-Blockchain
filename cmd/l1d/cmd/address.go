package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
)

func NewAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address",
		Short: "Address utilities",
	}
	cmd.AddCommand(NewAddressConvertCmd())
	return cmd
}

func NewAddressConvertCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "convert [address]",
		Short: "Convert an address to Aetheris raw and userfriendly forms",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bz, err := aetherisaddress.Parse(args[0])
			if err != nil {
				return err
			}
			raw := aetherisaddress.Format(bz)
			userFriendly, err := aetherisaddress.FormatUserFriendly(bz)
			if err != nil {
				return err
			}
			out := struct {
				Raw          string `json:"raw"`
				UserFriendly string `json:"user_friendly"`
			}{
				Raw:          raw,
				UserFriendly: userFriendly,
			}
			bzJSON, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(bzJSON))
			return err
		},
	}
}
