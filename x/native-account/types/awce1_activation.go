package types

import (
	"errors"
	"fmt"

	"github.com/sovereign-l1/l1/app/addressing"
)

var (
	ErrAccountAlreadyActive  = errors.New("AWCE1: account already active — duplicate activation rejected")
	ErrAccountNotFound       = errors.New("AWCE1: account not found")
	ErrAccountInactive       = errors.New("AWCE1: account is inactive")
	ErrAccountFrozen         = errors.New("AWCE1: account is frozen")
	ErrAccountClosed         = errors.New("AWCE1: account is closed")
	ErrInvalidLifecycleState = errors.New("AWCE1: invalid account lifecycle state")
	ErrAddressMismatch       = errors.New("AWCE1: address pair mismatch")
	ErrSecretInAccount       = errors.New("AWCE1: secret-like text found in account")
)

func ValidateActivation(account Account, pair addressing.AddressPair) error {
	if account.Status == AccountStatusActive {
		return ErrAccountAlreadyActive
	}
	if account.Status != AccountStatusInactive && account.Status != "" {
		return fmt.Errorf("AWCE1: cannot activate account in status %q", account.Status)
	}
	if account.AddressUser != "" && account.AddressUser != pair.User {
		return ErrAddressMismatch
	}
	if account.AddressRaw != "" && account.AddressRaw != pair.Raw {
		return ErrAddressMismatch
	}
	if err := ValidateAccountNoSecrets(account); err != nil {
		return ErrSecretInAccount
	}
	return nil
}

func ActivateAccount(pair addressing.AddressPair, height uint64) (*Account, error) {
	if err := ValidateAddressPairConsistency(pair); err != nil {
		return nil, fmt.Errorf("AWCE1 activation: %w", err)
	}
	features, _ := DefaultFeatureFlags(CurrentAccountVersion)
	account := &Account{
		Version:          CurrentAccountVersion,
		AddressUser:      pair.User,
		AddressRaw:       pair.Raw,
		Sequence:         ActivationInitialSequence,
		Status:           AccountStatusActive,
		AuthPolicy:       DefaultAuthPolicy(),
		FeatureFlags:     features,
		CreatedHeight:    height,
		LastActiveHeight: height,
		StorageRentDebt:  0,
	}
	return account, nil
}

func DefaultAuthPolicy() AuthPolicy {
	return AuthPolicy{
		Version:   1,
		Mode:      "single_key",
		Threshold: 1,
	}
}
