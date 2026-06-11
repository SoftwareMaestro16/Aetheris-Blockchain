package types

import (
	"errors"
	"fmt"
)

func ApplyExternalMessage(account Account, msg ExternalMessage) (Account, error) {
	if err := ValidateExternalMessage(account, msg); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.Sequence++
	return next, nil
}

func ValidateExternalMessage(account Account, msg ExternalMessage) error {
	if account.Status == AccountStatusInactive {
		return errors.New("inactive account cannot send external messages")
	}
	if account.Status != AccountStatusActive && account.Status != AccountStatusRecovered {
		return fmt.Errorf("%s account cannot send external messages", account.Status)
	}
	if msg.AccountUser != account.AddressUser {
		return errors.New("external message account address mismatch")
	}
	if msg.Sequence != account.Sequence {
		return fmt.Errorf("external message sequence %d does not match account sequence %d", msg.Sequence, account.Sequence)
	}
	return ValidateAuthPolicyForExternalMessage(account, msg)
}

func ValidateAuthPolicyForExternalMessage(account Account, msg ExternalMessage) error {
	_, err := AuthorizeAuthPolicy(account, msg)
	return err
}
