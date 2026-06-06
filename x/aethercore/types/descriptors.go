package types

import (
	"errors"
	"fmt"
	"sort"
)

type ZoneType string

const (
	ZoneTypeCore        ZoneType = "CORE"
	ZoneTypeAetherCore  ZoneType = "AETHER_CORE"
	ZoneTypeFinancial   ZoneType = "FINANCIAL"
	ZoneTypeIdentity    ZoneType = "IDENTITY"
	ZoneTypeStorage     ZoneType = "STORAGE"
	ZoneTypePayment     ZoneType = "PAYMENT"
	ZoneTypeContract    ZoneType = "CONTRACT"
	ZoneTypeApplication ZoneType = "APPLICATION"
	ZoneTypeService     ZoneType = "SERVICE"
)

type ZoneDescriptor struct {
	ZoneID                ZoneID
	ZoneType              ZoneType
	ModuleName            string
	Enabled               bool
	StateMachineVersion   uint64
	MempoolPolicyID       string
	FeePolicyID           string
	ShardLayoutEpoch      uint64
	MaxShards             uint32
	MessageCapabilities   []string
	ProofCapabilities     []string
	UpgradeHeightOptional uint64
}

type ServiceDescriptor struct {
	ServiceID        string
	ZoneID           ZoneID
	InterfaceID      string
	EndpointKey      string
	Version          uint64
	AvailabilityHash string
	Enabled          bool
}

func (d ZoneDescriptor) Validate(params AetherCoreParams) error {
	if err := ValidateZoneID(d.ZoneID); err != nil {
		return err
	}
	if !IsZoneType(d.ZoneType) {
		return fmt.Errorf("unknown aethercore zone type %q", d.ZoneType)
	}
	if err := validateModuleName(d.ModuleName); err != nil {
		return err
	}
	if d.StateMachineVersion == 0 {
		return errors.New("aethercore zone state machine version must be positive")
	}
	if err := validatePolicyID("aethercore zone mempool policy", d.MempoolPolicyID); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore zone fee policy", d.FeePolicyID); err != nil {
		return err
	}
	if d.FeePolicyID != NativeFeePolicyID {
		return fmt.Errorf("aethercore zone fee policy must use %s", NativeFeePolicyID)
	}
	if d.MaxShards == 0 || d.MaxShards > params.MaxShardsPerZone {
		return fmt.Errorf("aethercore zone max shards must be between 1 and %d", params.MaxShardsPerZone)
	}
	if err := validateCapabilitiesForField("aethercore zone message capabilities", d.MessageCapabilities); err != nil {
		return err
	}
	return validateCapabilitiesForField("aethercore zone proof capabilities", d.ProofCapabilities)
}

func (d ServiceDescriptor) Validate() error {
	if err := validatePolicyID("aethercore service id", d.ServiceID); err != nil {
		return err
	}
	if err := ValidateZoneID(d.ZoneID); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore service interface id", d.InterfaceID); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore service endpoint key", d.EndpointKey); err != nil {
		return err
	}
	if d.Version == 0 {
		return errors.New("aethercore service version must be positive")
	}
	return ValidateHash("aethercore service availability hash", d.AvailabilityHash)
}

func IsZoneType(zoneType ZoneType) bool {
	switch zoneType {
	case ZoneTypeCore, ZoneTypeAetherCore, ZoneTypeFinancial, ZoneTypeIdentity, ZoneTypeStorage, ZoneTypePayment, ZoneTypeContract, ZoneTypeApplication, ZoneTypeService:
		return true
	default:
		return false
	}
}

func CanonicalZoneDescriptor(d ZoneDescriptor) ZoneDescriptor {
	d.MessageCapabilities = append([]string(nil), d.MessageCapabilities...)
	d.ProofCapabilities = append([]string(nil), d.ProofCapabilities...)
	sort.Strings(d.MessageCapabilities)
	sort.Strings(d.ProofCapabilities)
	return d
}

func ComputeServiceDescriptorHash(d ServiceDescriptor) string {
	return hashParts(
		"aetheris-aek-service-descriptor-v1",
		d.ServiceID,
		string(d.ZoneID),
		d.InterfaceID,
		d.EndpointKey,
		fmt.Sprint(d.Version),
		d.AvailabilityHash,
		fmt.Sprint(d.Enabled),
	)
}

func ComputeServiceRoot(services []ServiceDescriptor) (string, error) {
	ordered := append([]ServiceDescriptor(nil), services...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].ServiceID < ordered[j].ServiceID
	})
	parts := []string{"aetheris-aek-services-root-v1", fmt.Sprint(len(ordered))}
	var previous string
	for i, service := range ordered {
		if err := service.Validate(); err != nil {
			return "", err
		}
		if i > 0 && previous >= service.ServiceID {
			return "", errors.New("aethercore services must be sorted canonically by service id")
		}
		parts = append(parts, ComputeServiceDescriptorHash(service))
		previous = service.ServiceID
	}
	return hashParts(parts...), nil
}
