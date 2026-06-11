package types

import (
	"fmt"
	"sort"
)

func DefaultFeatureFlags(version uint64) ([]string, error) {
	switch version {
	case AccountVersionV1:
		return nil, nil
	case AccountVersionV2:
		return []string{
			AccountFeatureInternalMessagesV2,
			AccountFeatureMetadataV2,
			AccountFeatureRecoveryPolicyV2,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported native account version %d", version)
	}
}

func MigrateAccountIfNeeded(account Account) (Account, error) {
	switch account.Version {
	case AccountVersionV1:
		return MigrateAccountV1ToV2(account)
	case AccountVersionV2:
		return normalizeAccount(account), ValidateAccountInvariant(account)
	default:
		return Account{}, fmt.Errorf("unsupported native account version %d", account.Version)
	}
}

func MigrateAccountV1ToV2(account Account) (Account, error) {
	if account.Version != AccountVersionV1 {
		return Account{}, fmt.Errorf("native account v1 to v2 migration requires version %d", AccountVersionV1)
	}
	if err := ValidateAccountInvariant(account); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.Version = AccountVersionV2
	if len(next.FeatureFlags) == 0 {
		defaults, err := DefaultFeatureFlags(AccountVersionV2)
		if err != nil {
			return Account{}, err
		}
		next.FeatureFlags = defaults
	}
	next = normalizeAccount(next)
	if err := ValidateAccountInvariant(next); err != nil {
		return Account{}, err
	}
	return next, nil
}

func normalizeAccount(account Account) Account {
	account.PubKeys = cloneStrings(account.PubKeys)
	account.FeatureFlags = cloneStrings(account.FeatureFlags)
	account.AuthPolicy = account.AuthPolicy.Normalize()
	sort.Strings(account.PubKeys)
	sort.Strings(account.FeatureFlags)
	return account
}

func cloneAccount(account Account) Account {
	account.PubKeys = cloneStrings(account.PubKeys)
	account.FeatureFlags = cloneStrings(account.FeatureFlags)
	account.AuthPolicy = account.AuthPolicy.Normalize()
	return account
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}
