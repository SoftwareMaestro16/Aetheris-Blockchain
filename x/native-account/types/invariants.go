package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	InvariantNoPrivateKeyOnChain                         = "no_private_key_on_chain"
	InvariantNoSeedPhraseOnChain                         = "no_seed_phrase_on_chain"
	InvariantAEAddressRoundtripStable                    = "ae_address_roundtrip_stable"
	InvariantRawAddressRoundtripStable                   = "raw_4_address_roundtrip_stable"
	InvariantAccountActivationIdempotency                = "account_activation_idempotency_enforced"
	InvariantAccountCannotActivateTwice                  = "account_cannot_activate_twice"
	InvariantPoolStakeDoesNotCreateCoins                 = "pool_stake_does_not_create_coins"
	InvariantRewardsCannotExceedAllocation               = "rewards_cannot_exceed_emissions_or_fees"
	InvariantMaxValidatorCountEnforced                   = "max_validator_count_enforced"
	InvariantMinValidatorStakeEnforced                   = "min_validator_stake_enforced"
	InvariantValidatorSelfStakeRatioEnforced             = "validator_self_stake_ratio_enforced"
	InvariantMinPoolDepositEnforced                      = "min_pool_deposit_enforced"
	InvariantDirectUserValidatorDelegationRejected       = "direct_user_validator_delegation_rejected"
	InvariantUnbondingCannotReleaseEarly                 = "unbonding_cannot_release_early"
	InvariantReputationRequiresStakeTime                 = "reputation_cannot_increase_without_stake_time"
	InvariantJailedValidatorNoPositiveBonus              = "jailed_validator_no_positive_bonus"
	InvariantExportImportPreservesState                  = "export_import_preserves_state"
	InvariantModuleBankAccountingConsistent              = "module_bank_accounting_consistent"
	InvariantProtocolCriticalStateNotFrozenByRent        = "protocol_critical_state_not_frozen_by_rent"
	InvariantSystemStorageReserveRunway                  = "system_storage_reserve_runway"
	InvariantSystemRentTopUpBeforeUserFreeze             = "system_rent_topup_before_user_freeze"
	InvariantProtocolCriticalExecutableUnderUnderfunding = "protocol_critical_executable_under_system_rent_underfunding"
)

type NativeAccountInvariant struct {
	ID          string
	Description string
}

type NativeAccountInvariantInput struct {
	ExportedPayloads []string
	Accounts         []Account

	AEAddressRoundtripStable  bool
	RawAddressRoundtripStable bool
	ActivationAttempts        map[string]uint64

	PoolActiveStake    uint64
	PoolUnbonding      uint64
	ValidatorSelfStake uint64
	LiquidBalances     uint64
	TotalSupply        uint64

	RewardsAccrued uint64
	RewardBudget   uint64

	ActiveValidatorCount uint64
	MaxValidatorCount    uint64
	ValidatorStakes      []uint64
	MinValidatorStake    uint64
	ValidatorTotalStake  uint64
	MinSelfStakeBps      uint64

	PoolDeposits             []uint64
	MinPoolDeposit           uint64
	DirectUserDelegationSeen bool

	UnbondingEntries []InvariantUnbondingEntry

	StakeTimeDelta    uint64
	ReputationDelta   uint64
	JailedBonusAmount uint64

	ExportImportStable bool

	BankAccountingScenarios []InvariantAccountingScenario

	ProtocolCriticalFrozenByRent bool
	SystemReserveRunwayBlocks    uint64
	MinSystemReserveRunwayBlocks uint64
	SystemReserveAlertRaised     bool
	SystemTopUpOrder             []string
	ProtocolCriticalExecutable   bool
	SystemRentUnderfunded        bool
}

type InvariantUnbondingEntry struct {
	ReleaseHeight uint64
	CurrentHeight uint64
	Released      bool
}

type InvariantAccountingScenario struct {
	Name                 string
	BankBalance          uint64
	ModuleAccounting     uint64
	ExpectedSupplyChange int64
	ActualSupplyChange   int64
}

type NativeAccountInvariantFailure struct {
	ID    string
	Error string
}

func RequiredNativeAccountInvariantIDs() []string {
	return []string{
		InvariantNoPrivateKeyOnChain,
		InvariantNoSeedPhraseOnChain,
		InvariantAEAddressRoundtripStable,
		InvariantRawAddressRoundtripStable,
		InvariantAccountActivationIdempotency,
		InvariantAccountCannotActivateTwice,
		InvariantPoolStakeDoesNotCreateCoins,
		InvariantRewardsCannotExceedAllocation,
		InvariantMaxValidatorCountEnforced,
		InvariantMinValidatorStakeEnforced,
		InvariantValidatorSelfStakeRatioEnforced,
		InvariantMinPoolDepositEnforced,
		InvariantDirectUserValidatorDelegationRejected,
		InvariantUnbondingCannotReleaseEarly,
		InvariantReputationRequiresStakeTime,
		InvariantJailedValidatorNoPositiveBonus,
		InvariantExportImportPreservesState,
		InvariantModuleBankAccountingConsistent,
		InvariantProtocolCriticalStateNotFrozenByRent,
		InvariantSystemStorageReserveRunway,
		InvariantSystemRentTopUpBeforeUserFreeze,
		InvariantProtocolCriticalExecutableUnderUnderfunding,
	}
}

func DefaultNativeAccountInvariantRegistry() []NativeAccountInvariant {
	return []NativeAccountInvariant{
		{InvariantNoPrivateKeyOnChain, "private keys must never appear in account, auth, genesis, event, or export payloads"},
		{InvariantNoSeedPhraseOnChain, "seed phrases and mnemonics must never appear on-chain"},
		{InvariantAEAddressRoundtripStable, "AE user-facing address roundtrip remains stable"},
		{InvariantRawAddressRoundtripStable, "4: raw address roundtrip remains stable"},
		{InvariantAccountActivationIdempotency, "account activation is idempotency-safe"},
		{InvariantAccountCannotActivateTwice, "activated accounts cannot be activated a second time"},
		{InvariantPoolStakeDoesNotCreateCoins, "pool active stake, unbonding, validator self stake, and liquid balances conserve supply"},
		{InvariantRewardsCannotExceedAllocation, "rewards cannot exceed emissions and fee allocation"},
		{InvariantMaxValidatorCountEnforced, "MaxValidatorCount is enforced"},
		{InvariantMinValidatorStakeEnforced, "MinValidatorStake is enforced"},
		{InvariantValidatorSelfStakeRatioEnforced, "validator self-stake ratio is enforced"},
		{InvariantMinPoolDepositEnforced, "MinPoolDeposit is enforced"},
		{InvariantDirectUserValidatorDelegationRejected, "direct user validator delegation is rejected"},
		{InvariantUnbondingCannotReleaseEarly, "unbonding cannot release before maturity"},
		{InvariantReputationRequiresStakeTime, "reputation cannot increase without stake-time"},
		{InvariantJailedValidatorNoPositiveBonus, "jailed validators cannot receive positive validator bonus"},
		{InvariantExportImportPreservesState, "export/import preserves all consensus state classes"},
		{InvariantModuleBankAccountingConsistent, "module account and bank accounting remain consistent"},
		{InvariantProtocolCriticalStateNotFrozenByRent, "protocol-critical state is not frozen by storage rent"},
		{InvariantSystemStorageReserveRunway, "system storage reserve has runway or raises an alert"},
		{InvariantSystemRentTopUpBeforeUserFreeze, "system rent top-up runs before user freeze processing"},
		{InvariantProtocolCriticalExecutableUnderUnderfunding, "protocol-critical modules remain executable during system rent underfunding"},
	}
}

func ValidateNativeAccountInvariantRegistry(registry []NativeAccountInvariant) error {
	required := RequiredNativeAccountInvariantIDs()
	seen := make(map[string]struct{}, len(registry))
	for _, invariant := range registry {
		if strings.TrimSpace(invariant.ID) == "" || strings.TrimSpace(invariant.Description) == "" {
			return errors.New("native account invariant id and description are required")
		}
		if _, found := seen[invariant.ID]; found {
			return fmt.Errorf("duplicate native account invariant %s", invariant.ID)
		}
		seen[invariant.ID] = struct{}{}
	}
	for _, id := range required {
		if _, found := seen[id]; !found {
			return fmt.Errorf("native account invariant registry missing %s", id)
		}
	}
	return nil
}

func RunAllNativeAccountInvariants(input NativeAccountInvariantInput) []NativeAccountInvariantFailure {
	registry := DefaultNativeAccountInvariantRegistry()
	failures := make([]NativeAccountInvariantFailure, 0)
	for _, invariant := range registry {
		if err := RunNativeAccountInvariant(invariant.ID, input); err != nil {
			failures = append(failures, NativeAccountInvariantFailure{ID: invariant.ID, Error: err.Error()})
		}
	}
	sort.SliceStable(failures, func(i, j int) bool { return failures[i].ID < failures[j].ID })
	return failures
}

func RunNativeAccountInvariant(id string, input NativeAccountInvariantInput) error {
	switch id {
	case InvariantNoPrivateKeyOnChain:
		return rejectSecretPayload(input, "private key", "private_key")
	case InvariantNoSeedPhraseOnChain:
		return rejectSecretPayload(input, "seed phrase", "seed_phrase", "mnemonic")
	case InvariantAEAddressRoundtripStable:
		if !input.AEAddressRoundtripStable {
			return errors.New("AE address roundtrip is not stable")
		}
	case InvariantRawAddressRoundtripStable:
		if !input.RawAddressRoundtripStable {
			return errors.New("raw 4: address roundtrip is not stable")
		}
	case InvariantAccountActivationIdempotency, InvariantAccountCannotActivateTwice:
		for account, attempts := range input.ActivationAttempts {
			if attempts > 1 {
				return fmt.Errorf("account %s activated %d times", account, attempts)
			}
		}
	case InvariantPoolStakeDoesNotCreateCoins:
		total := input.PoolActiveStake + input.PoolUnbonding + input.ValidatorSelfStake + input.LiquidBalances
		if total > input.TotalSupply {
			return fmt.Errorf("pool accounting creates coins: accounted %d exceeds supply %d", total, input.TotalSupply)
		}
	case InvariantRewardsCannotExceedAllocation:
		if input.RewardsAccrued > input.RewardBudget {
			return fmt.Errorf("rewards %d exceed budget %d", input.RewardsAccrued, input.RewardBudget)
		}
	case InvariantMaxValidatorCountEnforced:
		if input.ActiveValidatorCount > input.MaxValidatorCount {
			return fmt.Errorf("active validators %d exceed max %d", input.ActiveValidatorCount, input.MaxValidatorCount)
		}
	case InvariantMinValidatorStakeEnforced:
		for _, stake := range input.ValidatorStakes {
			if stake < input.MinValidatorStake {
				return fmt.Errorf("validator stake %d below min %d", stake, input.MinValidatorStake)
			}
		}
	case InvariantValidatorSelfStakeRatioEnforced:
		if input.ValidatorTotalStake > 0 && input.ValidatorSelfStake*10_000 < input.ValidatorTotalStake*input.MinSelfStakeBps {
			return errors.New("validator self-stake ratio below minimum")
		}
	case InvariantMinPoolDepositEnforced:
		for _, deposit := range input.PoolDeposits {
			if deposit < input.MinPoolDeposit {
				return fmt.Errorf("pool deposit %d below minimum %d", deposit, input.MinPoolDeposit)
			}
		}
	case InvariantDirectUserValidatorDelegationRejected:
		if input.DirectUserDelegationSeen {
			return errors.New("direct user validator delegation observed")
		}
	case InvariantUnbondingCannotReleaseEarly:
		for _, entry := range input.UnbondingEntries {
			if entry.Released && entry.CurrentHeight < entry.ReleaseHeight {
				return errors.New("unbonding released before maturity")
			}
		}
	case InvariantReputationRequiresStakeTime:
		if input.ReputationDelta > 0 && input.StakeTimeDelta == 0 {
			return errors.New("reputation increased without stake-time")
		}
	case InvariantJailedValidatorNoPositiveBonus:
		if input.JailedBonusAmount > 0 {
			return errors.New("jailed validator received positive bonus")
		}
	case InvariantExportImportPreservesState:
		if !input.ExportImportStable {
			return errors.New("export/import did not preserve full state")
		}
	case InvariantModuleBankAccountingConsistent:
		for _, scenario := range input.BankAccountingScenarios {
			if scenario.BankBalance != scenario.ModuleAccounting {
				return fmt.Errorf("%s bank balance %d differs from module accounting %d", scenario.Name, scenario.BankBalance, scenario.ModuleAccounting)
			}
			if scenario.ExpectedSupplyChange != scenario.ActualSupplyChange {
				return fmt.Errorf("%s supply delta mismatch", scenario.Name)
			}
		}
	case InvariantProtocolCriticalStateNotFrozenByRent:
		if input.ProtocolCriticalFrozenByRent {
			return errors.New("protocol-critical state frozen by storage rent")
		}
	case InvariantSystemStorageReserveRunway:
		if input.SystemReserveRunwayBlocks < input.MinSystemReserveRunwayBlocks && !input.SystemReserveAlertRaised {
			return errors.New("system storage reserve runway low without invariant alert")
		}
	case InvariantSystemRentTopUpBeforeUserFreeze:
		if !topUpBeforeFreeze(input.SystemTopUpOrder) {
			return errors.New("system rent top-up did not run before user freeze processing")
		}
	case InvariantProtocolCriticalExecutableUnderUnderfunding:
		if input.SystemRentUnderfunded && !input.ProtocolCriticalExecutable {
			return errors.New("protocol-critical module not executable during system rent underfunding")
		}
	default:
		return fmt.Errorf("unknown native account invariant %s", id)
	}
	return nil
}

func rejectSecretPayload(input NativeAccountInvariantInput, needles ...string) error {
	payloads := append([]string(nil), input.ExportedPayloads...)
	for _, account := range input.Accounts {
		if err := ValidateAccountInvariant(account); err != nil {
			return err
		}
		payloads = append(payloads, account.AuthPolicy.Mode)
		payloads = append(payloads, account.Metadata.MetadataHash, account.Metadata.DisplayNameHash, account.Metadata.DomainAlias)
		payloads = append(payloads, account.PubKeys...)
	}
	for _, payload := range payloads {
		lower := strings.ToLower(payload)
		for _, needle := range needles {
			if strings.Contains(lower, needle) {
				return fmt.Errorf("secret-like payload rejected: %s", needle)
			}
		}
	}
	return nil
}

func topUpBeforeFreeze(order []string) bool {
	topUp := -1
	freeze := -1
	for idx, item := range order {
		switch item {
		case "system_rent_top_up":
			topUp = idx
		case "user_freeze_processing":
			freeze = idx
		}
	}
	return topUp >= 0 && freeze >= 0 && topUp < freeze
}
