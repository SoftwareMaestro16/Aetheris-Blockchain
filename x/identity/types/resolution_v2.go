package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MaxUnifiedContractTargets      = 16
	MaxUnifiedServiceEndpoints     = 16
	MaxUnifiedInterfaceDescriptors = 16
	MaxUnifiedExecutionHints       = 16
	MaxUnifiedRecordKeyBytes       = 48
	MaxUnifiedRecordValueBytes     = 128
	MaxUnifiedEndpointBytes        = 128
	MaxUnifiedOwnerSignatureBytes  = 128
)

type ContractTargetV2 struct {
	Key     string
	Address sdk.AccAddress
}

type ServiceEndpointV2 struct {
	Key      string
	Endpoint string
}

type InterfaceDescriptorV2 struct {
	InterfaceID string
	Descriptor  string
}

type RoutingMetadataV2 struct {
	ZoneID     string
	ShardID    string
	VM         string
	Entrypoint string
}

type ExecutionHintV2 struct {
	Key   string
	Value string
}

type UnifiedResolutionRecordV2 struct {
	NameHash               string
	PrimaryAddress         sdk.AccAddress
	ContractTargets        []ContractTargetV2
	ServiceEndpoints       []ServiceEndpointV2
	InterfaceDescriptors   []InterfaceDescriptorV2
	RoutingMetadata        RoutingMetadataV2
	ExecutionHints         []ExecutionHintV2
	RecordVersion          uint64
	RecordTTL              uint64
	UpdatedAtHeight        uint64
	OwnerSignatureOptional []byte
}

type ReverseResolutionRecordV2 struct {
	Address         sdk.AccAddress
	NameHash        string
	Name            string
	Verified        bool
	UpdatedAtHeight uint64
	ExpiryHeight    uint64
}

func BuildUnifiedResolutionRecordV2(state IdentityState, name string, height uint64, ttl uint64) (UnifiedResolutionRecordV2, error) {
	view, err := BuildUnifiedResolverView(state, name, height)
	if err != nil {
		return UnifiedResolutionRecordV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(view.QueryDomain)
	if err != nil {
		return UnifiedResolutionRecordV2{}, err
	}
	record := UnifiedResolutionRecordV2{
		NameHash:        nameHash,
		PrimaryAddress:  cloneSpecAddress(view.Primary),
		RoutingMetadata: routeV2FromExecutionRoute(view.Route),
		RecordVersion:   1,
		RecordTTL:       ttl,
		UpdatedAtHeight: height,
	}
	if len(view.Contract) > 0 {
		record.ContractTargets = append(record.ContractTargets, ContractTargetV2{Key: ResolverKeyContract, Address: cloneSpecAddress(view.Contract)})
	}
	for _, addressRecord := range view.Records {
		record.ContractTargets = append(record.ContractTargets, ContractTargetV2{Key: addressRecord.Key, Address: cloneSpecAddress(addressRecord.Address)})
	}
	for _, entry := range view.Metadata {
		switch {
		case strings.HasPrefix(entry.Key, ResolverMetadataServicePrefix):
			record.ServiceEndpoints = append(record.ServiceEndpoints, ServiceEndpointV2{
				Key:      strings.TrimPrefix(entry.Key, ResolverMetadataServicePrefix),
				Endpoint: entry.Value,
			})
		case strings.HasPrefix(entry.Key, ResolverMetadataInterfacePrefix):
			record.InterfaceDescriptors = append(record.InterfaceDescriptors, InterfaceDescriptorV2{
				InterfaceID: strings.TrimPrefix(entry.Key, ResolverMetadataInterfacePrefix),
				Descriptor:  entry.Value,
			})
		case isResolverRouteMetadataKey(entry.Key):
			continue
		default:
			record.ExecutionHints = append(record.ExecutionHints, ExecutionHintV2{Key: entry.Key, Value: entry.Value})
		}
	}
	sortUnifiedResolutionRecordV2(&record)
	return record, ValidateUnifiedResolutionRecordV2(record)
}

func ValidateUnifiedResolutionRecordV2(record UnifiedResolutionRecordV2) error {
	if err := validateHexHash("identity v2 unified resolution name hash", record.NameHash); err != nil {
		return err
	}
	if len(record.PrimaryAddress) > 0 {
		if err := validateSpecAddress("identity v2 unified primary address", record.PrimaryAddress); err != nil {
			return err
		}
	}
	if len(record.ContractTargets) > MaxUnifiedContractTargets {
		return fmt.Errorf("identity v2 contract targets must not exceed %d", MaxUnifiedContractTargets)
	}
	if len(record.ServiceEndpoints) > MaxUnifiedServiceEndpoints {
		return fmt.Errorf("identity v2 service endpoints must not exceed %d", MaxUnifiedServiceEndpoints)
	}
	if len(record.InterfaceDescriptors) > MaxUnifiedInterfaceDescriptors {
		return fmt.Errorf("identity v2 interface descriptors must not exceed %d", MaxUnifiedInterfaceDescriptors)
	}
	if len(record.ExecutionHints) > MaxUnifiedExecutionHints {
		return fmt.Errorf("identity v2 execution hints must not exceed %d", MaxUnifiedExecutionHints)
	}
	if err := validateContractTargetsV2(record.ContractTargets); err != nil {
		return err
	}
	if err := validateServiceEndpointsV2(record.ServiceEndpoints); err != nil {
		return err
	}
	if err := validateInterfaceDescriptorsV2(record.InterfaceDescriptors); err != nil {
		return err
	}
	if err := validateRoutingMetadataV2(record.RoutingMetadata); err != nil {
		return err
	}
	if err := validateExecutionHintsV2(record.ExecutionHints); err != nil {
		return err
	}
	if record.RecordVersion == 0 {
		return errors.New("identity v2 unified record version is required")
	}
	if record.RecordTTL == 0 {
		return errors.New("identity v2 unified record ttl is required")
	}
	if record.UpdatedAtHeight == 0 {
		return errors.New("identity v2 unified updated_at_height is required")
	}
	if len(record.OwnerSignatureOptional) > MaxUnifiedOwnerSignatureBytes {
		return fmt.Errorf("identity v2 owner signature must not exceed %d bytes", MaxUnifiedOwnerSignatureBytes)
	}
	return nil
}

func NewReverseResolutionRecordV2(address sdk.AccAddress, name string, verified bool, updatedAtHeight uint64, expiryHeight uint64) (ReverseResolutionRecordV2, error) {
	if err := validateSpecAddress("identity v2 reverse address", address); err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(normalized)
	if err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	record := ReverseResolutionRecordV2{
		Address:         cloneSpecAddress(address),
		NameHash:        nameHash,
		Name:            normalized,
		Verified:        verified,
		UpdatedAtHeight: updatedAtHeight,
		ExpiryHeight:    expiryHeight,
	}
	return record, ValidateReverseResolutionRecordV2Format(record)
}

func ValidateReverseResolutionRecordV2Format(record ReverseResolutionRecordV2) error {
	if err := validateSpecAddress("identity v2 reverse address", record.Address); err != nil {
		return err
	}
	normalized, err := NormalizeAETDomain(record.Name)
	if err != nil {
		return err
	}
	if record.Name != normalized {
		return errors.New("identity v2 reverse name must be normalized")
	}
	expectedNameHash, err := DomainRecordV2NameHash(record.Name)
	if err != nil {
		return err
	}
	if record.NameHash != expectedNameHash {
		return errors.New("identity v2 reverse name_hash mismatch")
	}
	if record.UpdatedAtHeight == 0 {
		return errors.New("identity v2 reverse updated_at_height is required")
	}
	if record.ExpiryHeight <= record.UpdatedAtHeight {
		return errors.New("identity v2 reverse expiry_height must be after updated_at_height")
	}
	return nil
}

func ValidateReverseResolutionRecordV2(state IdentityState, record ReverseResolutionRecordV2, height uint64, authorizedAliasKeys []string) error {
	if err := ValidateReverseResolutionRecordV2Format(record); err != nil {
		return err
	}
	if record.ExpiryHeight <= height {
		return errors.New("identity v2 reverse record is expired")
	}
	if !record.Verified {
		return nil
	}
	resolution, err := ResolveIdentityRecordRecursive(state, record.Name, height)
	if err != nil {
		return err
	}
	if forwardResolutionContainsAddress(resolution.Record, record.Address, authorizedAliasKeys) {
		return nil
	}
	return errors.New("identity v2 verified reverse record requires forward primary or authorized alias")
}

func CanonicalReverseResolutionName(record ReverseResolutionRecordV2) (string, error) {
	if err := ValidateReverseResolutionRecordV2Format(record); err != nil {
		return "", err
	}
	if !record.Verified {
		return "", errors.New("identity v2 unverified reverse record is not canonical")
	}
	return record.Name, nil
}

func forwardResolutionContainsAddress(record ResolverRecord, address sdk.AccAddress, authorizedAliasKeys []string) bool {
	if addressesEqual(record.Primary, address) {
		return true
	}
	allowed := stringSet(authorizedAliasKeys)
	for _, key := range sortedResolverKeys(record.Records) {
		if _, found := allowed[key]; !found {
			continue
		}
		if addressesEqual(record.Records[key], address) {
			return true
		}
	}
	return false
}

func sortUnifiedResolutionRecordV2(record *UnifiedResolutionRecordV2) {
	sort.SliceStable(record.ContractTargets, func(i, j int) bool { return record.ContractTargets[i].Key < record.ContractTargets[j].Key })
	sort.SliceStable(record.ServiceEndpoints, func(i, j int) bool { return record.ServiceEndpoints[i].Key < record.ServiceEndpoints[j].Key })
	sort.SliceStable(record.InterfaceDescriptors, func(i, j int) bool {
		return record.InterfaceDescriptors[i].InterfaceID < record.InterfaceDescriptors[j].InterfaceID
	})
	sort.SliceStable(record.ExecutionHints, func(i, j int) bool { return record.ExecutionHints[i].Key < record.ExecutionHints[j].Key })
}

func routeV2FromExecutionRoute(route IdentityExecutionRoute) RoutingMetadataV2 {
	return RoutingMetadataV2{
		ZoneID:     route.ZoneID,
		ShardID:    route.ShardID,
		VM:         route.VM,
		Entrypoint: route.Entrypoint,
	}
}

func isResolverRouteMetadataKey(key string) bool {
	switch key {
	case ResolverMetadataRouteZone, ResolverMetadataRouteShard, ResolverMetadataRouteVM, ResolverMetadataRouteEntrypoint:
		return true
	default:
		return false
	}
}

func validateContractTargetsV2(targets []ContractTargetV2) error {
	seen := map[string]struct{}{}
	for i, target := range targets {
		if err := validateUnifiedRecordKey("identity v2 contract target key", target.Key); err != nil {
			return err
		}
		if err := validateSpecAddress("identity v2 contract target", target.Address); err != nil {
			return err
		}
		if _, found := seen[target.Key]; found {
			return fmt.Errorf("duplicate identity v2 contract target %q", target.Key)
		}
		seen[target.Key] = struct{}{}
		if i > 0 && targets[i-1].Key >= target.Key {
			return errors.New("identity v2 contract targets must be sorted canonically")
		}
	}
	return nil
}

func validateServiceEndpointsV2(endpoints []ServiceEndpointV2) error {
	seen := map[string]struct{}{}
	for i, endpoint := range endpoints {
		if err := validateUnifiedRecordKey("identity v2 service endpoint key", endpoint.Key); err != nil {
			return err
		}
		if err := validateUnifiedRecordValue("identity v2 service endpoint", endpoint.Endpoint, MaxUnifiedEndpointBytes); err != nil {
			return err
		}
		if _, found := seen[endpoint.Key]; found {
			return fmt.Errorf("duplicate identity v2 service endpoint %q", endpoint.Key)
		}
		seen[endpoint.Key] = struct{}{}
		if i > 0 && endpoints[i-1].Key >= endpoint.Key {
			return errors.New("identity v2 service endpoints must be sorted canonically")
		}
	}
	return nil
}

func validateInterfaceDescriptorsV2(descriptors []InterfaceDescriptorV2) error {
	seen := map[string]struct{}{}
	for i, descriptor := range descriptors {
		if err := validateUnifiedRecordKey("identity v2 interface id", descriptor.InterfaceID); err != nil {
			return err
		}
		if err := validateUnifiedRecordValue("identity v2 interface descriptor", descriptor.Descriptor, MaxUnifiedRecordValueBytes); err != nil {
			return err
		}
		if _, found := seen[descriptor.InterfaceID]; found {
			return fmt.Errorf("duplicate identity v2 interface descriptor %q", descriptor.InterfaceID)
		}
		seen[descriptor.InterfaceID] = struct{}{}
		if i > 0 && descriptors[i-1].InterfaceID >= descriptor.InterfaceID {
			return errors.New("identity v2 interface descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateRoutingMetadataV2(route RoutingMetadataV2) error {
	for field, value := range map[string]string{
		"zone_id":    route.ZoneID,
		"shard_id":   route.ShardID,
		"vm":         route.VM,
		"entrypoint": route.Entrypoint,
	} {
		if value == "" {
			continue
		}
		if err := validateUnifiedRecordValue("identity v2 routing "+field, value, MaxUnifiedRecordValueBytes); err != nil {
			return err
		}
	}
	return nil
}

func validateExecutionHintsV2(hints []ExecutionHintV2) error {
	seen := map[string]struct{}{}
	for i, hint := range hints {
		if err := validateUnifiedRecordKey("identity v2 execution hint key", hint.Key); err != nil {
			return err
		}
		if err := validateUnifiedRecordValue("identity v2 execution hint", hint.Value, MaxUnifiedRecordValueBytes); err != nil {
			return err
		}
		if _, found := seen[hint.Key]; found {
			return fmt.Errorf("duplicate identity v2 execution hint %q", hint.Key)
		}
		seen[hint.Key] = struct{}{}
		if i > 0 && hints[i-1].Key >= hint.Key {
			return errors.New("identity v2 execution hints must be sorted canonically")
		}
	}
	return nil
}

func validateUnifiedRecordKey(field string, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > MaxUnifiedRecordKeyBytes {
		return fmt.Errorf("%s must not exceed %d bytes", field, MaxUnifiedRecordKeyBytes)
	}
	return ValidateResolverMetadataKey(value)
}

func validateUnifiedRecordValue(field string, value string, maxBytes int) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have surrounding whitespace", field)
	}
	if len(value) > maxBytes {
		return fmt.Errorf("%s must not exceed %d bytes", field, maxBytes)
	}
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c < 0x21 || c > 0x7e {
			return fmt.Errorf("%s contains unsupported character %q", field, c)
		}
	}
	return nil
}
