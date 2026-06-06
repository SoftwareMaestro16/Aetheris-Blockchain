package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityAdaptiveSyncSnapshotRestoreProofReadyWithAuctionsAndExpiryV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt-a", 10)
	state, _ = registerSpecDomainInState(t, state, "bob", addr(2), "salt-b", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(3)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(3), addr(3), "alice.aet", 13)
	require.NoError(t, err)

	for i := range state.Domains {
		if state.Domains[i].Name == "bob.aet" {
			state.Domains[i].ExpiryHeight = 30
			state.Domains[i].UpdatedHeight = 14
		}
	}
	queuedCommitment, err := ComputeRegistrationCommitment("queued.aet", addr(8), "queued")
	require.NoError(t, err)
	state, err = CommitDomainRegistration(state, "queued.aet", addr(8), queuedCommitment, 18)
	require.NoError(t, err)
	state, auction, err := StartSealedAuction(state, "market", 20)
	require.NoError(t, err)
	bidCommitment, err := ComputeAuctionCommitment("market.aet", addr(4), 100, "left")
	require.NoError(t, err)
	state, auction, err = CommitAuctionBid(state, auction.Name, addr(4), bidCommitment, 21)
	require.NoError(t, err)
	require.Len(t, auction.Commitments, 1)

	delegation, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeResolverUpdate, []string{ResolverKeyPrimary}, 80, 0, ResolverKeyPrimary, 15)
	require.NoError(t, err)
	snapshot, err := BuildIdentityAdaptiveSyncSnapshotV2(state, []DelegationRecordV2{delegation}, 25)
	require.NoError(t, err)
	require.Len(t, snapshot.State.Domains, 2)
	require.Len(t, snapshot.State.Resolvers, 1)
	require.Len(t, snapshot.State.ReverseRecords, 1)
	require.Len(t, snapshot.State.DomainNFTs, 2)
	require.Len(t, snapshot.State.Auctions, 1)
	require.Len(t, snapshot.State.Commits, 1)
	require.Len(t, snapshot.Delegations, 1)
	require.Len(t, snapshot.ExpiryIndex, len(snapshot.State.Domains))
	require.NotEmpty(t, snapshot.StateRoot)
	require.NotEmpty(t, snapshot.ExpiryIndexRoot)
	require.NotEmpty(t, snapshot.SnapshotHash)

	restored, err := RestoreIdentityAdaptiveSyncSnapshotV2(snapshot, []string{"alice.aet"})
	require.NoError(t, err)
	require.True(t, restored.ProofReady)
	require.Equal(t, snapshot.StateRoot, restored.StateRoot)
	require.Len(t, restored.Delegations, 1)
	require.Len(t, restored.Events, 2)
	require.Equal(t, IdentityAdaptiveSyncEventRecoveredV2, restored.Events[0].Type)
	require.Equal(t, IdentityAdaptiveSyncEventCacheResyncV2, restored.Events[1].Type)

	proof, err := BuildIdentityResolutionProof(restored.State, "alice.aet", 25)
	require.NoError(t, err)
	_, err = VerifyIdentityResolutionProof(proof, 25)
	require.NoError(t, err)
}

func TestIdentityAdaptiveSyncSnapshotRejectsBrokenInvariantV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt-a", 10)
	state.DomainNFTs[0].Owner = addr(9)

	_, err := BuildIdentityAdaptiveSyncSnapshotV2(state, nil, 20)
	require.ErrorContains(t, err, "ownership mismatch")
}

func TestIdentityAdaptiveSyncStateExportCompatibilityV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt-a", 10)
	state, _ = registerSpecDomainInState(t, state, "bob", addr(2), "salt-b", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(3)}, 12)
	require.NoError(t, err)
	leftDelegation, err := NewDelegationRecordV2("bob.aet", addr(8), DelegationScopeSubdomainCreate, []string{"create"}, 90, 1, "", 15)
	require.NoError(t, err)
	rightDelegation, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeResolverUpdate, []string{ResolverKeyPrimary}, 80, 0, ResolverKeyPrimary, 15)
	require.NoError(t, err)

	snapshot, err := BuildIdentityAdaptiveSyncSnapshotV2(state, []DelegationRecordV2{leftDelegation, rightDelegation}, 25)
	require.NoError(t, err)
	imported, err := ImportIdentityState(snapshot.State)
	require.NoError(t, err)
	require.Equal(t, snapshot.State.Export(), imported)

	reordered, err := BuildIdentityAdaptiveSyncSnapshotV2(snapshot.State, []DelegationRecordV2{rightDelegation, leftDelegation}, 25)
	require.NoError(t, err)
	require.Equal(t, snapshot.StateRoot, reordered.StateRoot)
	require.Equal(t, snapshot.DelegationRoot, reordered.DelegationRoot)
	require.Equal(t, snapshot.ExpiryIndexRoot, reordered.ExpiryIndexRoot)
	require.Equal(t, snapshot.SnapshotHash, reordered.SnapshotHash)
}

func TestIdentityAdaptiveSyncCacheResyncPlanV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt-a", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(3)}, 12)
	require.NoError(t, err)
	snapshot, err := BuildIdentityAdaptiveSyncSnapshotV2(state, nil, 25)
	require.NoError(t, err)

	plan, err := BuildIdentityAdaptiveSyncCacheResyncPlanV2(snapshot, []string{"alice.aet"})
	require.NoError(t, err)
	require.Equal(t, uint64(25), plan.Height)
	require.Equal(t, []string{"alice.aet"}, plan.QueryNames)
	require.Len(t, plan.Events, 1)
	require.Equal(t, IdentityAdaptiveSyncEventCacheResyncV2, plan.Events[0].Type)
	require.Equal(t, "alice.aet", plan.Events[0].Name)
	require.Equal(t, plan.NameHash, plan.Events[0].NameHash)
	require.Contains(t, plan.Events[0].Attributes[0], snapshot.SnapshotHash)
}
