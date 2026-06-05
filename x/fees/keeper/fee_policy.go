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
	return types.ValidateFeeCoins(params, fees, true)
}
