package types

import (
	"errors"
	"fmt"
	"strings"
)

func ApplyInternalMessage(account Account, msg InternalMessage, policy InternalMessagePolicy) (Account, error) {
	if err := ValidateInternalMessage(account, msg, policy); err != nil {
		return Account{}, err
	}
	return cloneAccount(account), nil
}

func ValidateInternalMessage(account Account, msg InternalMessage, policy InternalMessagePolicy) error {
	if msg.AccountUser != account.AddressUser {
		return errors.New("internal message account address mismatch")
	}
	if account.Status == AccountStatusInactive || account.Status == AccountStatusClosed || account.Status == AccountStatusArchived {
		return fmt.Errorf("%s account cannot receive internal messages", account.Status)
	}
	if account.Status == AccountStatusFrozen && !msg.WhitelistedWhileFrozen {
		return errors.New("frozen account internal messages require explicit whitelist")
	}
	if policy.Version == 0 {
		return errors.New("internal message policy version must be positive")
	}
	if strings.TrimSpace(policy.EnabledFeature) == "" {
		return errors.New("internal message policy feature is required")
	}
	if strings.TrimSpace(msg.Feature) == "" || msg.Feature != policy.EnabledFeature {
		return errors.New("internal message feature is not enabled by policy")
	}
	if !accountHasFeature(account, msg.Feature) {
		return errors.New("internal message feature disabled on account")
	}
	switch msg.Source {
	case InternalMessageSourceModule, InternalMessageSourceContract, InternalMessageSourceSystem:
		return nil
	default:
		return fmt.Errorf("unsupported internal message source %q", msg.Source)
	}
}

func accountHasFeature(account Account, feature string) bool {
	for _, existing := range account.FeatureFlags {
		if existing == feature {
			return true
		}
	}
	return false
}
