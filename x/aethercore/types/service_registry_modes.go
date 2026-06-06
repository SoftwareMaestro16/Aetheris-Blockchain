package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type OnChainServiceRegistryState struct {
	Descriptor              ServiceDescriptor
	DescriptorHash          string
	InterfaceDescriptorHash string
	PaymentModel            string
	VerificationModel       ServiceVerificationModel
	OwnerAuthorizationHash  string
	ExpiryHeight            uint64
	StateHash               string
}

type HybridServiceRegistryAnchor struct {
	ServiceID         string
	Owner             string
	DescriptorHash    string
	InterfaceHash     string
	ProviderRoot      string
	ExpiryHeight      uint64
	VerificationModel ServiceVerificationModel
	AnchorHash        string
}

type ServiceRegistryModeState struct {
	Mode          ServiceRegistryMode
	OnChainStates []OnChainServiceRegistryState
	HybridAnchors []HybridServiceRegistryAnchor
	StateRoot     string
	UpdatedHeight uint64
}

func NewOnChainServiceRegistryState(descriptor ServiceDescriptor, ownerAuthorizationHash string) (OnChainServiceRegistryState, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return OnChainServiceRegistryState{}, err
	}
	if err := ValidateHash("aethercore on-chain registry owner authorization", ownerAuthorizationHash); err != nil {
		return OnChainServiceRegistryState{}, err
	}
	state := OnChainServiceRegistryState{
		Descriptor:              descriptor,
		DescriptorHash:          ComputeServiceDescriptorHash(descriptor),
		InterfaceDescriptorHash: descriptor.Interface.InterfaceHash,
		PaymentModel:            registryPaymentModel(descriptor),
		VerificationModel:       descriptor.Verification.Model,
		OwnerAuthorizationHash:  strings.ToLower(strings.TrimSpace(ownerAuthorizationHash)),
		ExpiryHeight:            descriptor.ExpiryHeight,
	}
	state.StateHash = ComputeOnChainServiceRegistryStateHash(state)
	return state, state.Validate()
}

func NewHybridServiceRegistryAnchor(descriptor ServiceDescriptor) (HybridServiceRegistryAnchor, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return HybridServiceRegistryAnchor{}, err
	}
	providerRoot := registryProviderSet(descriptor)
	if providerRoot == "" {
		return HybridServiceRegistryAnchor{}, errors.New("aethercore hybrid registry anchor requires provider root")
	}
	anchor := HybridServiceRegistryAnchor{
		ServiceID:         descriptor.ServiceID,
		Owner:             descriptor.Owner,
		DescriptorHash:    ComputeServiceDescriptorHash(descriptor),
		InterfaceHash:     descriptor.Interface.InterfaceHash,
		ProviderRoot:      providerRoot,
		ExpiryHeight:      descriptor.ExpiryHeight,
		VerificationModel: descriptor.Verification.Model,
	}
	anchor.AnchorHash = ComputeHybridServiceRegistryAnchorHash(anchor)
	return anchor, anchor.Validate()
}

func BuildServiceRegistryModeState(mode ServiceRegistryMode, descriptors []ServiceDescriptor, ownerAuthorizations map[string]string, height uint64) (ServiceRegistryModeState, error) {
	if height == 0 {
		return ServiceRegistryModeState{}, errors.New("aethercore service registry mode state height must be positive")
	}
	state := ServiceRegistryModeState{
		Mode:          mode,
		UpdatedHeight: height,
	}
	switch mode {
	case ServiceRegistryOnChain:
		state.OnChainStates = make([]OnChainServiceRegistryState, 0, len(descriptors))
		for _, descriptor := range descriptors {
			descriptor = CanonicalServiceDescriptor(descriptor)
			authHash := ownerAuthorizations[descriptor.ServiceID]
			if authHash == "" {
				return ServiceRegistryModeState{}, fmt.Errorf("aethercore on-chain registry service %s requires owner authorization", descriptor.ServiceID)
			}
			onChainState, err := NewOnChainServiceRegistryState(descriptor, authHash)
			if err != nil {
				return ServiceRegistryModeState{}, err
			}
			state.OnChainStates = append(state.OnChainStates, onChainState)
		}
		sortOnChainServiceRegistryStates(state.OnChainStates)
	case ServiceRegistryHybrid:
		state.HybridAnchors = make([]HybridServiceRegistryAnchor, 0, len(descriptors))
		for _, descriptor := range descriptors {
			anchor, err := NewHybridServiceRegistryAnchor(descriptor)
			if err != nil {
				return ServiceRegistryModeState{}, err
			}
			state.HybridAnchors = append(state.HybridAnchors, anchor)
		}
		sortHybridServiceRegistryAnchors(state.HybridAnchors)
	default:
		return ServiceRegistryModeState{}, fmt.Errorf("aethercore service registry mode state does not implement %q", mode)
	}
	state.StateRoot = ComputeServiceRegistryModeStateRoot(state)
	return state, state.Validate()
}

func (state OnChainServiceRegistryState) Validate() error {
	state.Descriptor = CanonicalServiceDescriptor(state.Descriptor)
	if err := state.Descriptor.Validate(); err != nil {
		return err
	}
	if err := ValidateHash("aethercore on-chain registry descriptor hash", state.DescriptorHash); err != nil {
		return err
	}
	if expected := ComputeServiceDescriptorHash(state.Descriptor); state.DescriptorHash != expected {
		return fmt.Errorf("aethercore on-chain registry descriptor hash mismatch: expected %s", expected)
	}
	if err := ValidateHash("aethercore on-chain registry interface descriptor hash", state.InterfaceDescriptorHash); err != nil {
		return err
	}
	if state.InterfaceDescriptorHash != state.Descriptor.Interface.InterfaceHash {
		return errors.New("aethercore on-chain registry interface descriptor hash mismatch")
	}
	if err := validatePolicyID("aethercore on-chain registry payment model", state.PaymentModel); err != nil {
		return err
	}
	if state.PaymentModel != registryPaymentModel(state.Descriptor) {
		return errors.New("aethercore on-chain registry payment model mismatch")
	}
	if !IsServiceVerificationModel(state.VerificationModel) {
		return fmt.Errorf("unknown aethercore on-chain registry verification model %q", state.VerificationModel)
	}
	if state.VerificationModel != state.Descriptor.Verification.Model {
		return errors.New("aethercore on-chain registry verification model mismatch")
	}
	if err := ValidateHash("aethercore on-chain registry owner authorization", state.OwnerAuthorizationHash); err != nil {
		return err
	}
	if state.ExpiryHeight != state.Descriptor.ExpiryHeight {
		return errors.New("aethercore on-chain registry expiry mismatch")
	}
	if err := ValidateHash("aethercore on-chain registry state hash", state.StateHash); err != nil {
		return err
	}
	if expected := ComputeOnChainServiceRegistryStateHash(state); state.StateHash != expected {
		return fmt.Errorf("aethercore on-chain registry state hash mismatch: expected %s", expected)
	}
	return nil
}

func (anchor HybridServiceRegistryAnchor) Validate() error {
	anchor = CanonicalHybridServiceRegistryAnchor(anchor)
	if err := validatePolicyID("aethercore hybrid registry service id", anchor.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aethercore hybrid registry owner", anchor.Owner); err != nil {
		return err
	}
	if err := ValidateHash("aethercore hybrid registry descriptor hash", anchor.DescriptorHash); err != nil {
		return err
	}
	if err := ValidateHash("aethercore hybrid registry interface hash", anchor.InterfaceHash); err != nil {
		return err
	}
	if err := ValidateHash("aethercore hybrid registry provider root", anchor.ProviderRoot); err != nil {
		return err
	}
	if !IsServiceVerificationModel(anchor.VerificationModel) {
		return fmt.Errorf("unknown aethercore hybrid registry verification model %q", anchor.VerificationModel)
	}
	if err := ValidateHash("aethercore hybrid registry anchor hash", anchor.AnchorHash); err != nil {
		return err
	}
	if expected := ComputeHybridServiceRegistryAnchorHash(anchor); anchor.AnchorHash != expected {
		return fmt.Errorf("aethercore hybrid registry anchor hash mismatch: expected %s", expected)
	}
	return nil
}

func (state ServiceRegistryModeState) Validate() error {
	if state.UpdatedHeight == 0 {
		return errors.New("aethercore service registry mode state updated height must be positive")
	}
	switch state.Mode {
	case ServiceRegistryOnChain:
		if len(state.HybridAnchors) != 0 {
			return errors.New("aethercore on-chain registry mode state cannot contain hybrid anchors")
		}
		if err := validateOnChainServiceRegistryStates(state.OnChainStates); err != nil {
			return err
		}
	case ServiceRegistryHybrid:
		if len(state.OnChainStates) != 0 {
			return errors.New("aethercore hybrid registry mode state cannot contain full on-chain descriptors")
		}
		if err := validateHybridServiceRegistryAnchors(state.HybridAnchors); err != nil {
			return err
		}
	default:
		return fmt.Errorf("aethercore service registry mode state does not implement %q", state.Mode)
	}
	if err := ValidateHash("aethercore service registry mode state root", state.StateRoot); err != nil {
		return err
	}
	if expected := ComputeServiceRegistryModeStateRoot(state); state.StateRoot != expected {
		return fmt.Errorf("aethercore service registry mode state root mismatch: expected %s", expected)
	}
	return nil
}

func (state ServiceRegistryModeState) OnChainDescriptorByID(serviceID string) (ServiceDescriptor, ServiceRegistryProof, bool) {
	if state.Mode != ServiceRegistryOnChain {
		return ServiceDescriptor{}, ServiceRegistryProof{}, false
	}
	for _, onChainState := range state.OnChainStates {
		if onChainState.Descriptor.ServiceID == serviceID {
			proof := ServiceRegistryProof{
				ServiceID:      serviceID,
				RegistryMode:   state.Mode,
				RegistryRoot:   state.StateRoot,
				RecordHash:     onChainState.StateHash,
				DescriptorHash: onChainState.DescriptorHash,
				InterfaceHash:  onChainState.InterfaceDescriptorHash,
				ProofHeight:    state.UpdatedHeight,
			}
			proof.ProofHash = ComputeServiceRegistryProofHash(proof)
			return onChainState.Descriptor, proof, true
		}
	}
	return ServiceDescriptor{}, ServiceRegistryProof{}, false
}

func (state ServiceRegistryModeState) HybridAnchorByID(serviceID string) (HybridServiceRegistryAnchor, ServiceRegistryProof, bool) {
	if state.Mode != ServiceRegistryHybrid {
		return HybridServiceRegistryAnchor{}, ServiceRegistryProof{}, false
	}
	for _, anchor := range state.HybridAnchors {
		if anchor.ServiceID == serviceID {
			proof := ServiceRegistryProof{
				ServiceID:      serviceID,
				RegistryMode:   state.Mode,
				RegistryRoot:   state.StateRoot,
				RecordHash:     anchor.AnchorHash,
				DescriptorHash: anchor.DescriptorHash,
				InterfaceHash:  anchor.InterfaceHash,
				ProofHeight:    state.UpdatedHeight,
			}
			proof.ProofHash = ComputeServiceRegistryProofHash(proof)
			return anchor, proof, true
		}
	}
	return HybridServiceRegistryAnchor{}, ServiceRegistryProof{}, false
}

func CanonicalHybridServiceRegistryAnchor(anchor HybridServiceRegistryAnchor) HybridServiceRegistryAnchor {
	anchor.ServiceID = strings.TrimSpace(anchor.ServiceID)
	anchor.Owner = strings.TrimSpace(anchor.Owner)
	anchor.DescriptorHash = strings.ToLower(strings.TrimSpace(anchor.DescriptorHash))
	anchor.InterfaceHash = strings.ToLower(strings.TrimSpace(anchor.InterfaceHash))
	anchor.ProviderRoot = strings.ToLower(strings.TrimSpace(anchor.ProviderRoot))
	anchor.AnchorHash = strings.ToLower(strings.TrimSpace(anchor.AnchorHash))
	if anchor.AnchorHash == "" {
		anchor.AnchorHash = ComputeHybridServiceRegistryAnchorHash(anchor)
	}
	return anchor
}

func ComputeServiceOwnerAuthorizationHash(serviceID, owner string, height uint64) string {
	return hashParts(
		"aetheris-aek-service-owner-authorization-v1",
		strings.TrimSpace(serviceID),
		strings.TrimSpace(owner),
		fmt.Sprint(height),
	)
}

func ComputeOnChainServiceRegistryStateHash(state OnChainServiceRegistryState) string {
	return hashParts(
		"aetheris-aek-on-chain-service-registry-state-v1",
		state.Descriptor.ServiceID,
		state.DescriptorHash,
		state.InterfaceDescriptorHash,
		state.PaymentModel,
		string(state.VerificationModel),
		state.OwnerAuthorizationHash,
		fmt.Sprint(state.ExpiryHeight),
	)
}

func ComputeHybridServiceRegistryAnchorHash(anchor HybridServiceRegistryAnchor) string {
	anchor.AnchorHash = ""
	return hashParts(
		"aetheris-aek-hybrid-service-registry-anchor-v1",
		anchor.ServiceID,
		anchor.Owner,
		anchor.DescriptorHash,
		anchor.InterfaceHash,
		anchor.ProviderRoot,
		fmt.Sprint(anchor.ExpiryHeight),
		string(anchor.VerificationModel),
	)
}

func ComputeServiceRegistryModeStateRoot(state ServiceRegistryModeState) string {
	onChainStates := append([]OnChainServiceRegistryState(nil), state.OnChainStates...)
	hybridAnchors := append([]HybridServiceRegistryAnchor(nil), state.HybridAnchors...)
	sortOnChainServiceRegistryStates(onChainStates)
	sortHybridServiceRegistryAnchors(hybridAnchors)
	parts := []string{
		"aetheris-aek-service-registry-mode-state-root-v1",
		string(state.Mode),
		fmt.Sprint(state.UpdatedHeight),
		fmt.Sprint(len(onChainStates)),
		fmt.Sprint(len(hybridAnchors)),
	}
	for _, onChainState := range onChainStates {
		parts = append(parts, onChainState.StateHash)
	}
	for _, anchor := range hybridAnchors {
		parts = append(parts, anchor.AnchorHash)
	}
	return hashParts(parts...)
}

func validateOnChainServiceRegistryStates(states []OnChainServiceRegistryState) error {
	var previous string
	seen := make(map[string]struct{}, len(states))
	for _, state := range states {
		if err := state.Validate(); err != nil {
			return err
		}
		serviceID := state.Descriptor.ServiceID
		if _, found := seen[serviceID]; found {
			return fmt.Errorf("duplicate aethercore on-chain registry state %s", serviceID)
		}
		seen[serviceID] = struct{}{}
		if previous != "" && previous >= serviceID {
			return errors.New("aethercore on-chain registry states must be sorted canonically")
		}
		previous = serviceID
	}
	return nil
}

func validateHybridServiceRegistryAnchors(anchors []HybridServiceRegistryAnchor) error {
	var previous string
	seen := make(map[string]struct{}, len(anchors))
	for _, anchor := range anchors {
		anchor = CanonicalHybridServiceRegistryAnchor(anchor)
		if err := anchor.Validate(); err != nil {
			return err
		}
		if _, found := seen[anchor.ServiceID]; found {
			return fmt.Errorf("duplicate aethercore hybrid registry anchor %s", anchor.ServiceID)
		}
		seen[anchor.ServiceID] = struct{}{}
		if previous != "" && previous >= anchor.ServiceID {
			return errors.New("aethercore hybrid registry anchors must be sorted canonically")
		}
		previous = anchor.ServiceID
	}
	return nil
}

func sortOnChainServiceRegistryStates(states []OnChainServiceRegistryState) {
	sort.SliceStable(states, func(i, j int) bool {
		return states[i].Descriptor.ServiceID < states[j].Descriptor.ServiceID
	})
}

func sortHybridServiceRegistryAnchors(anchors []HybridServiceRegistryAnchor) {
	sort.SliceStable(anchors, func(i, j int) bool { return anchors[i].ServiceID < anchors[j].ServiceID })
}
