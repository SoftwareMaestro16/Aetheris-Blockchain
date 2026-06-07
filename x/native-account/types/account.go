package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	AccountVersionV1      = uint64(1)
	AccountVersionV2      = uint64(2)
	CurrentAccountVersion = AccountVersionV2

	AccountStatusActive    = "active"
	AccountStatusFrozen    = "frozen"
	AccountStatusRecovered = "recovered"
	AccountStatusArchived  = "archived"
	AccountStatusClosed    = "closed"

	AccountFeatureInternalMessagesV2 = "internal_messages_v2"
	AccountFeatureRecoveryPolicyV2   = "recovery_policy_v2"
	AccountFeatureMetadataV2         = "metadata_v2"

	MaxMetadataHashBytes    = 128
	MaxDisplayNameHashBytes = 128
	MaxDomainAliasBytes     = 253
	MaxPubKeyTextBytes      = 256
)

type Account struct {
	Version                 uint64
	AddressUser             string
	AddressRaw              string
	PubKeys                 []string
	AccountNumber           uint64
	Sequence                uint64
	Status                  string
	AuthPolicy              AuthPolicy
	FeatureFlags            []string
	Metadata                AccountMetadata
	ReputationID            string
	CreatedHeight           uint64
	LastActiveHeight        uint64
	LastStorageChargeHeight uint64
	StorageRentDebt         uint64
}

type AuthPolicy struct {
	Version uint64
	Mode    string
}

type AccountMetadata struct {
	MetadataHash    string
	DisplayNameHash string
	DomainAlias     string
	CreatedHeight   uint64
}

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

func ValidateAccountInvariant(account Account) error {
	if account.Version != AccountVersionV1 && account.Version != AccountVersionV2 {
		return fmt.Errorf("unsupported native account version %d", account.Version)
	}
	if err := ValidateUserFacingAEAddress("native account user address", account.AddressUser); err != nil {
		return err
	}
	if err := ValidateRawAddress("native account raw address", account.AddressRaw); err != nil {
		return err
	}
	if err := ValidateAddressPair("native account address pair", account.AddressUser, account.AddressRaw); err != nil {
		return err
	}
	if !isAccountStatus(account.Status) {
		return fmt.Errorf("unsupported native account status %q", account.Status)
	}
	if account.CreatedHeight == 0 {
		return errors.New("native account created height must be positive")
	}
	if account.LastActiveHeight != 0 && account.LastActiveHeight < account.CreatedHeight {
		return errors.New("native account last active height must not precede creation")
	}
	if account.LastStorageChargeHeight != 0 && account.LastStorageChargeHeight < account.CreatedHeight {
		return errors.New("native account last storage charge height must not precede creation")
	}
	if err := validatePubKeys(account.PubKeys); err != nil {
		return err
	}
	if err := account.Metadata.Validate(); err != nil {
		return err
	}
	if err := validateFeatureFlags(account.Version, account.FeatureFlags); err != nil {
		return err
	}
	return nil
}

func (m AccountMetadata) Validate() error {
	if len(m.MetadataHash) > MaxMetadataHashBytes {
		return fmt.Errorf("native account metadata hash exceeds %d bytes", MaxMetadataHashBytes)
	}
	if len(m.DisplayNameHash) > MaxDisplayNameHashBytes {
		return fmt.Errorf("native account display name hash exceeds %d bytes", MaxDisplayNameHashBytes)
	}
	if len(m.DomainAlias) > MaxDomainAliasBytes {
		return fmt.Errorf("native account domain alias exceeds %d bytes", MaxDomainAliasBytes)
	}
	for _, value := range []string{m.MetadataHash, m.DisplayNameHash, m.DomainAlias} {
		if containsSecretLikeText(value) {
			return errors.New("native account metadata must not contain private keys or seed phrases")
		}
	}
	return nil
}

func ValidateUserFacingAEAddress(field, text string) error {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, addressing.UserFriendlyPrefix) {
		return fmt.Errorf("%s must use AE user-facing address format", field)
	}
	return addressing.ValidateUserAddress(field, text)
}

func ValidateRawAddress(field, text string) error {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, addressing.RawPrefix) {
		return fmt.Errorf("%s must use 4: raw address format", field)
	}
	_, err := addressing.Parse(text)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", field, err)
	}
	return nil
}

func ValidateAddressPair(field, userAddress, rawAddress string) error {
	userBytes, err := addressing.Parse(userAddress)
	if err != nil {
		return fmt.Errorf("invalid %s user address: %w", field, err)
	}
	rawBytes, err := addressing.Parse(rawAddress)
	if err != nil {
		return fmt.Errorf("invalid %s raw address: %w", field, err)
	}
	userKey, err := addressing.AddressTextBytesKey(userAddress)
	if err != nil {
		return err
	}
	rawKey, err := addressing.AddressTextBytesKey(rawAddress)
	if err != nil {
		return err
	}
	if userKey != rawKey || string(userBytes) != string(rawBytes) {
		return fmt.Errorf("%s AE and raw addresses must represent the same account", field)
	}
	return nil
}

func normalizeAccount(account Account) Account {
	account.PubKeys = cloneStrings(account.PubKeys)
	account.FeatureFlags = cloneStrings(account.FeatureFlags)
	sort.Strings(account.PubKeys)
	sort.Strings(account.FeatureFlags)
	return account
}

func cloneAccount(account Account) Account {
	account.PubKeys = cloneStrings(account.PubKeys)
	account.FeatureFlags = cloneStrings(account.FeatureFlags)
	return account
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func validateFeatureFlags(version uint64, flags []string) error {
	if version == AccountVersionV1 && len(flags) != 0 {
		return errors.New("native account v1 must not carry v2 feature flags")
	}
	previous := ""
	for _, flag := range flags {
		if strings.TrimSpace(flag) == "" {
			return errors.New("native account feature flag is required")
		}
		if flag <= previous {
			return errors.New("native account feature flags must be sorted and unique")
		}
		previous = flag
	}
	return nil
}

func validatePubKeys(pubKeys []string) error {
	previous := ""
	for _, pubKey := range pubKeys {
		pubKey = strings.TrimSpace(pubKey)
		if pubKey == "" {
			return errors.New("native account pubkey is required")
		}
		if len(pubKey) > MaxPubKeyTextBytes {
			return fmt.Errorf("native account pubkey exceeds %d bytes", MaxPubKeyTextBytes)
		}
		if containsSecretLikeText(pubKey) {
			return errors.New("native account pubkeys must not contain private keys or seed phrases")
		}
		if pubKey <= previous {
			return errors.New("native account pubkeys must be sorted and unique")
		}
		previous = pubKey
	}
	return nil
}

func containsSecretLikeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "private key") ||
		strings.Contains(lower, "private_key") ||
		strings.Contains(lower, "seed phrase") ||
		strings.Contains(lower, "seed_phrase") ||
		strings.Contains(lower, "mnemonic")
}

func isAccountStatus(status string) bool {
	return status == AccountStatusActive ||
		status == AccountStatusFrozen ||
		status == AccountStatusRecovered ||
		status == AccountStatusArchived ||
		status == AccountStatusClosed
}
