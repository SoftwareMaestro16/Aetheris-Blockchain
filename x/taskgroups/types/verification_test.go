package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestVerificationReceiptAcceptsAllResultValuesAndHashesDeterministically(t *testing.T) {
	group := proposerTestGroup()
	objectHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	results := []string{
		VerificationResultValid,
		VerificationResultInvalid,
		VerificationResultAbstain,
		VerificationResultUnavailable,
	}
	for _, result := range results {
		receipt, err := NewVerificationReceipt(group, "val-a", objectHash, result, "sig-"+result, sdkmath.NewInt(10), 40)
		require.NoError(t, err)
		require.Equal(t, group.EpochID, receipt.EpochID)
		require.Equal(t, group.TaskGroupID, receipt.TaskGroupID)
		require.Equal(t, group.WorkloadID, receipt.WorkloadID)
		require.Len(t, ComputeVerificationReceiptHash(receipt), 64)
	}
}

func TestVerificationReceiptRejectsInvalidMembershipSignatureHashAndWindow(t *testing.T) {
	group := proposerTestGroup()
	objectHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	_, err := NewVerificationReceipt(group, "val-x", objectHash, VerificationResultValid, "sig", sdkmath.ZeroInt(), 40)
	require.ErrorContains(t, err, "not assigned")

	_, err = NewVerificationReceipt(group, "val-a", "bad", VerificationResultValid, "sig", sdkmath.ZeroInt(), 40)
	require.ErrorContains(t, err, "hex chars")

	_, err = NewVerificationReceipt(group, "val-a", objectHash, "maybe", "sig", sdkmath.ZeroInt(), 40)
	require.ErrorContains(t, err, "unsupported verification result")

	_, err = NewVerificationReceipt(group, "val-a", objectHash, VerificationResultValid, "", sdkmath.ZeroInt(), 40)
	require.ErrorContains(t, err, "signature")

	_, err = NewVerificationReceipt(group, "val-a", objectHash, VerificationResultValid, "sig", sdkmath.NewInt(-1), 40)
	require.ErrorContains(t, err, "gas or cost")

	_, err = NewVerificationReceipt(group, "val-a", objectHash, VerificationResultValid, "sig", sdkmath.ZeroInt(), 61)
	require.ErrorContains(t, err, "activity window")
}

func TestVerificationReceiptSetSortsAndRootsReceipts(t *testing.T) {
	group := proposerTestGroup()
	hashA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	hashB := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	receiptB, err := NewVerificationReceipt(group, "val-b", hashB, VerificationResultValid, "sig-b", sdkmath.NewInt(20), 40)
	require.NoError(t, err)
	receiptA, err := NewVerificationReceipt(group, "val-a", hashA, VerificationResultInvalid, "sig-a", sdkmath.NewInt(10), 40)
	require.NoError(t, err)

	left, err := NewVerificationReceiptSet(group, []VerificationReceipt{receiptB, receiptA})
	require.NoError(t, err)
	right, err := NewVerificationReceiptSet(group, []VerificationReceipt{receiptA, receiptB})
	require.NoError(t, err)
	require.Equal(t, left.Root, right.Root)
	require.Equal(t, []VerificationReceipt{receiptA, receiptB}, left.Receipts)
	require.NoError(t, left.Validate(group))

	_, err = NewVerificationReceiptSet(group, []VerificationReceipt{receiptA, receiptA})
	require.ErrorContains(t, err, "duplicate verification receipt")
}

func TestRequiredVerificationDutiesEnumeratesValidatorDuties(t *testing.T) {
	require.Equal(t, []string{
		VerificationDutyReexecuteStateTransition,
		VerificationDutyValidateCrossDomainProof,
		VerificationDutyVerifyTaskGroup,
		VerificationDutyValidateConsensusOrdering,
		VerificationDutyVerifyMessageReceipt,
		VerificationDutySignValidOutput,
		VerificationDutyRejectInvalidOutput,
		VerificationDutySubmitEvidence,
	}, RequiredVerificationDuties())
}
