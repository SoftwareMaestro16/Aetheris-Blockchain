package txhandlers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sovereign-l1/l1/app/stakingpolicy"
)

func RejectDirectUserStakingDecorator(next sdk.AnteHandler) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		for _, msg := range tx.GetMsgs() {
			switch msg := msg.(type) {
			case *stakingtypes.MsgDelegate:
				if err := stakingpolicy.ValidateDelegate(stakingpolicy.DefaultDirectDelegationPolicy(), msg); err != nil {
					return ctx, err
				}
			case *stakingtypes.MsgBeginRedelegate:
				if err := stakingpolicy.ValidateBeginRedelegate(stakingpolicy.DefaultDirectDelegationPolicy(), msg); err != nil {
					return ctx, err
				}
			case *stakingtypes.MsgUndelegate:
				if err := stakingpolicy.ValidateUndelegate(stakingpolicy.DefaultDirectDelegationPolicy(), msg); err != nil {
					return ctx, err
				}
			}
		}
		return next(ctx, tx, simulate)
	}
}
