package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	reputationkeeper "github.com/sovereign-l1/l1/x/reputation/keeper"
	"github.com/sovereign-l1/l1/x/reputation/types"
	reputationpb "github.com/sovereign-l1/l1/x/reputation/types/reputationpb"
)

func TestNativeReputationRewardPenaltyAndExport(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := reputationkeeper.NewMsgServerImpl(app.ReputationKeeper)
	subject := addr(0x11)

	_, err := msgServer.ApplyReputationReward(ctx, &reputationpb.MsgApplyReputationReward{
		Authority:   app.ReputationKeeper.Authority(),
		SubjectType: types.SubjectValidator,
		Subject:     subject,
		Component:   types.ComponentUptime,
		Epoch:       1,
	})
	require.NoError(t, err)

	_, err = msgServer.ApplyReputationPenalty(ctx, &reputationpb.MsgApplyReputationPenalty{
		Authority:   addr(0x22),
		SubjectType: types.SubjectValidator,
		Subject:     subject,
		Component:   types.ComponentSlashing,
		Epoch:       2,
	})
	require.Error(t, err)

	exported, err := app.ReputationKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	imported := l1app.Setup(t, false)
	importedCtx := imported.NewContext(false)
	require.NoError(t, imported.ReputationKeeper.InitGenesis(importedCtx, *exported))
	roundTrip, err := imported.ReputationKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, exported, roundTrip)
}

func TestStakeReputationClaimQueryAndExport(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	account := addr(0x33)

	_, err := app.ReputationKeeper.ClaimStakeReputation(ctx, types.MsgClaimStakeReputation{
		Authority:       app.ReputationKeeper.Authority(),
		Account:         account,
		PoolID:          "pool-a",
		PoolShares:      100,
		PoolTotalShares: 100,
		PoolActiveStake: 10_000,
		TimestampUnix:   1,
	})
	require.NoError(t, err)
	claim, err := app.ReputationKeeper.ClaimStakeReputation(ctx, types.MsgClaimStakeReputation{
		Authority:       app.ReputationKeeper.Authority(),
		Account:         account,
		PoolID:          "pool-a",
		PoolShares:      100,
		PoolTotalShares: 100,
		PoolActiveStake: 10_000,
		TimestampUnix:   3_601,
	})
	require.NoError(t, err)
	require.Equal(t, uint16(100), claim.ReputationDelta)

	stake, err := app.ReputationKeeper.StakeReputation(ctx, types.QueryStakeReputationRequest{Account: account})
	require.NoError(t, err)
	require.Equal(t, account, stake.Record.AccountUser)
	accountReputation, err := app.ReputationKeeper.AccountReputation(ctx, types.QueryAccountReputationRequest{Account: account})
	require.NoError(t, err)
	require.Equal(t, uint16(100), accountReputation.Record.StakingScore)
	require.Equal(t, stake.Record, accountReputation.StakeReputation)

	exported, err := app.ReputationKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	imported := l1app.Setup(t, false)
	importedCtx := imported.NewContext(false)
	require.NoError(t, imported.ReputationKeeper.InitGenesis(importedCtx, *exported))
	roundTrip, err := imported.ReputationKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, exported, roundTrip)
}

func addr(fill byte) string {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = fill
	}
	return aetraaddress.FormatAccAddress(sdk.AccAddress(bz))
}
