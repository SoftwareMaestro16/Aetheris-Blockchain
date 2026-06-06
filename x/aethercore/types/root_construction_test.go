package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootConstructionAggregatesModuleAndZoneRootsCanonically(t *testing.T) {
	accounts := testRootContribution(t, StateProofRootType, "accounts", "accounts/root")
	messages := testRootContribution(t, MessageProofRootType, "outbox", "outbox/root")
	receipts := testRootContribution(t, ReceiptProofRootType, "receipts", "receipts/root")

	left, err := BuildZoneRootAggregation(17, ZoneIDFinancial, []RootContribution{receipts, accounts, messages})
	require.NoError(t, err)
	require.NoError(t, left.Validate())
	require.Equal(t, []RootContribution{messages, receipts, accounts}, left.ModuleRoots)

	right, err := BuildZoneRootAggregation(17, ZoneIDFinancial, []RootContribution{messages, receipts, accounts})
	require.NoError(t, err)
	require.Equal(t, left.ZoneRoot, right.ZoneRoot)

	tampered := left
	tampered.ModuleRoots[0].RootHash = testHash("tampered-module-root")
	tampered.ModuleRoots[0].ContributionHash = ComputeRootContributionHash(tampered.ModuleRoots[0])
	require.ErrorContains(t, tampered.Validate(), "zone aggregate root mismatch")
}

func TestAEKRootAggregationOrdersZoneAndGlobalRootsLexicographically(t *testing.T) {
	financialZone := testRootContribution(t, RootType("zone"), string(ZoneIDFinancial), "financial/zone/root")
	contractZone := testRootContribution(t, RootType("zone"), string(ZoneIDContract), "contract/zone/root")
	services := testRootContribution(t, RootType("services"), "global", "services/root")
	routing := testRootContribution(t, RootType("routing"), "global", "routing/root")

	left, err := BuildAEKRootAggregation(18, []RootContribution{financialZone, contractZone}, []RootContribution{services, routing})
	require.NoError(t, err)
	require.NoError(t, left.Validate())
	require.Equal(t, []RootContribution{contractZone, financialZone}, left.ZoneRoots)
	require.Equal(t, []RootContribution{routing, services}, left.GlobalRoots)

	right, err := BuildAEKRootAggregation(18, []RootContribution{contractZone, financialZone}, []RootContribution{routing, services})
	require.NoError(t, err)
	require.Equal(t, left.AggregateRoot, right.AggregateRoot)

	_, err = BuildAEKRootAggregation(18, []RootContribution{contractZone, contractZone}, nil)
	require.ErrorContains(t, err, "duplicate")
}

func TestDeterministicEmptyRootCommitmentUsedForEmptyModuleRoots(t *testing.T) {
	emptyA, err := NewRootContribution(RootType("storage"), "archive-index", "")
	require.NoError(t, err)
	emptyB, err := NewRootContribution(RootType("storage"), "archive-index", "")
	require.NoError(t, err)
	other, err := NewRootContribution(RootType("storage"), "object-index", "")
	require.NoError(t, err)

	require.Equal(t, emptyA.RootHash, emptyB.RootHash)
	require.NotEqual(t, EmptyRootHash, emptyA.RootHash)
	require.NotEqual(t, emptyA.RootHash, other.RootHash)
	require.Equal(t, ComputeRootContributionHash(emptyA), emptyA.ContributionHash)
}

func TestStateCommitmentProofTypesAndNonExistenceProofValidate(t *testing.T) {
	path := []string{testHash("proof/sibling/0"), testHash("proof/sibling/1")}
	for _, proofType := range []CommitmentProofType{
		ZoneProofType,
		ServiceProofType,
		IdentityProofType,
		StorageProofType,
		MessageProofType,
		ReceiptProofType,
		PaymentProofType,
		ContractProofType,
		RoutingProofType,
	} {
		proof, err := NewStateCommitmentProof(proofType, 19, RootType(proofType), "subject", testHash(string(proofType)+"/key"), testHash(string(proofType)+"/value"), testHash(string(proofType)+"/root"), path, true)
		require.NoError(t, err)
		require.NoError(t, proof.Validate())

		tampered := proof
		tampered.ValueHash = testHash(string(proofType) + "/tampered")
		require.ErrorContains(t, tampered.Validate(), "proof hash mismatch")
	}

	var zoneProof ZoneProof
	zoneProof, err := NewStateCommitmentProof(ZoneProofType, 19, RootType("zone"), "FINANCIAL_ZONE", testHash("zone/key"), testHash("zone/value"), testHash("zone/root"), path, true)
	require.NoError(t, err)
	require.NoError(t, zoneProof.Validate())

	nonExistence, err := NewNonExistenceProof(20, RootType("identity"), "missing.aet", testHash("missing/key"), testHash("identity/root"), path)
	require.NoError(t, err)
	require.False(t, nonExistence.Exists)
	require.Equal(t, NonExistenceProofType, nonExistence.ProofType)
	require.NoError(t, nonExistence.Validate())

	invalid := nonExistence
	invalid.Exists = true
	invalid.ProofHash = ComputeStateCommitmentProofHash(invalid)
	require.ErrorContains(t, invalid.Validate(), "non-existence proof")
}

func testRootContribution(t *testing.T, rootType RootType, id string, seed string) RootContribution {
	t.Helper()
	contribution, err := NewRootContribution(rootType, id, testHash(seed))
	require.NoError(t, err)
	return contribution
}
