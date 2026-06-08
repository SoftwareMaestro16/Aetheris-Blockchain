package keeperwiring

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

	aetraeconomicskeeper "github.com/sovereign-l1/l1/x/aetra-economics/keeper"
	aetrastakingpolicykeeper "github.com/sovereign-l1/l1/x/aetra-staking-policy/keeper"
	aetravalidatorscorekeeper "github.com/sovereign-l1/l1/x/aetra-validator-score/keeper"
	burnkeeper "github.com/sovereign-l1/l1/x/burn/keeper"
	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	delegatorprotectionkeeper "github.com/sovereign-l1/l1/x/delegator-protection/keeper"
	delegatorprotectiontypes "github.com/sovereign-l1/l1/x/delegator-protection/types"
	dynamiccommissionkeeper "github.com/sovereign-l1/l1/x/dynamic-commission/keeper"
	dynamiccommissiontypes "github.com/sovereign-l1/l1/x/dynamic-commission/types"
	emissionskeeper "github.com/sovereign-l1/l1/x/emissions/keeper"
	emissionstypes "github.com/sovereign-l1/l1/x/emissions/types"
	feecollectorkeeper "github.com/sovereign-l1/l1/x/fee-collector/keeper"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	mintauthoritykeeper "github.com/sovereign-l1/l1/x/mint-authority/keeper"
	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
	performancekeeper "github.com/sovereign-l1/l1/x/performance/keeper"
	performancetypes "github.com/sovereign-l1/l1/x/performance/types"
	reputationkeeper "github.com/sovereign-l1/l1/x/reputation/keeper"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
	stakeconcentrationkeeper "github.com/sovereign-l1/l1/x/stake-concentration/keeper"
	stakeconcentrationtypes "github.com/sovereign-l1/l1/x/stake-concentration/types"
	treasurykeeper "github.com/sovereign-l1/l1/x/treasury/keeper"
	treasurytypes "github.com/sovereign-l1/l1/x/treasury/types"
)

type NativeKeeperDeps struct {
	AppCodec      codec.Codec
	Keys          map[string]*storetypes.KVStoreKey
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.BaseKeeper
	DistrKeeper   distrkeeper.Keeper
	GovAuthority  string
}

type NativeKeepers struct {
	BurnKeeper                burnkeeper.Keeper
	TreasuryKeeper            treasurykeeper.Keeper
	EmissionsKeeper           emissionskeeper.Keeper
	MintAuthorityKeeper       mintauthoritykeeper.Keeper
	DelegatorProtectionKeeper delegatorprotectionkeeper.Keeper
	ReputationKeeper          reputationkeeper.Keeper
	PerformanceKeeper         performancekeeper.Keeper
	DynamicCommissionKeeper   dynamiccommissionkeeper.Keeper
	StakeConcentrationKeeper  stakeconcentrationkeeper.Keeper
	FeeCollectorKeeper        feecollectorkeeper.Keeper
	FeesKeeper                feeskeeper.Keeper
	AetraStakingPolicyKeeper  aetrastakingpolicykeeper.Keeper
	AetraEconomicsKeeper      aetraeconomicskeeper.Keeper
	AetraValidatorScoreKeeper aetravalidatorscorekeeper.Keeper
}

func NewNativeKeepers(deps NativeKeeperDeps) NativeKeepers {
	return NativeKeepers{
		BurnKeeper: burnkeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[burntypes.StoreKey]),
			deps.BankKeeper,
			deps.GovAuthority,
		),
		TreasuryKeeper: treasurykeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[treasurytypes.StoreKey]),
			deps.AccountKeeper,
			deps.BankKeeper,
			deps.GovAuthority,
		),
		EmissionsKeeper: emissionskeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[emissionstypes.StoreKey]),
			deps.GovAuthority,
		),
		MintAuthorityKeeper: mintauthoritykeeper.NewKeeper(
			runtime.NewKVStoreService(deps.Keys[mintauthoritytypes.StoreKey]),
			deps.BankKeeper,
			deps.GovAuthority,
		),
		DelegatorProtectionKeeper: delegatorprotectionkeeper.NewKeeper(
			runtime.NewKVStoreService(deps.Keys[delegatorprotectiontypes.StoreKey]),
			deps.GovAuthority,
		),
		ReputationKeeper: reputationkeeper.NewKeeper(
			runtime.NewKVStoreService(deps.Keys[reputationtypes.StoreKey]),
			deps.GovAuthority,
		),
		PerformanceKeeper: performancekeeper.NewKeeper(
			runtime.NewKVStoreService(deps.Keys[performancetypes.StoreKey]),
			deps.GovAuthority,
		),
		DynamicCommissionKeeper: dynamiccommissionkeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[dynamiccommissiontypes.StoreKey]),
			deps.GovAuthority,
		),
		StakeConcentrationKeeper: stakeconcentrationkeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[stakeconcentrationtypes.StoreKey]),
			deps.GovAuthority,
		),
		FeeCollectorKeeper: feecollectorkeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[feecollectortypes.StoreKey]),
			deps.AccountKeeper,
			deps.BankKeeper,
			deps.GovAuthority,
		),
		FeesKeeper: feeskeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[feestypes.StoreKey]),
			deps.AccountKeeper,
			deps.BankKeeper,
			deps.DistrKeeper,
			deps.GovAuthority,
		),
		AetraStakingPolicyKeeper:  aetrastakingpolicykeeper.NewKeeper(deps.GovAuthority),
		AetraEconomicsKeeper:      aetraeconomicskeeper.NewKeeper(deps.GovAuthority),
		AetraValidatorScoreKeeper: aetravalidatorscorekeeper.NewKeeper(deps.GovAuthority),
	}
}
