package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityBlockSTMParallelSafeMessageClassesV2(t *testing.T) {
	aliceHash := mustDomainHashV2(t, "alice.aet")
	bobHash := mustDomainHashV2(t, "bob.aet")
	aliceRegister := MsgRegisterDirectV2{Auth: txAuth(IdentitySignerScopeRegistration, 1), Name: "alice.aet", NameHash: aliceHash, Owner: addr(1), ExpectedRecordVersion: 1}
	bobRegister := MsgRegisterDirectV2{Auth: txAuth(IdentitySignerScopeRegistration, 2), Name: "bob.aet", NameHash: bobHash, Owner: addr(2), ExpectedRecordVersion: 1}
	alicePlan := mustBlockSTMPlanV2(t, aliceRegister)
	bobPlan := mustBlockSTMPlanV2(t, bobRegister)
	require.False(t, alicePlan.AccessSet.Conflicts(bobPlan.AccessSet))
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictNoneV2), IdentityBlockSTMConflictClassifyV2(alicePlan, bobPlan))
	require.NotContains(t, alicePlan.AccessSet.Writes, alicePlan.FeeKey)

	aliceResolver := MsgUpdateResolverRecordV2{Auth: txAuth(IdentitySignerScopeResolverUpdate, 3), Name: "alice.aet", NameHash: aliceHash, Patch: ResolverPatch{Contract: addr(3)}, ExpectedRecordVersion: 1, RecordTTL: 30}
	bobResolver := MsgUpdateResolverRecordV2{Auth: txAuth(IdentitySignerScopeResolverUpdate, 4), Name: "bob.aet", NameHash: bobHash, Patch: ResolverPatch{Contract: addr(4)}, ExpectedRecordVersion: 1, RecordTTL: 30}
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictNoneV2), IdentityBlockSTMConflictClassifyV2(mustBlockSTMPlanV2(t, aliceResolver), mustBlockSTMPlanV2(t, bobResolver)))

	aliceReverse, err := NewReverseResolutionRecordV2(addr(11), "alice.aet", false, 12, 100)
	require.NoError(t, err)
	bobReverse, err := NewReverseResolutionRecordV2(addr(12), "bob.aet", false, 12, 100)
	require.NoError(t, err)
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictNoneV2), IdentityBlockSTMConflictClassifyV2(
		mustBlockSTMPlanV2(t, MsgSetReverseRecordV2{Auth: txAuth(IdentitySignerScopeReverseUpdate, 5), Record: aliceReverse, ExpectedRecordVersion: 1}),
		mustBlockSTMPlanV2(t, MsgSetReverseRecordV2{Auth: txAuth(IdentitySignerScopeReverseUpdate, 6), Record: bobReverse, ExpectedRecordVersion: 1}),
	))

	parentHash := mustDomainHashV2(t, "parent.aet")
	otherHash := mustDomainHashV2(t, "other.aet")
	leftChild := MsgCreateSubdomainV2{Auth: txAuth(IdentitySignerScopeSubdomainAdmin, 7), ParentName: "parent.aet", ParentNameHash: parentHash, Label: "api", ChildOwner: addr(7), ExpectedParentVersion: 1}
	rightChild := MsgCreateSubdomainV2{Auth: txAuth(IdentitySignerScopeSubdomainAdmin, 8), ParentName: "other.aet", ParentNameHash: otherHash, Label: "api", ChildOwner: addr(8), ExpectedParentVersion: 1}
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictNoneV2), IdentityBlockSTMConflictClassifyV2(mustBlockSTMPlanV2(t, leftChild), mustBlockSTMPlanV2(t, rightChild)))

	leftReveal := MsgRevealBidV2{Auth: txAuth(IdentitySignerScopeAuctionBidder, 9), AuctionID: identityHash("auction-left"), NameHash: aliceHash, Bid: 100, Salt: "salt-left", CommitmentHash: identityHash("commit-left")}
	rightFinalize := MsgFinalizeAuctionV2{Auth: txAuth(IdentitySignerScopeAuctionAdmin, 10), AuctionID: identityHash("auction-right"), NameHash: bobHash, ExpectedAuctionVersion: 1}
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictNoneV2), IdentityBlockSTMConflictClassifyV2(mustBlockSTMPlanV2(t, leftReveal), mustBlockSTMPlanV2(t, rightFinalize)))
}

func TestIdentityBlockSTMConflictProneClassesV2(t *testing.T) {
	nameHash := mustDomainHashV2(t, "alice.aet")
	left := mustBlockSTMPlanV2(t, MsgUpdateResolverRecordV2{Auth: txAuth(IdentitySignerScopeResolverUpdate, 1), Name: "alice.aet", NameHash: nameHash, Patch: ResolverPatch{Contract: addr(3)}, ExpectedRecordVersion: 1, RecordTTL: 30})
	right := mustBlockSTMPlanV2(t, MsgUpdateResolverRecordV2{Auth: txAuth(IdentitySignerScopeResolverUpdate, 2), Name: "alice.aet", NameHash: nameHash, Patch: ResolverPatch{ZoneEndpoint: "zone.alice"}, ExpectedRecordVersion: 1, RecordTTL: 30})
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictSameNameV2), IdentityBlockSTMConflictClassifyV2(left, right))

	transfer := mustBlockSTMPlanV2(t, MsgTransferDomainV2{Auth: txAuth(IdentitySignerScopeOwner, 3), Name: "alice.aet", NameHash: nameHash, NewOwner: addr(9), ExpectedRecordVersion: 1})
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictTransferResolverUpdateV2), IdentityBlockSTMConflictClassifyV2(transfer, left))

	reverse, err := NewReverseResolutionRecordV2(addr(4), "alice.aet", true, 12, 100)
	require.NoError(t, err)
	primaryUpdate := mustBlockSTMPlanV2(t, MsgUpdateResolverRecordV2{Auth: txAuth(IdentitySignerScopeResolverUpdate, 4), Name: "alice.aet", NameHash: nameHash, Patch: ResolverPatch{Primary: addr(4)}, ExpectedRecordVersion: 1, RecordTTL: 30})
	reverseVerify := mustBlockSTMPlanV2(t, MsgVerifyReverseRecordV2{Auth: txAuth(IdentitySignerScopeReverseUpdate, 5), Record: reverse, ExpectedRecordVersion: 1})
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictReversePrimaryResolverV2), IdentityBlockSTMConflictClassifyV2(primaryUpdate, reverseVerify))

	auctionID := identityHash("auction")
	lateReveal := mustBlockSTMPlanV2(t, MsgRevealBidV2{Auth: txAuth(IdentitySignerScopeAuctionBidder, 6), AuctionID: auctionID, NameHash: nameHash, Bid: 100, Salt: "late", CommitmentHash: identityHash("commit-late")})
	finalize := mustBlockSTMPlanV2(t, MsgFinalizeAuctionV2{Auth: txAuth(IdentitySignerScopeAuctionAdmin, 7), AuctionID: auctionID, NameHash: nameHash, ExpectedAuctionVersion: 1})
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictAuctionFinalizeLateRevealV2), IdentityBlockSTMConflictClassifyV2(lateReveal, finalize))
}

func TestIdentityBlockSTMParentChildConflictAndVersionedRejectionV2(t *testing.T) {
	parentHash := mustDomainHashV2(t, "parent.aet")
	delegation, err := NewDelegationRecordV2("parent.aet", addr(7), DelegationScopeZoneAdmin, []string{"create"}, 100, 2, "", 10)
	require.NoError(t, err)
	policyUpdate := mustBlockSTMPlanV2(t, MsgDelegateSubdomainV2{Auth: txAuth(IdentitySignerScopeDelegationAdmin, 1), Delegation: delegation, ExpectedRecordVersion: 1})
	childCreate := mustBlockSTMPlanV2(t, MsgCreateSubdomainV2{Auth: txAuth(IdentitySignerScopeSubdomainAdmin, 2), ParentName: "parent.aet", ParentNameHash: parentHash, Label: "api", ChildOwner: addr(8), ExpectedParentVersion: 1})
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictParentPolicyChildCreateV2), IdentityBlockSTMConflictClassifyV2(policyUpdate, childCreate))

	require.NoError(t, IdentityBlockSTMValidateVersionedUpdateV2(7, 7))
	require.ErrorContains(t, IdentityBlockSTMValidateVersionedUpdateV2(8, 7), "version conflict")
	require.ErrorContains(t, IdentityBlockSTMValidateVersionedUpdateV2(0, 7), "current record version")

	batch := MsgBatchUpdateResolversV2{Auth: txAuth(IdentitySignerScopeBatchAdmin, 3), Updates: []ResolverBatchUpdateV2{
		{Name: "alice.aet", NameHash: mustDomainHashV2(t, "alice.aet"), Patch: ResolverPatch{Primary: addr(1)}, ExpectedRecordVersion: 1, RecordTTL: 30},
		{Name: "alice.aet", NameHash: mustDomainHashV2(t, "alice.aet"), Patch: ResolverPatch{Contract: addr(2)}, ExpectedRecordVersion: 1, RecordTTL: 30},
	}}
	_, _, err = IdentityBlockSTMBatchResolverAccessSetV2(batch)
	require.ErrorContains(t, err, "duplicate domain")
}

func TestIdentityBlockSTMBatchResolverUpdatesUseDisjointNameHashesV2(t *testing.T) {
	left := MsgBatchUpdateResolversV2{Auth: txAuth(IdentitySignerScopeBatchAdmin, 1), Updates: []ResolverBatchUpdateV2{
		{Name: "alice.aet", NameHash: mustDomainHashV2(t, "alice.aet"), Patch: ResolverPatch{Primary: addr(1)}, ExpectedRecordVersion: 1, RecordTTL: 30},
		{Name: "bob.aet", NameHash: mustDomainHashV2(t, "bob.aet"), Patch: ResolverPatch{Primary: addr(2)}, ExpectedRecordVersion: 1, RecordTTL: 30},
	}}
	right := MsgBatchUpdateResolversV2{Auth: txAuth(IdentitySignerScopeBatchAdmin, 2), Updates: []ResolverBatchUpdateV2{
		{Name: "carol.aet", NameHash: mustDomainHashV2(t, "carol.aet"), Patch: ResolverPatch{Primary: addr(3)}, ExpectedRecordVersion: 1, RecordTTL: 30},
		{Name: "dave.aet", NameHash: mustDomainHashV2(t, "dave.aet"), Patch: ResolverPatch{Primary: addr(4)}, ExpectedRecordVersion: 1, RecordTTL: 30},
	}}
	leftPlan := mustBlockSTMPlanV2(t, left)
	rightPlan := mustBlockSTMPlanV2(t, right)
	require.Equal(t, []string{mustDomainHashV2(t, "alice.aet"), mustDomainHashV2(t, "bob.aet")}, leftPlan.NameHashes)
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictNoneV2), IdentityBlockSTMConflictClassifyV2(leftPlan, rightPlan))
	require.NotEqual(t, "", leftPlan.FeeKey)
	require.NotContains(t, leftPlan.AccessSet.Writes, leftPlan.FeeKey)

	overlap := MsgBatchUpdateResolversV2{Auth: txAuth(IdentitySignerScopeBatchAdmin, 3), Updates: []ResolverBatchUpdateV2{
		{Name: "bob.aet", NameHash: mustDomainHashV2(t, "bob.aet"), Patch: ResolverPatch{Primary: addr(5)}, ExpectedRecordVersion: 1, RecordTTL: 30},
	}}
	require.Equal(t, IdentityBlockSTMConflictClassV2(IdentityBlockSTMConflictSameNameV2), IdentityBlockSTMConflictClassifyV2(leftPlan, mustBlockSTMPlanV2(t, overlap)))
}

func mustBlockSTMPlanV2(t *testing.T, msg IdentityMsgV2) IdentityBlockSTMPlanV2 {
	t.Helper()
	plan, err := IdentityBlockSTMAccessSetV2(msg, 100)
	require.NoError(t, err)
	return plan
}

func mustDomainHashV2(t *testing.T, name string) string {
	t.Helper()
	hash, err := DomainRecordV2NameHash(name)
	require.NoError(t, err)
	return hash
}
