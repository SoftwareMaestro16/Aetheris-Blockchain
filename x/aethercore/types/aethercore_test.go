package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterZonesAndAggregateGlobalRoot(t *testing.T) {
	state := EmptyState()
	var err error
	for _, zone := range []ZoneDescriptor{
		testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"),
		testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"),
		testDescriptor(ZoneIDContract, ZoneTypeContract, "contract"),
	} {
		state, err = RegisterZoneDescriptor(state, zone)
		require.NoError(t, err)
	}
	state, err = RegisterServiceDescriptor(state, testService("identity-resolver", ZoneIDIdentity))
	require.NoError(t, err)

	for _, commitment := range []ZoneCommitment{
		testCommitment(t, 10, ZoneIDIdentity),
		testCommitment(t, 10, ZoneIDFinancial),
		testCommitment(t, 10, ZoneIDContract),
	} {
		state, err = AppendZoneCommitment(state, commitment)
		require.NoError(t, err)
	}

	root, err := BuildGlobalStateRoot(10, state, testContributions(10))
	require.NoError(t, err)
	require.NoError(t, root.ValidateHash())

	state, err = AppendGlobalRoot(state, root)
	require.NoError(t, err)
	manifest, err := NewExportManifest(root, testHash("app-hash"), state)
	require.NoError(t, err)
	state, err = AddExportManifest(state, manifest)
	require.NoError(t, err)

	require.Len(t, state.GlobalRoots, 1)
	require.Len(t, state.ExportManifests, 1)
	require.Equal(t, root.GlobalRoot, state.ExportManifests[0].GlobalRoot)
	require.NoError(t, state.Validate())
}

func TestRootReplayIdenticalAcrossNodes(t *testing.T) {
	nodeA := populatedState(t, []ZoneID{ZoneIDFinancial, ZoneIDContract})
	nodeB := populatedState(t, []ZoneID{ZoneIDContract, ZoneIDFinancial})

	rootA, err := BuildGlobalStateRoot(7, nodeA, testContributions(7))
	require.NoError(t, err)
	rootB, err := BuildGlobalStateRoot(7, nodeB, testContributions(7))
	require.NoError(t, err)

	nodeA, err = AppendGlobalRoot(nodeA, rootA)
	require.NoError(t, err)
	nodeB, err = AppendGlobalRoot(nodeB, rootB)
	require.NoError(t, err)

	require.Equal(t, rootA, rootB)
	require.Equal(t, nodeA.Export(), nodeB.Export())
}

func TestProposalScheduleGroupsByZoneAndShardDeterministically(t *testing.T) {
	items := []ProposalItem{
		testProposalItem(ZoneIDContract, "2", "c", 4, 15, 2),
		testProposalItem(ZoneIDFinancial, "0", "b", 2, 14, 1),
		testProposalItem(ZoneIDContract, "1", "a", 4, 13, 0),
		testProposalItem(ZoneIDFinancial, "0", "a", 1, 16, 3),
	}

	schedule, err := BuildProposalSchedule(9, items, TestnetParams())
	require.NoError(t, err)
	require.NoError(t, schedule.Validate())
	require.Equal(t, []ProposalGroup{
		{
			ZoneID:  ZoneIDContract,
			ShardID: "1",
			Items:   []ProposalItem{testProposalItem(ZoneIDContract, "1", "a", 4, 13, 0)},
		},
		{
			ZoneID:  ZoneIDContract,
			ShardID: "2",
			Items:   []ProposalItem{testProposalItem(ZoneIDContract, "2", "c", 4, 15, 2)},
		},
		{
			ZoneID:  ZoneIDFinancial,
			ShardID: "0",
			Items: []ProposalItem{
				testProposalItem(ZoneIDFinancial, "0", "a", 1, 16, 3),
				testProposalItem(ZoneIDFinancial, "0", "b", 2, 14, 1),
			},
		},
	}, schedule.Groups)
}

func TestShardLayoutsAndRoutingTablesCommitDeterministically(t *testing.T) {
	nodeA := EmptyState()
	nodeB := EmptyState()
	var err error
	for _, zone := range []ZoneDescriptor{
		testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"),
		testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"),
	} {
		nodeA, err = RegisterZoneDescriptor(nodeA, zone)
		require.NoError(t, err)
	}
	for _, zone := range []ZoneDescriptor{
		testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"),
		testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"),
	} {
		nodeB, err = RegisterZoneDescriptor(nodeB, zone)
		require.NoError(t, err)
	}

	financial := testShardLayout(t, ZoneIDFinancial, 2, []ShardID{"1", "0"})
	identity := testShardLayout(t, ZoneIDIdentity, 1, []ShardID{"0"})
	for _, layout := range []ShardLayout{financial, identity} {
		nodeA, err = RegisterShardLayout(nodeA, layout)
		require.NoError(t, err)
	}
	for _, layout := range []ShardLayout{identity, financial} {
		nodeB, err = RegisterShardLayout(nodeB, layout)
		require.NoError(t, err)
	}

	tableA, err := BuildRoutingTableCommitment(3, 10, []ShardLayout{financial, identity})
	require.NoError(t, err)
	tableB, err := BuildRoutingTableCommitment(3, 10, []ShardLayout{identity, financial})
	require.NoError(t, err)
	require.Equal(t, tableA, tableB)

	nodeA, err = CommitRoutingTable(nodeA, tableA)
	require.NoError(t, err)
	nodeB, err = CommitRoutingTable(nodeB, tableB)
	require.NoError(t, err)
	require.Equal(t, nodeA.Export(), nodeB.Export())

	_, snapshot, err := CommitBlockRoots(nodeA, 10)
	require.NoError(t, err)
	require.True(t, hasProofRoot(snapshot, ShardLayoutRootType, financial.LayoutHash))
	require.True(t, hasProofRoot(snapshot, ShardLayoutRootType, identity.LayoutHash))
	require.True(t, hasProofRoot(snapshot, RoutingTableRootType, tableA.TableHash))
}

func TestRoutingTableRejectsLayoutMismatch(t *testing.T) {
	state := EmptyState()
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"))
	require.NoError(t, err)
	layout := testShardLayout(t, ZoneIDFinancial, 1, []ShardID{"0"})
	state, err = RegisterShardLayout(state, layout)
	require.NoError(t, err)

	table, err := NewRoutingTableCommitment(2, 5, []RoutingZoneEntry{{
		ZoneID:       ZoneIDFinancial,
		LayoutEpoch:  layout.LayoutEpoch,
		ActiveShards: 2,
		LayoutHash:   layout.LayoutHash,
	}})
	require.NoError(t, err)
	_, err = CommitRoutingTable(state, table)
	require.ErrorContains(t, err, "active shard count mismatch")
}

func TestProposalScheduleRequiresCommittedActiveShard(t *testing.T) {
	state := EmptyState()
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"))
	require.NoError(t, err)
	layout := testShardLayout(t, ZoneIDFinancial, 1, []ShardID{"0"})
	state, err = RegisterShardLayout(state, layout)
	require.NoError(t, err)

	schedule, err := BuildProposalSchedule(9, []ProposalItem{
		testProposalItem(ZoneIDFinancial, "0", "accepted", 1, 1, 0),
	}, TestnetParams())
	require.NoError(t, err)
	require.NoError(t, ValidateProposalScheduleForState(schedule, state))

	schedule, err = BuildProposalSchedule(9, []ProposalItem{
		testProposalItem(ZoneIDFinancial, "1", "missing-shard", 1, 1, 0),
	}, TestnetParams())
	require.NoError(t, err)
	require.ErrorContains(t, ValidateProposalScheduleForState(schedule, state), "not active")
}

func TestExportImportRoundTripDeterministic(t *testing.T) {
	state := populatedState(t, []ZoneID{ZoneIDFinancial, ZoneIDContract})
	root, err := BuildGlobalStateRoot(7, state, testContributions(7))
	require.NoError(t, err)
	state, err = AppendGlobalRoot(state, root)
	require.NoError(t, err)

	imported, err := ImportState(state.Export())
	require.NoError(t, err)
	require.Equal(t, state.Export(), imported.Export())
}

func TestInvalidCoreStateRejected(t *testing.T) {
	state := EmptyState()
	bad := testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial")
	bad.StateMachineVersion = 0
	_, err := RegisterZoneDescriptor(state, bad)
	require.ErrorContains(t, err, "state machine version")

	_, err = AppendZoneCommitment(state, testCommitment(t, 1, ZoneIDFinancial))
	require.ErrorContains(t, err, "not registered")

	_, err = BuildProposalSchedule(1, []ProposalItem{testProposalItem("bad zone", "0", "x", 1, 1, 0)}, TestnetParams())
	require.ErrorContains(t, err, "zone id")
}

func TestMessageReceiptAndProofRootsValidate(t *testing.T) {
	messageRoot, err := NewGlobalMessageRoot(1, testHash("inbox"), testHash("outbox"), 2)
	require.NoError(t, err)
	require.NoError(t, messageRoot.ValidateHash())

	receiptRoot, err := NewExecutionReceiptRoot(1, testHash("receipts"), 2)
	require.NoError(t, err)
	require.NoError(t, receiptRoot.Validate())

	proofRoot, err := NewProofRoot(1, MessageProofRootType, messageRoot.MessageRoot, "aethercore.global_messages")
	require.NoError(t, err)
	require.NoError(t, proofRoot.Validate())
}

func populatedState(t *testing.T, order []ZoneID) AetherCoreState {
	t.Helper()
	state := EmptyState()
	var err error
	for _, zoneID := range order {
		var zoneType ZoneType
		var name string
		switch zoneID {
		case ZoneIDFinancial:
			zoneType = ZoneTypeFinancial
			name = "financial"
		case ZoneIDContract:
			zoneType = ZoneTypeContract
			name = "contract"
		default:
			t.Fatalf("unsupported zone %s", zoneID)
		}
		state, err = RegisterZoneDescriptor(state, testDescriptor(zoneID, zoneType, name))
		require.NoError(t, err)
		state, err = AppendZoneCommitment(state, testCommitment(t, 7, zoneID))
		require.NoError(t, err)
	}
	return state
}

func testDescriptor(id ZoneID, zoneType ZoneType, moduleName string) ZoneDescriptor {
	return ZoneDescriptor{
		ZoneID:              id,
		ZoneType:            zoneType,
		ModuleName:          moduleName,
		Enabled:             true,
		StateMachineVersion: 1,
		MempoolPolicyID:     DefaultMempoolPolicy,
		FeePolicyID:         NativeFeePolicyID,
		ShardLayoutEpoch:    1,
		MaxShards:           4,
		MessageCapabilities: []string{"async-inbox", "async-outbox"},
		ProofCapabilities:   []string{"account", "message", "receipt"},
	}
}

func testService(serviceID string, zoneID ZoneID) ServiceDescriptor {
	return ServiceDescriptor{
		ServiceID:        serviceID,
		ZoneID:           zoneID,
		InterfaceID:      "l1.identity.v1.Query",
		EndpointKey:      "identity.query",
		Version:          1,
		AvailabilityHash: testHash(serviceID + "/availability"),
		Enabled:          true,
	}
}

func testCommitment(t *testing.T, height uint64, zoneID ZoneID) ZoneCommitment {
	t.Helper()
	commitment, err := NewZoneCommitment(
		height,
		zoneID,
		testHash(fmt.Sprintf("%d/%s/state", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/inbox", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/outbox", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/receipts", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/events", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/params", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/execution", height, zoneID)),
	)
	require.NoError(t, err)
	return commitment
}

func testShardLayout(t *testing.T, zoneID ZoneID, epoch uint64, shardIDs []ShardID) ShardLayout {
	t.Helper()
	shards := make([]ShardDescriptor, len(shardIDs))
	for i, shardID := range shardIDs {
		shards[i] = ShardDescriptor{
			ShardID:          shardID,
			StatePrefix:      fmt.Sprintf("zone/%s/shard/%s", zoneID, shardID),
			ActivationHeight: 1,
			ValidatorSetHash: testHash(fmt.Sprintf("%s/%s/validators", zoneID, shardID)),
			Available:        true,
		}
	}
	layout, err := NewShardLayout(zoneID, epoch, 1, testHash(fmt.Sprintf("%s/%d/routing-seed", zoneID, epoch)), shards)
	require.NoError(t, err)
	return layout
}

func testContributions(height uint64) RootContributions {
	return RootContributions{
		IdentityRoot: testHash(fmt.Sprintf("%d/identity", height)),
		StorageRoot:  testHash(fmt.Sprintf("%d/storage", height)),
		MessageRoot:  testHash(fmt.Sprintf("%d/messages", height)),
		ReceiptsRoot: testHash(fmt.Sprintf("%d/receipts", height)),
		PaymentsRoot: testHash(fmt.Sprintf("%d/payments", height)),
		VMRoot:       testHash(fmt.Sprintf("%d/vm", height)),
		ParamsHash:   testHash(fmt.Sprintf("%d/params", height)),
	}
}

func testProposalItem(zoneID ZoneID, shardID ShardID, seed string, priority uint32, height uint64, txIndex uint32) ProposalItem {
	return ProposalItem{
		ZoneID:          zoneID,
		ShardID:         shardID,
		TxHash:          testHash(seed),
		PriorityClass:   priority,
		AdmissionHeight: height,
		TxIndex:         txIndex,
	}
}

func hasProofRoot(snapshot RootSnapshot, rootType RootType, hash string) bool {
	for _, root := range snapshot.ProofRoots {
		if root.RootType == rootType && root.RootHash == hash {
			return true
		}
	}
	return false
}

func testHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
