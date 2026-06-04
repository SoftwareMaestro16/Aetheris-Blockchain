package types

import (
	"fmt"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{Denoms: []DenomAuthorityMetadata{}}
}

func (gs GenesisState) Validate() error {
	seen := map[string]struct{}{}
	for _, denom := range gs.Denoms {
		normalized, err := NormalizeDenomAuthorityMetadata(denom)
		if err != nil {
			return fmt.Errorf("invalid denom authority metadata for %s: %w", denom.Denom, err)
		}
		if _, ok := seen[normalized.Denom]; ok {
			return fmt.Errorf("duplicate denom %s", normalized.Denom)
		}
		seen[normalized.Denom] = struct{}{}
	}
	return nil
}
