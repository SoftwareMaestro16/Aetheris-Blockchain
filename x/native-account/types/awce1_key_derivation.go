package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

func ValidateAddressPairConsistency(pair addressing.AddressPair) error {
	if pair.User == "" {
		return fmt.Errorf("user-facing address (AE...) must not be empty")
	}
	if pair.Raw == "" {
		return fmt.Errorf("raw internal address (4:...) must not be empty")
	}
	if !strings.HasPrefix(pair.User, addressing.UserFriendlyPrefix) {
		return fmt.Errorf("user-facing address must start with %s", addressing.UserFriendlyPrefix)
	}
	if !strings.HasPrefix(pair.Raw, addressing.RawPrefix) {
		return fmt.Errorf("raw internal address must start with %s", addressing.RawPrefix)
	}
	userBytes, err := addressing.Parse(pair.User)
	if err != nil {
		return fmt.Errorf("user address parse error: %w", err)
	}
	rawBytes, err := addressing.Parse(pair.Raw)
	if err != nil {
		return fmt.Errorf("raw address parse error: %w", err)
	}
	if len(userBytes) != len(rawBytes) {
		return fmt.Errorf("address byte lengths differ: user=%d raw=%d", len(userBytes), len(rawBytes))
	}
	for i := range userBytes {
		if userBytes[i] != rawBytes[i] {
			return fmt.Errorf("AE and 4: addresses do not map to same identity")
		}
	}
	return nil
}

func DeriveAWCE1AddressPairFromPublicKey(pubKeyHex string, keyType string) (addressing.AddressPair, error) {
	if keyType != "secp256k1" && keyType != "secp256k1.PubKey" && keyType != "cosmos.crypto.secp256k1.PubKey" {
		return addressing.AddressPair{}, fmt.Errorf("AWCE1: unsupported key type %q", keyType)
	}
	raw, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return addressing.AddressPair{}, fmt.Errorf("AWCE1: invalid public key hex: %w", err)
	}
	if len(raw) != 33 {
		return addressing.AddressPair{}, fmt.Errorf("AWCE1: secp256k1 public key must be 33 bytes, got %d", len(raw))
	}
	hash := sha256.Sum256(raw)
	addressBytes := hash[:20]
	padded := make([]byte, 32)
	copy(padded[12:], addressBytes)
	rawAddr := addressing.Format(padded)
	userAddr, err := addressing.FormatUserFriendly(padded)
	if err != nil {
		return addressing.AddressPair{}, fmt.Errorf("AWCE1: format user address: %w", err)
	}
	return addressing.AddressPair{
		Role: addressing.AddressRoleAccount,
		User: userAddr,
		Raw:  rawAddr,
	}, nil
}

type KeyRotationRequest struct {
	AccountAddress  string
	NewAuthPolicy   AuthPolicy
	RotationHeight  uint64
	Justification   string
}

type KeyRotationResult struct {
	AccountAddress  string
	PreviousKeyCount int
	NewKeyCount     int
	AddressPreserved bool
	RotationHeight  uint64
}

func ValidateKeyRotation(account Account, request KeyRotationRequest) (*KeyRotationResult, error) {
	if err := ValidateAddressPairConsistency(addressing.AddressPair{
		Role: addressing.AddressRoleAccount,
		User: account.AddressUser,
		Raw:  account.AddressRaw,
	}); err != nil {
		return nil, fmt.Errorf("AWCE1 key rotation: address pair invalid: %w", err)
	}
	if err := ValidateNoSecretLikeText(request.Justification); err != nil {
		return nil, fmt.Errorf("AWCE1 key rotation: justification: %w", err)
	}
	if request.RotationHeight < account.CreatedHeight {
		return nil, fmt.Errorf("AWCE1 key rotation: rotation height %d before account creation %d",
			request.RotationHeight, account.CreatedHeight)
	}
	return &KeyRotationResult{
		AccountAddress:   account.AddressUser,
		PreviousKeyCount: len(account.AuthPolicy.Keys),
		NewKeyCount:      len(request.NewAuthPolicy.Keys),
		AddressPreserved: true,
		RotationHeight:   request.RotationHeight,
	}, nil
}
