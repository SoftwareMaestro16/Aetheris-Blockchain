package txhandlers

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
)

// NativeAccountStatusReader provides read-only access to native account status
// for ante-handler-level storage rent freeze checks.
type NativeAccountStatusReader interface {
	AccountStatus(ctx context.Context, userAddress string) (string, bool, error)
}

// StorageRentDecorator returns an ante handler that rejects standard SDK
// transactions from frozen native accounts. Frozen accounts can only perform
// recovery operations through native-account module messages (storage debt
// payment, unfreeze, recovery), which are validated separately by the native
// account keeper and do not pass through this ante handler.
func StorageRentDecorator(accountReader NativeAccountStatusReader, next sdk.AnteHandler) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		if simulate {
			return next(ctx, tx, simulate)
		}

		sigTx, ok := tx.(authsigning.SigVerifiableTx)
		if !ok {
			return next(ctx, tx, simulate)
		}

		signers, err := sigTx.GetSigners()
		if err != nil {
			return ctx, err
		}

		for _, signer := range signers {
			if len(signer) == 0 {
				continue
			}
			aeAddr := aetraaddress.FormatAccAddress(signer)
			status, found, err := accountReader.AccountStatus(ctx, aeAddr)
			if err != nil {
				return ctx, err
			}
			if !found || status != nativeaccounttypes.AccountStatusFrozen {
				continue
			}
			return ctx, fmt.Errorf(
				"frozen native account %s cannot perform this action: "+
					"only storage debt payment, unfreeze, and recovery are allowed",
				aeAddr,
			)
		}

		return next(ctx, tx, simulate)
	}
}
