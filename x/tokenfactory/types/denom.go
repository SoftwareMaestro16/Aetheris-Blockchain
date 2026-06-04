package types

import (
	"fmt"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MinSubdenomLength = 3
	MaxSubdenomLength = 64

	NativeBaseDenom    = "norb"
	NativeDisplayDenom = "orb"

	DefaultDenomQueryLimit = 100
	MaxDenomQueryLimit     = 500
)

var subdenomRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]{2,63}$`)

func ValidateSubdenom(subdenom string) error {
	if subdenom != strings.TrimSpace(subdenom) {
		return ErrInvalidDenom.Wrap("subdenom must not contain leading or trailing whitespace")
	}
	if len(subdenom) < MinSubdenomLength || len(subdenom) > MaxSubdenomLength {
		return ErrInvalidDenom.Wrapf("subdenom length must be %d-%d chars", MinSubdenomLength, MaxSubdenomLength)
	}
	if !subdenomRe.MatchString(subdenom) {
		return ErrInvalidDenom.Wrap("subdenom must start with a letter and contain only letters, digits, dot, underscore, or dash")
	}

	lower := strings.ToLower(subdenom)
	if lower == NativeBaseDenom || lower == NativeDisplayDenom {
		return ErrInvalidDenom.Wrap("subdenom is reserved for the native token")
	}
	if lower == "lp" || strings.HasPrefix(lower, "lp-") || strings.HasPrefix(lower, "lp_") || strings.HasPrefix(lower, "lp.") {
		return ErrInvalidDenom.Wrap("subdenom is reserved for DEX LP tokens")
	}

	return nil
}

func BuildFactoryDenom(creator, subdenom string) (string, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	if err != nil {
		return "", ErrInvalidAddress.Wrapf("invalid creator address: %v", err)
	}
	if err := ValidateSubdenom(subdenom); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", FactoryDenomPrefix, creatorAddr.String(), subdenom), nil
}

func ValidateFactoryDenom(denom string) error {
	_, _, err := ParseFactoryDenom(denom)
	return err
}

func ParseFactoryDenom(denom string) (sdk.AccAddress, string, error) {
	if err := sdk.ValidateDenom(denom); err != nil {
		return nil, "", ErrInvalidDenom.Wrapf("invalid denom %s: %v", denom, err)
	}

	parts := strings.Split(denom, "/")
	if len(parts) != 3 || parts[0] != FactoryDenomPrefix {
		return nil, "", ErrInvalidDenom.Wrapf("denom must have form %s/{creator}/{subdenom}", FactoryDenomPrefix)
	}

	creator, err := sdk.AccAddressFromBech32(parts[1])
	if err != nil {
		return nil, "", ErrInvalidAddress.Wrapf("invalid denom creator address: %v", err)
	}
	if err := ValidateSubdenom(parts[2]); err != nil {
		return nil, "", err
	}

	expected, err := BuildFactoryDenom(creator.String(), parts[2])
	if err != nil {
		return nil, "", err
	}
	if denom != expected {
		return nil, "", ErrInvalidDenom.Wrap("denom creator address must be canonical")
	}

	return creator, parts[2], nil
}

func NormalizeDenomAuthorityMetadata(meta DenomAuthorityMetadata) (DenomAuthorityMetadata, error) {
	if err := ValidateFactoryDenom(meta.Denom); err != nil {
		return DenomAuthorityMetadata{}, err
	}
	admin, err := NormalizeAdminAddress(meta.Admin)
	if err != nil {
		return DenomAuthorityMetadata{}, err
	}
	meta.Admin = admin
	return meta, nil
}

func ValidateDenomAuthorityMetadata(meta DenomAuthorityMetadata) error {
	_, err := NormalizeDenomAuthorityMetadata(meta)
	return err
}

func NormalizeAdminAddress(admin string) (string, error) {
	if admin == "" {
		return "", nil
	}
	adminAddr, err := sdk.AccAddressFromBech32(admin)
	if err != nil {
		return "", ErrInvalidAddress.Wrapf("invalid admin address: %v", err)
	}
	if admin != adminAddr.String() {
		return "", ErrInvalidAddress.Wrap("admin address must be canonical")
	}
	return adminAddr.String(), nil
}
