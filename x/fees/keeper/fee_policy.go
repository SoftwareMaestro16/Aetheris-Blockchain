package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/fees/types"
)

func (k Keeper) ValidateTxFees(ctx context.Context, fees sdk.Coins) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := params.Validate(); err != nil {
		return types.ErrInvalidParams.Wrap(err.Error())
	}
	allowedDenoms := make(map[string]struct{}, len(params.AllowedFeeDenoms))
	for _, denom := range params.AllowedFeeDenoms {
		allowedDenoms[denom] = struct{}{}
	}

	seenDenoms := make(map[string]struct{}, len(fees))
	for _, fee := range fees {
		if !fee.IsValid() {
			return types.ErrInvalidFee.Wrap("fee coins must be valid")
		}
		if _, seen := seenDenoms[fee.Denom]; seen {
			return types.ErrInvalidFee.Wrap("fee coins must be valid")
		}
		seenDenoms[fee.Denom] = struct{}{}
		if _, ok := allowedDenoms[fee.Denom]; !ok {
			return types.ErrInvalidFee.Wrapf("fee denom %s not accepted; use %s", fee.Denom, types.BondDenom)
		}
	}
	return nil
}
