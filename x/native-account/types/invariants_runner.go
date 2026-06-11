package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

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
