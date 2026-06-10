package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

type FeeCollectorKeeper interface {
	RecordCollectedFees(ctx context.Context, fees sdk.Coins, feeType string) error
}

// ReputationReader reads on-chain identity reputation scores for fee adjustments.
// AWCE-1 temporary integration boundary — do not access reputation module internals here.
type ReputationReader interface {
	// GetIdentityReputationScore returns the reputation score [0..10000] for addr.
	// found=false means no record exists; treat as neutral (score == 5000).
	GetIdentityReputationScore(ctx context.Context, addr sdk.AccAddress) (score uint32, found bool, err error)
}
