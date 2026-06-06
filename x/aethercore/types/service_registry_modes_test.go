package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOnChainRegistryModeStoresFullDescriptorAndProofFields(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)
	authHash := ComputeServiceOwnerAuthorizationHash(service.ServiceID, service.Owner, 10)

	state, err := NewOnChainServiceRegistryState(service, authHash)
	require.NoError(t, err)
	require.Equal(t, service.ServiceID, state.Descriptor.ServiceID)
	require.Equal(t, ComputeServiceDescriptorHash(service), state.DescriptorHash)
	require.Equal(t, service.Interface.InterfaceHash, state.InterfaceDescriptorHash)
	require.Equal(t, registryPaymentModel(service), state.PaymentModel)
	require.Equal(t, service.Verification.Model, state.VerificationModel)
	require.Equal(t, authHash, state.OwnerAuthorizationHash)
	require.Equal(t, service.ExpiryHeight, state.ExpiryHeight)
	require.NoError(t, state.Validate())

	modeState, err := BuildServiceRegistryModeState(
		ServiceRegistryOnChain,
		[]ServiceDescriptor{service},
		map[string]string{service.ServiceID: authHash},
		10,
	)
	require.NoError(t, err)
	require.Len(t, modeState.OnChainStates, 1)
	require.Empty(t, modeState.HybridAnchors)
	require.NoError(t, modeState.Validate())

	descriptor, proof, found := modeState.OnChainDescriptorByID(service.ServiceID)
	require.True(t, found)
	require.Equal(t, service.ServiceID, descriptor.ServiceID)
	require.Equal(t, ServiceRegistryOnChain, proof.RegistryMode)
	require.Equal(t, modeState.StateRoot, proof.RegistryRoot)
	require.Equal(t, state.StateHash, proof.RecordHash)
	require.Equal(t, state.DescriptorHash, proof.DescriptorHash)
	require.NoError(t, proof.Validate())
}

func TestHybridRegistryModeStoresMinimalAnchor(t *testing.T) {
	service := testOffChainService("indexer-feed", ZoneIDApplication)
	service.Discovery.ProviderRoot = testHash(service.ServiceID + "/providers")

	anchor, err := NewHybridServiceRegistryAnchor(service)
	require.NoError(t, err)
	require.Equal(t, service.ServiceID, anchor.ServiceID)
	require.Equal(t, service.Owner, anchor.Owner)
	require.Equal(t, ComputeServiceDescriptorHash(service), anchor.DescriptorHash)
	require.Equal(t, service.Interface.InterfaceHash, anchor.InterfaceHash)
	require.Equal(t, service.Discovery.ProviderRoot, anchor.ProviderRoot)
	require.Equal(t, service.ExpiryHeight, anchor.ExpiryHeight)
	require.Equal(t, service.Verification.Model, anchor.VerificationModel)
	require.NoError(t, anchor.Validate())

	modeState, err := BuildServiceRegistryModeState(ServiceRegistryHybrid, []ServiceDescriptor{service}, nil, 12)
	require.NoError(t, err)
	require.Empty(t, modeState.OnChainStates)
	require.Len(t, modeState.HybridAnchors, 1)
	require.NoError(t, modeState.Validate())

	lookup, proof, found := modeState.HybridAnchorByID(service.ServiceID)
	require.True(t, found)
	require.Equal(t, anchor.ServiceID, lookup.ServiceID)
	require.Equal(t, ServiceRegistryHybrid, proof.RegistryMode)
	require.Equal(t, modeState.StateRoot, proof.RegistryRoot)
	require.Equal(t, anchor.AnchorHash, proof.RecordHash)
	require.Equal(t, anchor.DescriptorHash, proof.DescriptorHash)
	require.NoError(t, proof.Validate())
}

func TestServiceRegistryModeStateBuildsDeterministicRoots(t *testing.T) {
	first := testService("identity-resolver", ZoneIDIdentity)
	second := testService("payments-settlement", ZoneIDPayment)
	auth := map[string]string{
		first.ServiceID:  ComputeServiceOwnerAuthorizationHash(first.ServiceID, first.Owner, 20),
		second.ServiceID: ComputeServiceOwnerAuthorizationHash(second.ServiceID, second.Owner, 20),
	}

	onChain, err := BuildServiceRegistryModeState(ServiceRegistryOnChain, []ServiceDescriptor{second, first}, auth, 20)
	require.NoError(t, err)
	require.Equal(t, []string{"identity-resolver", "payments-settlement"}, []string{
		onChain.OnChainStates[0].Descriptor.ServiceID,
		onChain.OnChainStates[1].Descriptor.ServiceID,
	})
	require.Equal(t, ComputeServiceRegistryModeStateRoot(onChain), onChain.StateRoot)

	hybridFirst := testOffChainService("indexer-feed", ZoneIDApplication)
	hybridFirst.Discovery.ProviderRoot = testHash(hybridFirst.ServiceID + "/providers")
	hybridSecond := testFogMarketService("fog-compute", ZoneIDApplication)
	hybrid, err := BuildServiceRegistryModeState(ServiceRegistryHybrid, []ServiceDescriptor{hybridFirst, hybridSecond}, nil, 21)
	require.NoError(t, err)
	require.Equal(t, []string{"fog-compute", "indexer-feed"}, []string{
		hybrid.HybridAnchors[0].ServiceID,
		hybrid.HybridAnchors[1].ServiceID,
	})
	require.Equal(t, ComputeServiceRegistryModeStateRoot(hybrid), hybrid.StateRoot)
}

func TestHybridRegistryRejectsMissingProviderRoot(t *testing.T) {
	service := testOffChainService("indexer-feed", ZoneIDApplication)
	service.Discovery.ProviderRoot = ""
	service.Execution.ProviderPoolID = ""

	_, err := NewHybridServiceRegistryAnchor(service)
	require.ErrorContains(t, err, "provider root")
}

func TestOnChainRegistryRejectsMissingOwnerAuthorization(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)

	_, err := NewOnChainServiceRegistryState(service, "")
	require.ErrorContains(t, err, "owner authorization")

	_, err = BuildServiceRegistryModeState(ServiceRegistryOnChain, []ServiceDescriptor{service}, nil, 10)
	require.ErrorContains(t, err, "requires owner authorization")
}
