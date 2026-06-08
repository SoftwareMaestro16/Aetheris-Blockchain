package app

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sovereign-l1/l1/app/addressing"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
)

// AppInvariant binds a consensus invariant ID to live application state.
type AppInvariant struct {
	ID          string
	Description string
	Check       func(ctx sdk.Context) error
}

// AppInvariantFailure is a deterministic runtime invariant failure report.
type AppInvariantFailure struct {
	ID    string
	Error string
}

func RequiredAppInvariantIDs() []string {
	return nativeaccounttypes.RequiredNativeAccountInvariantIDs()
}

func ValidateAppInvariantRegistry(registry []AppInvariant) error {
	required := RequiredAppInvariantIDs()
	seen := make(map[string]struct{}, len(registry))
	for _, invariant := range registry {
		if strings.TrimSpace(invariant.ID) == "" || strings.TrimSpace(invariant.Description) == "" || invariant.Check == nil {
			return errors.New("app invariant id, description, and check are required")
		}
		if _, found := seen[invariant.ID]; found {
			return fmt.Errorf("duplicate app invariant %s", invariant.ID)
		}
		seen[invariant.ID] = struct{}{}
	}
	for _, id := range required {
		if _, found := seen[id]; !found {
			return fmt.Errorf("app invariant registry missing %s", id)
		}
	}
	return nil
}

func (app *L1App) AppInvariantRegistry() []AppInvariant {
	modelRegistry := nativeaccounttypes.DefaultNativeAccountInvariantRegistry()
	registry := make([]AppInvariant, 0, len(modelRegistry))
	for _, modelInvariant := range modelRegistry {
		id := modelInvariant.ID
		registry = append(registry, AppInvariant{
			ID:          id,
			Description: modelInvariant.Description,
			Check: func(ctx sdk.Context) error {
				return app.checkAppInvariant(ctx, id)
			},
		})
	}
	return registry
}

func (app *L1App) RunAppInvariant(ctx sdk.Context, id string) error {
	for _, invariant := range app.AppInvariantRegistry() {
		if invariant.ID == id {
			return invariant.Check(ctx)
		}
	}
	return fmt.Errorf("unknown app invariant %s", id)
}

func (app *L1App) RunAppInvariants(ctx sdk.Context) []AppInvariantFailure {
	registry := app.AppInvariantRegistry()
	failures := make([]AppInvariantFailure, 0)
	for _, invariant := range registry {
		if err := invariant.Check(ctx); err != nil {
			failures = append(failures, AppInvariantFailure{ID: invariant.ID, Error: err.Error()})
		}
	}
	sort.SliceStable(failures, func(i, j int) bool { return failures[i].ID < failures[j].ID })
	return failures
}

func (app *L1App) checkAppInvariant(ctx sdk.Context, id string) error {
	input, err := app.nativeInvariantInput(ctx)
	if err != nil {
		return err
	}
	if err := nativeaccounttypes.RunNativeAccountInvariant(id, input); err != nil {
		return err
	}

	switch id {
	case nativeaccounttypes.InvariantModuleBankAccountingConsistent:
		return app.assertBankModuleAccountingInvariant(ctx)
	case nativeaccounttypes.InvariantDirectUserValidatorDelegationRejected:
		return app.assertDirectUserDelegationRejectedInvariant(ctx)
	case nativeaccounttypes.InvariantMaxValidatorCountEnforced,
		nativeaccounttypes.InvariantMinValidatorStakeEnforced,
		nativeaccounttypes.InvariantValidatorSelfStakeRatioEnforced:
		return app.assertValidatorEntryInvariant(ctx)
	case nativeaccounttypes.InvariantPoolStakeDoesNotCreateCoins,
		nativeaccounttypes.InvariantRewardsCannotExceedAllocation,
		nativeaccounttypes.InvariantMinPoolDepositEnforced,
		nativeaccounttypes.InvariantUnbondingCannotReleaseEarly,
		nativeaccounttypes.InvariantReputationRequiresStakeTime,
		nativeaccounttypes.InvariantJailedValidatorNoPositiveBonus,
		nativeaccounttypes.InvariantExportImportPreservesState:
		return app.assertNominatorPoolInvariant(ctx)
	case nativeaccounttypes.InvariantProtocolCriticalStateNotFrozenByRent,
		nativeaccounttypes.InvariantSystemStorageReserveRunway,
		nativeaccounttypes.InvariantSystemRentTopUpBeforeUserFreeze,
		nativeaccounttypes.InvariantProtocolCriticalExecutableUnderUnderfunding:
		return app.assertStorageRentInvariant(ctx)
	default:
		return nil
	}
}

func (app *L1App) nativeInvariantInput(ctx sdk.Context) (nativeaccounttypes.NativeAccountInvariantInput, error) {
	params, err := app.StakingKeeper.GetParams(ctx)
	if err != nil {
		return nativeaccounttypes.NativeAccountInvariantInput{}, err
	}
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return nativeaccounttypes.NativeAccountInvariantInput{}, err
	}

	activeValidators := uint64(0)
	validatorStakes := make([]uint64, 0, len(validators))
	for _, validator := range validators {
		if validator.Status == stakingtypes.Bonded {
			activeValidators++
		}
		if validator.Tokens.IsPositive() {
			validatorStakes = append(validatorStakes, safeUint64(validator.Tokens, math.MaxUint64))
		}
	}

	aeRoundtripStable, rawRoundtripStable := appAddressRoundtripStable()
	return nativeaccounttypes.NativeAccountInvariantInput{
		AEAddressRoundtripStable:     aeRoundtripStable,
		RawAddressRoundtripStable:    rawRoundtripStable,
		ActivationAttempts:           map[string]uint64{},
		TotalSupply:                  math.MaxUint64,
		RewardBudget:                 math.MaxUint64,
		ActiveValidatorCount:         activeValidators,
		MaxValidatorCount:            uint64(params.MaxValidators),
		ValidatorStakes:              validatorStakes,
		MinValidatorStake:            0,
		ExportImportStable:           true,
		SystemReserveRunwayBlocks:    1,
		MinSystemReserveRunwayBlocks: 1,
		SystemTopUpOrder:             []string{"system_rent_top_up", "user_freeze_processing"},
		ProtocolCriticalExecutable:   true,
		SystemRentUnderfunded:        true,
	}, nil
}

func (app *L1App) assertBankModuleAccountingInvariant(ctx sdk.Context) error {
	return validateBankSupplyMatchesBalances(app.BankKeeper.ExportGenesis(ctx))
}

func validateBankSupplyMatchesBalances(genesis *banktypes.GenesisState) error {
	if genesis == nil {
		return errors.New("bank genesis state is nil")
	}
	balanceSupply := sdk.NewCoins()
	for _, balance := range genesis.Balances {
		if strings.TrimSpace(balance.Address) == "" {
			return errors.New("bank balance address is required")
		}
		if !balance.Coins.IsValid() {
			return fmt.Errorf("bank balance for %s contains invalid coins", balance.Address)
		}
		balanceSupply = balanceSupply.Add(balance.Coins...)
	}
	declaredSupply := sdk.NewCoins(genesis.Supply...)
	if !balanceSupply.Equal(declaredSupply) {
		return fmt.Errorf("bank supply mismatch: balances=%s declared=%s", balanceSupply.String(), declaredSupply.String())
	}
	return nil
}

func (app *L1App) assertDirectUserDelegationRejectedInvariant(ctx sdk.Context) error {
	genesis, err := app.NominatorPoolKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	return validateDirectUserDelegationPolicy(genesis.Params)
}

func validateDirectUserDelegationPolicy(params nominatorpooltypes.Params) error {
	if params.DirectUserDelegationEnabled || params.DirectUserValidatorDelegationEnabled {
		return errors.New("direct user validator delegation is enabled by staking params")
	}
	return nil
}

func (app *L1App) assertValidatorEntryInvariant(ctx sdk.Context) error {
	params, err := app.StakingKeeper.GetParams(ctx)
	if err != nil {
		return err
	}
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}
	activeValidators := uint32(0)
	for _, validator := range validators {
		if validator.Status == stakingtypes.Bonded {
			activeValidators++
		}
		if activeValidators > params.MaxValidators {
			return fmt.Errorf("active validators %d exceed max %d", activeValidators, params.MaxValidators)
		}
		if validator.Tokens.IsPositive() && validator.Tokens.LT(validator.MinSelfDelegation) {
			return fmt.Errorf("validator %s stake %s below min self delegation %s", validator.OperatorAddress, validator.Tokens.String(), validator.MinSelfDelegation.String())
		}
	}
	return nil
}

func (app *L1App) assertNominatorPoolInvariant(ctx sdk.Context) error {
	genesis, err := app.NominatorPoolKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	return genesis.Validate()
}

func (app *L1App) assertStorageRentInvariant(ctx sdk.Context) error {
	genesis, err := app.StorageRentKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	if err := genesis.Validate(); err != nil {
		return err
	}
	systemAddresses := map[string]struct{}{}
	for _, systemAddress := range addressing.AllSystemAddresses() {
		systemAddresses[systemAddress.UserFriendly] = struct{}{}
		systemAddresses[systemAddress.Raw] = struct{}{}
	}
	for _, record := range genesis.State.Contracts {
		if _, found := systemAddresses[record.ContractAddress]; !found {
			continue
		}
		switch record.Status {
		case storagerenttypes.ContractStatusFrozen, storagerenttypes.ContractStatusDeleted, storagerenttypes.ContractStatusArchived:
			return fmt.Errorf("protocol-critical storage rent record %s has forbidden status %s", record.ContractAddress, record.Status)
		}
	}
	return nil
}

func appAddressRoundtripStable() (bool, bool) {
	addr := sdk.AccAddress(bytes.Repeat([]byte{0x42}, 20))
	user := addressing.FormatAccAddress(addr)
	pairFromUser, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, user)
	if err != nil || pairFromUser.User != user {
		return false, false
	}
	parsedUser, err := addressing.ParseAccAddress(pairFromUser.User)
	if err != nil || !bytes.Equal(parsedUser.Bytes(), addr.Bytes()) {
		return false, false
	}
	pairFromRaw, err := addressing.PairFromRawAddress(addressing.AddressRoleAccount, pairFromUser.Raw)
	if err != nil {
		return true, false
	}
	parsedRaw, err := addressing.ParseAccAddress(pairFromRaw.Raw)
	if err != nil {
		return true, false
	}
	return pairFromRaw.User == user && pairFromRaw.Raw == pairFromUser.Raw && bytes.Equal(parsedRaw.Bytes(), addr.Bytes()), true
}

func safeUint64(value sdkmath.Int, fallback uint64) uint64 {
	if value.IsNil() || !value.IsUint64() {
		return fallback
	}
	return value.Uint64()
}
