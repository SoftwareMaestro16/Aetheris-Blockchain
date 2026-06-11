package types

import (
	"fmt"
	"regexp"

	"github.com/sovereign-l1/l1/app/addressing"
)

// ---------------
// Blocker 6: AWCE-1 Wallet Compatibility Extension
// ---------------
// Aetra Wallet Compatibility Extension = Cosmos wallet signing compatibility
// + Aetra native account identity/runtime metadata
// + AE <-> 4: dual-address abstraction
// + policy-controlled account lifecycle

const (
	AWCE1Version = uint32(1)

	BIP44Purpose       = uint32(44)
	BIP44CoinType      = uint32(118)
	BIP44Account       = uint32(0)
	BIP44Change        = uint32(0)
	BIP44AddressIndex  = uint32(0)

	BIP44FullPath = "m/44'/118'/0'/0/0"

	SignDocTypeURL = "/cosmos.tx.v1beta1.Tx"

	AWCE1FeatureKeyRotationV2    = "key_rotation_v2"
	AWCE1FeatureStorageRentV2    = "storage_rent_v2"
	AWCE1FeatureRecoveryPolicyV2 = "recovery_policy_v2"
	AWCE1FeatureAuthPolicyV2     = "auth_policy_v2"

	MaxAWCE1FeatureFlags       = 16
	MaxAWCE1SecretPatternBytes = 256
)

var (
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)private.?key`),
		regexp.MustCompile(`(?i)mnemonic`),
		regexp.MustCompile(`(?i)seed.?phrase`),
		regexp.MustCompile(`(?i)secret`),
		regexp.MustCompile(`(?i)password`),
		regexp.MustCompile(`(?i)passphrase`),
		regexp.MustCompile(`(?i)privkey`),
		regexp.MustCompile(`(?i)priv_key`),
	}
)

// ---------------
// AWCE-1 Wallet Compatibility Profile
// ---------------

type AWCE1WalletProfile struct {
	Version          uint32
	SpecVersion      uint32
	AddressPair      addressing.AddressPair
	BIP44Path        string
	KeyType          string
	Features         []string
	AccountStatus    string
	ActivationHeight uint64
}

func NewAWCE1WalletProfile(pair addressing.AddressPair) *AWCE1WalletProfile {
	features, _ := DefaultFeatureFlags(AccountVersionV2)
	return &AWCE1WalletProfile{
		Version:          AWCE1Version,
		SpecVersion:      AWCE1Version,
		AddressPair:      pair,
		BIP44Path:        BIP44FullPath,
		KeyType:          "secp256k1",
		Features:         features,
		AccountStatus:    AccountStatusInactive,
		ActivationHeight: 0,
	}
}

func (p *AWCE1WalletProfile) Validate() error {
	if p.Version == 0 {
		return fmt.Errorf("AWCE1: version must be > 0")
	}
	if p.AddressPair.User == "" || p.AddressPair.Raw == "" {
		return fmt.Errorf("AWCE1: address pair must have both User and Raw")
	}
	if err := ValidateAddressPairConsistency(p.AddressPair); err != nil {
		return fmt.Errorf("AWCE1: address pair inconsistent: %w", err)
	}
	if err := ValidateNoSecretLikeText(p.BIP44Path); err != nil {
		return fmt.Errorf("AWCE1: BIP44 path: %w", err)
	}
	if p.KeyType != "secp256k1" {
		return fmt.Errorf("AWCE1: unsupported key type %q; only secp256k1 is supported for AWCE-1", p.KeyType)
	}
	if len(p.Features) > MaxAWCE1FeatureFlags {
		return fmt.Errorf("AWCE1: too many feature flags (%d > %d)", len(p.Features), MaxAWCE1FeatureFlags)
	}
	return nil
}

// ---------------
// AE <-> 4: Dual-Address Verification
// ---------------

// ---------------
// Account Lifecycle State Machine
// ---------------

type AccountLifecycleTransition struct {
	From string
	To   string
}

var ValidLifecycleTransitions = []AccountLifecycleTransition{
	{AccountStatusInactive, AccountStatusActive},
	{AccountStatusActive, AccountStatusFrozen},
	{AccountStatusActive, AccountStatusRecovered},
	{AccountStatusActive, AccountStatusArchived},
	{AccountStatusFrozen, AccountStatusActive},
	{AccountStatusFrozen, AccountStatusArchived},
	{AccountStatusRecovered, AccountStatusActive},
	{AccountStatusRecovered, AccountStatusArchived},
	{AccountStatusArchived, AccountStatusClosed},
}

func ValidateLifecycleTransition(from, to string) error {
	if from == to {
		return nil
	}
	for _, t := range ValidLifecycleTransitions {
		if t.From == from && t.To == to {
			return nil
		}
	}
	return fmt.Errorf("AWCE1: invalid lifecycle transition %s -> %s", from, to)
}

func CanActivate(status string) bool {
	return status == AccountStatusInactive
}

func CanTransact(status string) bool {
	return status == AccountStatusActive || status == AccountStatusRecovered
}

func IsTerminalStatus(status string) bool {
	return status == AccountStatusClosed
}

// ---------------
// Secret-Like Text Detection
// ---------------

func ValidateNoSecretLikeText(text string) error {
	for _, pattern := range secretPatterns {
		if pattern.MatchString(text) {
			return fmt.Errorf("AWCE1: text contains secret-like pattern matching %q", pattern.String())
		}
	}
	return nil
}

func ValidateAccountNoSecrets(account Account) error {
	for _, pk := range account.PubKeys {
		if err := ValidateNoSecretLikeText(pk); err != nil {
			return fmt.Errorf("public key field: %w", err)
		}
	}
	if err := ValidateNoSecretLikeText(account.ReputationID); err != nil {
		return fmt.Errorf("reputation ID: %w", err)
	}
	if err := ValidateNoSecretLikeText(account.Metadata.MetadataHash); err != nil {
		return fmt.Errorf("metadata hash: %w", err)
	}
	if err := ValidateNoSecretLikeText(account.Metadata.DisplayNameHash); err != nil {
		return fmt.Errorf("display name hash: %w", err)
	}
	if err := ValidateNoSecretLikeText(account.Metadata.DomainAlias); err != nil {
		return fmt.Errorf("domain alias: %w", err)
	}
	return nil
}

