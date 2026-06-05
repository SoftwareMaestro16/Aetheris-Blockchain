package wasmconfig

import (
	"errors"
	"fmt"

	"github.com/sovereign-l1/l1/app/addressing"
)

func (p Policy) Validate() error {
	if !validUploadPermission(p.UploadPermission) {
		return fmt.Errorf("invalid wasm upload permission %q", p.UploadPermission)
	}
	if !validInstantiatePermission(p.InstantiatePermission) {
		return fmt.Errorf("invalid wasm instantiate permission %q", p.InstantiatePermission)
	}
	if p.AdminPolicy != AdminPolicyRequired {
		return fmt.Errorf("invalid wasm admin policy %q", p.AdminPolicy)
	}
	if !validPinnedCodePolicy(p.PinnedCodePolicy) {
		return fmt.Errorf("invalid wasm pinned code policy %q", p.PinnedCodePolicy)
	}
	if err := validateLimits(p); err != nil {
		return err
	}
	if !p.Enabled {
		return nil
	}
	if err := addressing.ValidateAuthorityAddress("wasm governance authority", p.GovernanceAuthority); err != nil {
		return err
	}
	if p.UploadPermission == UploadPermissionAllowlist && len(p.UploadAllowlist) == 0 {
		return errors.New("wasm upload allowlist must not be empty when allowlist upload is enabled")
	}
	for _, actor := range p.UploadAllowlist {
		if err := addressing.ValidateUserAddress("wasm upload allowlist actor", actor); err != nil {
			return err
		}
	}
	return nil
}

func CanUpdatePolicy(actor string, current Policy, next Policy) error {
	if err := current.Validate(); err != nil {
		return err
	}
	authority := current.GovernanceAuthority
	if authority == "" {
		authority = next.GovernanceAuthority
	}
	if err := addressing.ValidateAuthorityAddress("wasm governance authority", authority); err != nil {
		return err
	}
	if same, err := sameAddress(actor, authority); err != nil {
		return err
	} else if !same {
		return errors.New("wasm policy update requires governance authority")
	}
	if next.GovernanceAuthority == "" {
		next.GovernanceAuthority = authority
	}
	return next.Validate()
}

func requireEnabled(p Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if !p.Enabled {
		return errors.New("CosmWasm is disabled by feature gate")
	}
	return nil
}
