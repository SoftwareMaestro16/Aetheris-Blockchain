package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestNominatorPoolUnbondingUsesStakingPolicyWindow(t *testing.T) {
	gs := DefaultGenesis()

	require.Equal(t, appparams.StakingUnbondingDefaultBlocks, gs.Params.UnbondingBlocks)
	require.NoError(t, gs.Validate())

	gs.Params.UnbondingBlocks = appparams.StakingUnbondingMinBlocks - 1
	require.ErrorContains(t, gs.Validate(), "14-21 days")
}

func TestDepositMintsShares(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")

	share, err := k.DepositToPool(types.MsgDepositToPool{
		Authority: prototype.DefaultAuthority,
		PoolID:    pool.PoolID,
		Delegator: rawPoolAddress("22"),
		Amount:    1_000,
		Height:    2,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1_000), share.Shares)
	stored, found := k.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, uint64(1_000), stored.TotalShares)
	require.Equal(t, uint64(1_000), stored.TotalBondedStake)
}

func TestOfficialLiquidStakingSmallDepositMintsDeterministicShares(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	user := aePoolAddress(t, "22")

	share, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:   prototype.DefaultAuthority,
		PoolID:      pool.PoolID,
		UserAddress: user,
		Amount:      types.DefaultMinPoolDeposit,
		Height:      2,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(10), share.Shares)

	secondShare, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:   prototype.DefaultAuthority,
		PoolID:      pool.PoolID,
		UserAddress: user,
		Amount:      90,
		Height:      3,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(100), secondShare.Shares)

	stored, found := k.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, uint64(100), stored.TotalBondedStake)
	require.Equal(t, uint64(100), stored.TotalShares)
	require.Equal(t, rawPoolAddress("22"), stored.DelegatorShares[0].Delegator)
	require.Equal(t, uint64(100), stored.DelegatorShares[0].Shares)
}

func TestOfficialLiquidStakingDepositRejectsValidatorAddress(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")

	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:        prototype.DefaultAuthority,
		PoolID:           pool.PoolID,
		UserAddress:      aePoolAddress(t, "22"),
		ValidatorAddress: aePoolAddress(t, "33"),
		Amount:           100,
		Height:           2,
	})
	require.ErrorContains(t, err, "must not include a validator address")
}

func TestOfficialLiquidStakingDepositBelowMinimumRejected(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")

	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:   prototype.DefaultAuthority,
		PoolID:      pool.PoolID,
		UserAddress: aePoolAddress(t, "22"),
		Amount:      types.DefaultMinPoolDeposit - 1,
		Height:      2,
	})
	require.ErrorContains(t, err, "below configured minimum")
}

func TestOfficialLiquidStakingDepositRequiresAEUserAddress(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")

	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:   prototype.DefaultAuthority,
		PoolID:      pool.PoolID,
		UserAddress: rawPoolAddress("22"),
		Amount:      100,
		Height:      2,
	})
	require.ErrorContains(t, err, "must use AE user-facing address format")
}

func TestOfficialLiquidStakingReceiptExportImportRoundTrip(t *testing.T) {
	source := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &source, "official-pool")
	user := aePoolAddress(t, "22")
	share, err := source.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:   prototype.DefaultAuthority,
		PoolID:      pool.PoolID,
		UserAddress: user,
		Amount:      100,
		Height:      2,
	})
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))

	roundTripShare, found := target.PoolDelegator(pool.PoolID, rawPoolAddress("22"))
	require.True(t, found)
	require.Equal(t, share, roundTripShare)
	require.Equal(t, exported, target.ExportGenesis())
}

func TestDirectUserDelegationDisabledByDefault(t *testing.T) {
	k := NewKeeper()

	err := k.DelegateUserToValidator(types.MsgDelegateToValidator{
		Authority:        prototype.DefaultAuthority,
		UserAddress:      aePoolAddress(t, "22"),
		ValidatorAddress: aePoolAddress(t, "33"),
		Amount:           100,
		Height:           2,
	})
	require.ErrorContains(t, err, "direct user delegation to validators is disabled")
}

func TestOfficialContractCanInjectPooledStake(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:   prototype.DefaultAuthority,
		PoolID:      pool.PoolID,
		UserAddress: aePoolAddress(t, "22"),
		Amount:      100,
		Height:      2,
	})
	require.NoError(t, err)

	updated, err := k.InjectPooledStake(types.MsgInjectPooledStake{
		CallerContractUser: pool.ContractAddressUser,
		PoolID:             pool.PoolID,
		ValidatorAddress:   aePoolAddress(t, "33"),
		Amount:             100,
		Height:             3,
	})
	require.NoError(t, err)
	require.Equal(t, []types.PoolAllocation{{
		ValidatorAddress: aePoolAddress(t, "33"),
		Amount:           100,
		Height:           3,
	}}, updated.Allocations)
}

func TestUnauthorizedContractCannotInjectPooledStake(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:   prototype.DefaultAuthority,
		PoolID:      pool.PoolID,
		UserAddress: aePoolAddress(t, "22"),
		Amount:      100,
		Height:      2,
	})
	require.NoError(t, err)

	_, err = k.InjectPooledStake(types.MsgInjectPooledStake{
		CallerContractUser: aePoolAddress(t, "77"),
		PoolID:             pool.PoolID,
		ValidatorAddress:   aePoolAddress(t, "33"),
		Amount:             100,
		Height:             3,
	})
	require.ErrorContains(t, err, "requires official liquid staking contract")
}

func TestPooledStakeInjectionCannotExceedPoolAccounting(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:   prototype.DefaultAuthority,
		PoolID:      pool.PoolID,
		UserAddress: aePoolAddress(t, "22"),
		Amount:      100,
		Height:      2,
	})
	require.NoError(t, err)

	_, err = k.InjectPooledStake(types.MsgInjectPooledStake{
		CallerContractUser: pool.ContractAddressUser,
		PoolID:             pool.PoolID,
		ValidatorAddress:   aePoolAddress(t, "33"),
		Amount:             101,
		Height:             3,
	})
	require.ErrorContains(t, err, "exceeds unallocated pool stake")
}

func TestFrozenLimitedOfficialPoolRejectsDeposits(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	gs := k.ExportGenesis()
	gs.State.Pools[0].Status = types.PoolStatusFrozenLimited
	require.NoError(t, k.InitGenesis(gs))

	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:   prototype.DefaultAuthority,
		PoolID:      pool.PoolID,
		UserAddress: aePoolAddress(t, "22"),
		Amount:      100,
		Height:      2,
	})
	require.ErrorContains(t, err, "must be active for deposits")
}

func TestWithdrawalBurnsShares(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")
	deposit(t, &k, pool.PoolID, rawPoolAddress("22"), 1_000, 2)

	withdrawal, err := k.RequestPoolWithdrawal(types.MsgRequestPoolWithdrawal{
		Authority:    prototype.DefaultAuthority,
		PoolID:       pool.PoolID,
		WithdrawalID: "withdraw-1",
		Delegator:    rawPoolAddress("22"),
		Shares:       400,
		Height:       3,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(400), withdrawal.Amount)
	stored, found := k.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, uint64(600), stored.TotalShares)
	require.Equal(t, uint64(600), stored.TotalBondedStake)
	require.Len(t, stored.UnbondingQueue, 1)
}

func TestRewardsDistributeProportionally(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")
	a := rawPoolAddress("22")
	b := rawPoolAddress("33")
	deposit(t, &k, pool.PoolID, a, 1_000, 2)
	deposit(t, &k, pool.PoolID, b, 3_000, 3)

	_, err := k.ApplyPoolReward(pool.PoolID, 400)
	require.NoError(t, err)
	rewardA, found := k.PoolRewards(pool.PoolID, a)
	require.True(t, found)
	rewardB, found := k.PoolRewards(pool.PoolID, b)
	require.True(t, found)
	require.Equal(t, uint64(99), rewardA)
	require.Equal(t, uint64(297), rewardB)

	claimed, err := k.ClaimPoolRewards(types.MsgClaimPoolRewards{Authority: prototype.DefaultAuthority, PoolID: pool.PoolID, Delegator: a, Height: 4})
	require.NoError(t, err)
	require.Equal(t, uint64(99), claimed)
}

func TestSyncPoolRewardsAfterEpochProgressionAndIllustrativeEconomics(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-rewards")
	user := rawPoolAddress("22")
	deposit(t, &k, pool.PoolID, user, 3_000_000, 2)

	summary, err := k.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:          prototype.DefaultAuthority,
		PoolID:             pool.PoolID,
		Epoch:              1,
		RewardRateBps:      1_440,
		EmissionsAllocated: 864_000,
		Height:             3,
		Allocations: []types.ValidatorRewardAllocation{{
			Validator:          rawPoolAddress("12"),
			PoolAllocatedStake: 3_000_000,
			ValidatorSelfStake: 3_000_000,
			PerformanceBps:     types.MaxBasisPoints,
			CommissionBps:      1_000,
			InfrastructureCost: 20_000,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(432_000), summary.GrossPoolRewards)
	require.Equal(t, uint64(43_200), summary.ValidatorCommission)
	require.Equal(t, uint64(3_888), summary.PoolProtocolFee)
	require.Equal(t, uint64(384_912), summary.PoolUserRewards)
	require.Equal(t, uint64(432_000), summary.ValidatorSelfStakeRewards)
	require.Equal(t, uint64(475_200), summary.ValidatorGrossIncome)
	require.Equal(t, int64(455_200), summary.ValidatorNetIncome)
	require.Greater(t, summary.ValidatorGrossIncome, summary.PoolUserRewards)

	claimed, err := k.ClaimPoolRewards(types.MsgClaimPoolRewards{Authority: prototype.DefaultAuthority, PoolID: pool.PoolID, Delegator: user, Height: 4})
	require.NoError(t, err)
	require.Equal(t, uint64(384_912), claimed)

	claimedAgain, err := k.ClaimPoolRewards(types.MsgClaimPoolRewards{Authority: prototype.DefaultAuthority, PoolID: pool.PoolID, Delegator: user, Height: 5})
	require.NoError(t, err)
	require.Zero(t, claimedAgain)
}

func TestSyncPoolRewardsProportionalSharesAndDeterministicRounding(t *testing.T) {
	k := NewKeeper()
	pool := createPoolWithCommission(t, &k, "pool-proportional", 0)
	a := rawPoolAddress("22")
	b := rawPoolAddress("33")
	deposit(t, &k, pool.PoolID, a, 1_000, 2)
	deposit(t, &k, pool.PoolID, b, 3_000, 3)

	require.Equal(t, uint64(333_333_333), types.RewardDelta(1, 3))
	require.Equal(t, uint64(0), types.IndexedRewardAmount(types.RewardDelta(1, 3), 3))

	summary, err := k.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:          prototype.DefaultAuthority,
		PoolID:             pool.PoolID,
		Epoch:              1,
		RewardRateBps:      1_000,
		EmissionsAllocated: 400,
		Height:             4,
		Allocations: []types.ValidatorRewardAllocation{{
			Validator:          rawPoolAddress("12"),
			PoolAllocatedStake: 4_000,
			PerformanceBps:     types.MaxBasisPoints,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(400), summary.PoolUserRewards)

	rewardA, found := k.PoolRewards(pool.PoolID, a)
	require.True(t, found)
	rewardB, found := k.PoolRewards(pool.PoolID, b)
	require.True(t, found)
	require.Equal(t, uint64(100), rewardA)
	require.Equal(t, uint64(300), rewardB)
}

func TestValidatorPerformanceCommissionPoolFeeAndJailAreDeterministic(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-performance")
	deposit(t, &k, pool.PoolID, rawPoolAddress("22"), 2_000, 2)

	summary, err := k.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:          prototype.DefaultAuthority,
		PoolID:             pool.PoolID,
		Epoch:              1,
		RewardRateBps:      1_000,
		EmissionsAllocated: 2_000,
		Height:             3,
		Allocations: []types.ValidatorRewardAllocation{
			{
				Validator:          rawPoolAddress("12"),
				PoolAllocatedStake: 1_000,
				PerformanceBps:     types.MaxBasisPoints,
				CommissionBps:      1_000,
			},
			{
				Validator:                   rawPoolAddress("13"),
				PoolAllocatedStake:          1_000,
				PerformanceBps:              5_000,
				CommissionBps:               1_000,
				Jailed:                      true,
				OperatorPerformanceBonusBps: 100,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(100), summary.GrossPoolRewards)
	require.Equal(t, uint64(10), summary.ValidatorCommission)
	require.Equal(t, uint64(0), summary.OperatorPerformanceBonus)
	require.Equal(t, uint64(90), summary.PoolUserRewards)

	allocations, found := k.PoolAllocations(types.QueryPoolAllocationsRequest{PoolID: pool.PoolID})
	require.True(t, found)
	require.Len(t, allocations.Allocations, 2)
	require.Equal(t, uint64(100), allocations.Allocations[0].GrossPoolRewards)
	require.Zero(t, allocations.Allocations[1].GrossPoolRewards)
	require.Zero(t, allocations.Allocations[1].OperatorPerformanceBonus)
}

func TestSyncPoolRewardsRejectsEmissionFeeCapExceeded(t *testing.T) {
	k := NewKeeper()
	pool := createPoolWithCommission(t, &k, "pool-cap", 0)
	deposit(t, &k, pool.PoolID, rawPoolAddress("22"), 1_000, 2)

	_, err := k.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:          prototype.DefaultAuthority,
		PoolID:             pool.PoolID,
		Epoch:              1,
		RewardRateBps:      1_000,
		EmissionsAllocated: 99,
		Height:             3,
		Allocations: []types.ValidatorRewardAllocation{{
			Validator:          rawPoolAddress("12"),
			PoolAllocatedStake: 1_000,
			PerformanceBps:     types.MaxBasisPoints,
		}},
	})
	require.ErrorContains(t, err, "exceed emissions")
}

func TestExportImportPreservesSyncedRewardStateAndPendingRewards(t *testing.T) {
	source := NewKeeper()
	pool := createPoolWithCommission(t, &source, "pool-export", 0)
	user := rawPoolAddress("22")
	deposit(t, &source, pool.PoolID, user, 3, 2)
	_, err := source.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:          prototype.DefaultAuthority,
		PoolID:             pool.PoolID,
		Epoch:              1,
		RewardRateBps:      3_334,
		EmissionsAllocated: 1,
		Height:             3,
		Allocations: []types.ValidatorRewardAllocation{{
			Validator:          rawPoolAddress("12"),
			PoolAllocatedStake: 3,
			PerformanceBps:     types.MaxBasisPoints,
		}},
	})
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	rewards, found := target.PoolRewards(pool.PoolID, user)
	require.True(t, found)
	require.Equal(t, uint64(0), rewards)
	stored, found := target.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, uint64(1), stored.RewardRemainder)
	require.Len(t, stored.ValidatorAllocations, 1)
}

func TestClaimPoolRewardsTouchesBoundedKeysWithLargeUserSet(t *testing.T) {
	k := NewKeeper()
	const users = 1_000_000
	shares := make([]types.DelegatorShare, users)
	for i := range shares {
		shares[i] = types.DelegatorShare{
			Delegator:             rawPoolAddressFromInt(i + 1),
			Shares:                1,
			RewardIndexCheckpoint: 0,
		}
	}
	k.genesis.State.Pools = []types.NominatorPool{{
		PoolID:            "pool-large",
		PoolOperator:      rawPoolAddress("11"),
		ValidatorTarget:   rawPoolAddress("12"),
		TotalShares:       users,
		TotalBondedStake:  users,
		DelegatorShares:   shares,
		RewardIndex:       types.IndexScale,
		PoolCommissionBps: 100,
		Status:            types.PoolStatusActive,
	}}
	k.rebuildIndexes()
	k.ResetOperationCounters()

	claimed, err := k.ClaimPoolRewards(types.MsgClaimPoolRewards{
		Authority: prototype.DefaultAuthority,
		PoolID:    "pool-large",
		Delegator: rawPoolAddressFromInt(users),
		Height:    2,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), claimed)
	counters := k.OperationCounters()
	require.Equal(t, uint64(1), counters.PoolLookups)
	require.Equal(t, uint64(1), counters.DelegatorLookups)
	require.Equal(t, uint64(1), counters.DelegatorRewardUpdates)
}

func TestStakingRewardsCompatibilityIsInternalOnly(t *testing.T) {
	k := NewKeeper()
	_, err := k.ClaimStakingRewards(types.MsgClaimStakingRewards{
		Authority: prototype.DefaultAuthority,
		Delegator: rawPoolAddress("22"),
		Validator: rawPoolAddress("12"),
		Height:    1,
	})
	require.ErrorContains(t, err, "internal migration only")

	amount, err := k.ClaimStakingRewards(types.MsgClaimStakingRewards{
		Authority:         prototype.DefaultAuthority,
		Delegator:         rawPoolAddress("22"),
		Validator:         rawPoolAddress("12"),
		Height:            1,
		InternalMigration: true,
	})
	require.NoError(t, err)
	require.Zero(t, amount)

	_, err = k.StakingRewards(types.QueryStakingRewardsRequest{Delegator: rawPoolAddress("22"), Validator: rawPoolAddress("12")})
	require.ErrorContains(t, err, "internal migration only")
}

func TestSlashAppliesProportionally(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")
	a := rawPoolAddress("22")
	b := rawPoolAddress("33")
	deposit(t, &k, pool.PoolID, a, 1_000, 2)
	deposit(t, &k, pool.PoolID, b, 3_000, 3)

	slashed, err := k.ApplyPoolSlash(pool.PoolID, 800)
	require.NoError(t, err)
	require.Equal(t, uint64(3_200), slashed.TotalBondedStake)

	shareA, found := k.PoolDelegator(pool.PoolID, a)
	require.True(t, found)
	shareB, found := k.PoolDelegator(pool.PoolID, b)
	require.True(t, found)
	require.Equal(t, uint64(800), types.ShareValue(slashed, shareA.Shares))
	require.Equal(t, uint64(2_400), types.ShareValue(slashed, shareB.Shares))
}

func TestPoolCannotWithdrawMoreThanTotalStake(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")
	deposit(t, &k, pool.PoolID, rawPoolAddress("22"), 1_000, 2)

	_, err := k.RequestPoolWithdrawal(types.MsgRequestPoolWithdrawal{
		Authority:    prototype.DefaultAuthority,
		PoolID:       pool.PoolID,
		WithdrawalID: "withdraw-too-much",
		Delegator:    rawPoolAddress("22"),
		Shares:       1_001,
		Height:       3,
	})
	require.ErrorContains(t, err, "more than total stake")
}

func TestExportImportPreservesRewardIndex(t *testing.T) {
	source := NewKeeper()
	pool := createPool(t, &source, "pool-a")
	deposit(t, &source, pool.PoolID, rawPoolAddress("22"), 1_000, 2)
	_, err := source.ApplyPoolReward(pool.PoolID, 123)
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
	stored, found := target.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, exported.State.Pools[0].RewardIndex, stored.RewardIndex)
}

func TestPoolValidatorChangeDelayEnforced(t *testing.T) {
	k := NewKeeper()
	gs := k.ExportGenesis()
	gs.Params.ValidatorChangeDelay = 10
	require.NoError(t, k.InitGenesis(gs))
	pool := createPool(t, &k, "pool-a")
	nextValidator := rawPoolAddress("44")

	pending, err := k.ChangePoolValidator(types.MsgChangePoolValidator{
		Authority:       prototype.DefaultAuthority,
		PoolID:          pool.PoolID,
		PoolOperator:    pool.PoolOperator,
		ValidatorTarget: nextValidator,
		ValidatorStatus: validatorregistrytypes.StatusActive,
		Height:          5,
	})
	require.NoError(t, err)
	require.Equal(t, pool.ValidatorTarget, pending.ValidatorTarget)
	require.Equal(t, nextValidator, pending.PendingValidatorTarget)
	require.Equal(t, uint64(15), pending.ValidatorChangeHeight)

	stillPending, err := k.ChangePoolValidator(types.MsgChangePoolValidator{
		Authority:       prototype.DefaultAuthority,
		PoolID:          pool.PoolID,
		PoolOperator:    pool.PoolOperator,
		ValidatorTarget: nextValidator,
		ValidatorStatus: validatorregistrytypes.StatusActive,
		Height:          14,
	})
	require.NoError(t, err)
	require.NotEqual(t, nextValidator, stillPending.ValidatorTarget)

	finalized, err := k.ChangePoolValidator(types.MsgChangePoolValidator{
		Authority:       prototype.DefaultAuthority,
		PoolID:          pool.PoolID,
		PoolOperator:    pool.PoolOperator,
		ValidatorTarget: nextValidator,
		ValidatorStatus: validatorregistrytypes.StatusActive,
		Height:          15,
	})
	require.NoError(t, err)
	require.Equal(t, nextValidator, finalized.ValidatorTarget)
	require.Empty(t, finalized.PendingValidatorTarget)
}

func TestPoolCannotDelegateToJailedValidator(t *testing.T) {
	k := NewKeeper()
	_, err := k.CreateNominatorPool(types.MsgCreateNominatorPool{
		Authority:         prototype.DefaultAuthority,
		PoolID:            "pool-jailed",
		PoolOperator:      rawPoolAddress("11"),
		ValidatorTarget:   rawPoolAddress("12"),
		PoolCommissionBps: 100,
		Height:            1,
		ValidatorStatus:   validatorregistrytypes.StatusJailed,
	})
	require.ErrorContains(t, err, "jailed validator")
}

func createPool(t *testing.T, k *Keeper, poolID string) types.NominatorPool {
	t.Helper()
	return createPoolWithCommission(t, k, poolID, 100)
}

func createPoolWithCommission(t *testing.T, k *Keeper, poolID string, commissionBps uint32) types.NominatorPool {
	t.Helper()
	pool, err := k.CreateNominatorPool(types.MsgCreateNominatorPool{
		Authority:         prototype.DefaultAuthority,
		PoolID:            poolID,
		PoolOperator:      rawPoolAddress("11"),
		ValidatorTarget:   rawPoolAddress("12"),
		PoolCommissionBps: commissionBps,
		Height:            1,
		ValidatorStatus:   validatorregistrytypes.StatusActive,
	})
	require.NoError(t, err)
	return pool
}

func createOfficialLiquidStakingPool(t *testing.T, k *Keeper, poolID string) types.NominatorPool {
	t.Helper()
	contractRaw := rawPoolAddress("66")
	pool, err := k.CreateOfficialLiquidStakingPool(types.MsgCreateOfficialLiquidStakingPool{
		Authority:           prototype.DefaultAuthority,
		PoolID:              poolID,
		ContractAddressUser: aeFromRaw(t, contractRaw),
		ContractAddressRaw:  contractRaw,
		PoolOperator:        rawPoolAddress("11"),
		PoolCommissionBps:   100,
		Height:              1,
	})
	require.NoError(t, err)
	return pool
}

func deposit(t *testing.T, k *Keeper, poolID string, delegator string, amount uint64, height uint64) {
	t.Helper()
	_, err := k.DepositToPool(types.MsgDepositToPool{
		Authority: prototype.DefaultAuthority,
		PoolID:    poolID,
		Delegator: delegator,
		Amount:    amount,
		Height:    height,
	})
	require.NoError(t, err)
}

func rawPoolAddress(hexByte string) string {
	return "4:000000000000000000000000" + fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s", hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte)
}

func rawPoolAddressFromInt(value int) string {
	return fmt.Sprintf("4:%064x", value)
}

func aePoolAddress(t *testing.T, hexByte string) string {
	t.Helper()
	return aeFromRaw(t, rawPoolAddress(hexByte))
}

func aeFromRaw(t *testing.T, raw string) string {
	t.Helper()
	bz, err := addressing.Parse(raw)
	require.NoError(t, err)
	user, err := addressing.FormatUserFriendly(bz)
	require.NoError(t, err)
	return user
}
