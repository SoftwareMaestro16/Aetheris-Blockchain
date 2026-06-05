package wasmconfig

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	ModuleName = "wasm"
	StoreKey   = ModuleName

	RecommendedWasmdVersion  = "v0.70.2"
	RecommendedWasmVMVersion = "v3.0.6"
	RecommendedSDKMinor      = "v0.54"

	DefaultMaxContractSizeBytes         uint64 = 800 * 1024
	DefaultMaxProposalContractSizeBytes uint64 = 3 * 1024 * 1024
	DefaultSmartQueryGasLimit           uint64 = 3_000_000
	DefaultSimulationGasLimit           uint64 = 20_000_000
	DefaultGasMultiplier                uint64 = 140_000
	DefaultMemoryCacheSizeMiB           uint32 = 100

	maxSmartQueryGasLimit uint64 = 10_000_000
	maxSimulationGasLimit uint64 = 100_000_000
	maxMemoryCacheSizeMiB uint32 = 256
)

type UploadPermission string

const (
	UploadPermissionGovernanceOnly UploadPermission = "governance-only"
	UploadPermissionAllowlist      UploadPermission = "allowlist"
)

type InstantiatePermission string

const (
	InstantiatePermissionCodeOwnerOnly InstantiatePermission = "code-owner-only"
	InstantiatePermissionEverybody     InstantiatePermission = "everybody"
)

type AdminPolicy string

const (
	AdminPolicyRequired AdminPolicy = "required"
)

type Policy struct {
	Enabled                      bool
	UploadPermission             UploadPermission
	InstantiatePermission        InstantiatePermission
	AdminPolicy                  AdminPolicy
	GovernanceAuthority          string
	UploadAllowlist              []string
	MaxContractSizeBytes         uint64
	MaxProposalContractSizeBytes uint64
	SmartQueryGasLimit           uint64
	SimulationGasLimit           uint64
	GasMultiplier                uint64
	MemoryCacheSizeMiB           uint32
}

func DefaultPolicy() Policy {
	return Policy{
		Enabled:                      false,
		UploadPermission:             UploadPermissionGovernanceOnly,
		InstantiatePermission:        InstantiatePermissionCodeOwnerOnly,
		AdminPolicy:                  AdminPolicyRequired,
		MaxContractSizeBytes:         DefaultMaxContractSizeBytes,
		MaxProposalContractSizeBytes: DefaultMaxProposalContractSizeBytes,
		SmartQueryGasLimit:           DefaultSmartQueryGasLimit,
		SimulationGasLimit:           DefaultSimulationGasLimit,
		GasMultiplier:                DefaultGasMultiplier,
		MemoryCacheSizeMiB:           DefaultMemoryCacheSizeMiB,
	}
}

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

func CanUpload(actor string, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm upload actor", actor); err != nil {
		return err
	}
	if same, err := sameAddress(actor, p.GovernanceAuthority); err != nil {
		return err
	} else if same {
		return nil
	}
	if p.UploadPermission != UploadPermissionAllowlist {
		return errors.New("wasm upload requires governance authority")
	}
	for _, allowed := range p.UploadAllowlist {
		same, err := sameAddress(actor, allowed)
		if err != nil {
			return err
		}
		if same {
			return nil
		}
	}
	return errors.New("wasm upload actor is not allowlisted")
}

func CanInstantiate(actor, codeOwner string, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm instantiate actor", actor); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm code owner", codeOwner); err != nil {
		return err
	}
	if p.InstantiatePermission == InstantiatePermissionEverybody {
		return nil
	}
	same, err := sameAddress(actor, codeOwner)
	if err != nil {
		return err
	}
	if !same {
		return errors.New("wasm instantiate requires code owner")
	}
	return nil
}

func CanExecute(actor, contract string, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm execute actor", actor); err != nil {
		return err
	}
	return addressing.ValidateUserAddress("wasm contract address", contract)
}

func CanMigrate(actor, admin string, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm migrate actor", actor); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm contract admin", admin); err != nil {
		return err
	}
	same, err := sameAddress(actor, admin)
	if err != nil {
		return err
	}
	if !same {
		return errors.New("wasm migrate requires contract admin")
	}
	return nil
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

func sameAddress(left, right string) (bool, error) {
	leftAddr, err := addressing.ParseUserAddress("left address", left)
	if err != nil {
		return false, err
	}
	rightAddr, err := addressing.ParseUserAddress("right address", right)
	if err != nil {
		return false, err
	}
	return bytes.Equal(leftAddr.Bytes(), rightAddr.Bytes()), nil
}

func validUploadPermission(permission UploadPermission) bool {
	switch permission {
	case UploadPermissionGovernanceOnly, UploadPermissionAllowlist:
		return true
	default:
		return false
	}
}

func validInstantiatePermission(permission InstantiatePermission) bool {
	switch permission {
	case InstantiatePermissionCodeOwnerOnly, InstantiatePermissionEverybody:
		return true
	default:
		return false
	}
}

func validateLimits(p Policy) error {
	if p.MaxContractSizeBytes == 0 || p.MaxContractSizeBytes > DefaultMaxContractSizeBytes {
		return fmt.Errorf("wasm max contract size must be between 1 and %d bytes", DefaultMaxContractSizeBytes)
	}
	if p.MaxProposalContractSizeBytes < p.MaxContractSizeBytes ||
		p.MaxProposalContractSizeBytes > DefaultMaxProposalContractSizeBytes {
		return fmt.Errorf("wasm proposal contract size must be between max contract size and %d bytes", DefaultMaxProposalContractSizeBytes)
	}
	if p.SmartQueryGasLimit == 0 || p.SmartQueryGasLimit > maxSmartQueryGasLimit {
		return fmt.Errorf("wasm smart query gas limit must be between 1 and %d", maxSmartQueryGasLimit)
	}
	if p.SimulationGasLimit == 0 || p.SimulationGasLimit > maxSimulationGasLimit {
		return fmt.Errorf("wasm simulation gas limit must be between 1 and %d", maxSimulationGasLimit)
	}
	if p.GasMultiplier != DefaultGasMultiplier {
		return fmt.Errorf("wasm gas multiplier must remain %d until benchmarked otherwise", DefaultGasMultiplier)
	}
	if p.MemoryCacheSizeMiB > maxMemoryCacheSizeMiB {
		return fmt.Errorf("wasm memory cache size must be at most %d MiB", maxMemoryCacheSizeMiB)
	}
	return nil
}
