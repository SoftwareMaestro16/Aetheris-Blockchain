package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAetherMeshRoutingGraphSetConnectsAllGraphRootsDeterministically(t *testing.T) {
	zoneA, err := NewAetherMeshZoneEdge(AetherMeshZoneEdge{
		SourceZone:          "APPLICATION_ZONE",
		DestinationZone:     "CONTRACT_ZONE",
		Enabled:             true,
		CommittedGasCost:    100,
		CongestionWeightBps: 2500,
		ForwardingFeeWeight: 7,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(132), zoneA.EdgeWeight)

	zoneB, err := NewAetherMeshZoneEdge(AetherMeshZoneEdge{
		SourceZone:       "CONTRACT_ZONE",
		DestinationZone:  "FINANCIAL_ZONE",
		Enabled:          true,
		CommittedGasCost: 80,
	})
	require.NoError(t, err)

	service, err := NewAetherMeshServiceEdge(AetherMeshServiceEdge{
		SourceService:          "svc.payments",
		DependencyService:      "svc.storage",
		InterfaceHash:          HashParts("iface", "payments-storage"),
		InterfaceCompatible:    true,
		AvailabilityCommitment: HashParts("availability", "svc.storage"),
		AvailabilityWeightBps:  9500,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(501), service.EdgeWeight)

	message, err := NewAetherMeshMessageEdge(AetherMeshMessageEdge{
		SourceQueue:       HashParts("queue", "contract-outbox"),
		DestinationQueue:  HashParts("queue", "financial-inbox"),
		DeliveryLane:      "cross-zone/settlement",
		QueueBacklog:      13,
		ForwardingFee:     5,
		PriorityWeightBps: 9000,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1018), message.EdgeWeight)

	paymentRoot := HashParts("payment-route-graph")
	storageRoot := HashParts("storage-retrieval-graph")
	left, err := BuildAetherMeshRoutingGraphSet(7, []AetherMeshZoneEdge{zoneA, zoneB}, []AetherMeshServiceEdge{service}, []AetherMeshMessageEdge{message}, paymentRoot, storageRoot)
	require.NoError(t, err)
	require.NoError(t, left.Validate())
	require.Equal(t, paymentRoot, left.PaymentRoot)
	require.Equal(t, storageRoot, left.StorageRoot)
	require.Len(t, left.Graphs, 5)

	right, err := BuildAetherMeshRoutingGraphSet(7, []AetherMeshZoneEdge{zoneB, zoneA}, []AetherMeshServiceEdge{service}, []AetherMeshMessageEdge{message}, paymentRoot, storageRoot)
	require.NoError(t, err)
	require.Equal(t, left.GraphSetRoot, right.GraphSetRoot)
	require.Equal(t, left.Graphs, right.Graphs)
}

func TestAetherMeshRoutingGraphsRejectInvalidWeightsAndUncommittedDependencies(t *testing.T) {
	_, err := NewAetherMeshZoneEdge(AetherMeshZoneEdge{
		SourceZone:       "APPLICATION_ZONE",
		DestinationZone:  "CONTRACT_ZONE",
		Enabled:          false,
		CommittedGasCost: 100,
	})
	require.ErrorContains(t, err, "enabled")

	_, err = NewAetherMeshServiceEdge(AetherMeshServiceEdge{
		SourceService:          "svc.a",
		DependencyService:      "svc.b",
		InterfaceHash:          HashParts("iface"),
		InterfaceCompatible:    false,
		AvailabilityCommitment: HashParts("availability"),
		AvailabilityWeightBps:  9000,
	})
	require.ErrorContains(t, err, "interface compatible")

	_, err = NewAetherMeshMessageEdge(AetherMeshMessageEdge{
		SourceQueue:       HashParts("queue", "a"),
		DestinationQueue:  HashParts("queue", "b"),
		DeliveryLane:      "lane.primary",
		QueueBacklog:      1,
		ForwardingFee:     0,
		PriorityWeightBps: 1000,
	})
	require.ErrorContains(t, err, "forwarding fee")
}

func TestAetherMeshGraphSetRejectsRootTampering(t *testing.T) {
	zone, err := NewAetherMeshZoneEdge(AetherMeshZoneEdge{
		SourceZone:       "APPLICATION_ZONE",
		DestinationZone:  "CONTRACT_ZONE",
		Enabled:          true,
		CommittedGasCost: 10,
	})
	require.NoError(t, err)
	graphSet, err := BuildAetherMeshRoutingGraphSet(9, []AetherMeshZoneEdge{zone}, nil, nil, "", "")
	require.NoError(t, err)

	tampered := graphSet
	tampered.Graphs = append([]AetherMeshCommittedGraph(nil), graphSet.Graphs...)
	tampered.Graphs[0].RootHash = HashParts("wrong-root")
	tampered.GraphSetRoot = ComputeAetherMeshRoutingGraphSetRoot(tampered)
	require.ErrorContains(t, tampered.Validate(), "committed graph root mismatch")

	tamperedSetRoot := graphSet
	tamperedSetRoot.GraphSetRoot = HashParts("wrong-set-root")
	require.ErrorContains(t, tamperedSetRoot.Validate(), "graph set root mismatch")
}
