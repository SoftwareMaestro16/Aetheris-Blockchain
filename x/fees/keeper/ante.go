package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/fees/types"
)

func (k Keeper) AnteHandlerDecorator(next sdk.AnteHandler) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return ctx, types.ErrInvalidFee.Wrap("transaction must expose fees")
		}
		if err := k.ValidateTxFees(ctx, feeTx.GetFee()); err != nil {
			return ctx, err
		}
		return next(ctx, tx, simulate)
	}
}
