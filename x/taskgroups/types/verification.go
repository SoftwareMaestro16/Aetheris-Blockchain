package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

const (
	VerificationDutyReexecuteStateTransition  = "reexecute_state_transition"
	VerificationDutyValidateCrossDomainProof  = "validate_cross_domain_proof"
	VerificationDutyVerifyTaskGroup           = "verify_task_group"
	VerificationDutyValidateConsensusOrdering = "validate_consensus_ordering"
	VerificationDutyVerifyMessageReceipt      = "verify_message_inclusion_receipt"
	VerificationDutySignValidOutput           = "sign_valid_output"
	VerificationDutyRejectInvalidOutput       = "reject_invalid_output"
	VerificationDutySubmitEvidence            = "submit_evidence"

	VerificationResultValid       = "valid"
	VerificationResultInvalid     = "invalid"
	VerificationResultAbstain     = "abstain"
	VerificationResultUnavailable = "unavailable"
)

type VerificationReceipt struct {
	EpochID            uint64
	TaskGroupID        string
	WorkloadID         string
	ValidatorAddress   string
	VerifiedObjectHash string
	Result             string
	Signature          string
	GasOrCostOptional  sdkmath.Int
	CreatedHeight      uint64
}

type VerificationReceiptSet struct {
	EpochID  uint64
	Receipts []VerificationReceipt
	Root     string
}

func RequiredVerificationDuties() []string {
	return []string{
		VerificationDutyReexecuteStateTransition,
		VerificationDutyValidateCrossDomainProof,
		VerificationDutyVerifyTaskGroup,
		VerificationDutyValidateConsensusOrdering,
		VerificationDutyVerifyMessageReceipt,
		VerificationDutySignValidOutput,
		VerificationDutyRejectInvalidOutput,
		VerificationDutySubmitEvidence,
	}
}

func NewVerificationReceipt(group postypes.TaskGroup, validatorAddress string, verifiedObjectHash string, result string, signature string, gasOrCostOptional sdkmath.Int, createdHeight uint64) (VerificationReceipt, error) {
	receipt := VerificationReceipt{
		EpochID:            group.EpochID,
		TaskGroupID:        group.TaskGroupID,
		WorkloadID:         group.WorkloadID,
		ValidatorAddress:   strings.TrimSpace(validatorAddress),
		VerifiedObjectHash: strings.TrimSpace(verifiedObjectHash),
		Result:             strings.TrimSpace(result),
		Signature:          strings.TrimSpace(signature),
		GasOrCostOptional:  gasOrCostOptional,
		CreatedHeight:      createdHeight,
	}
	return receipt, receipt.Validate(group)
}

func (r VerificationReceipt) Validate(group postypes.TaskGroup) error {
	if r.EpochID != group.EpochID {
		return errors.New("verification receipt epoch does not match task group")
	}
	if r.TaskGroupID != group.TaskGroupID {
		return errors.New("verification receipt task group mismatch")
	}
	if r.WorkloadID != group.WorkloadID {
		return errors.New("verification receipt workload mismatch")
	}
	if !containsString(group.ValidatorMembers, r.ValidatorAddress) {
		return errors.New("verification receipt validator is not assigned to task group")
	}
	if len(r.VerifiedObjectHash) != postypes.PosHashHexLength {
		return fmt.Errorf("verified object hash must be %d hex chars", postypes.PosHashHexLength)
	}
	if _, err := hex.DecodeString(r.VerifiedObjectHash); err != nil {
		return fmt.Errorf("verified object hash must be hex: %w", err)
	}
	if !isVerificationResult(r.Result) {
		return fmt.Errorf("unsupported verification result %q", r.Result)
	}
	if r.Signature == "" {
		return errors.New("verification receipt signature is required")
	}
	if r.GasOrCostOptional.IsNegative() {
		return errors.New("verification receipt gas or cost cannot be negative")
	}
	if r.CreatedHeight < group.ActivationHeight || r.CreatedHeight > group.ExpiryHeight {
		return errors.New("verification receipt height outside task group activity window")
	}
	return nil
}

func NewVerificationReceiptSet(group postypes.TaskGroup, receipts []VerificationReceipt) (VerificationReceiptSet, error) {
	ordered := make([]VerificationReceipt, len(receipts))
	copy(ordered, receipts)
	sortVerificationReceipts(ordered)
	set := VerificationReceiptSet{
		EpochID:  group.EpochID,
		Receipts: ordered,
		Root:     ComputeVerificationReceiptRoot(group, ordered),
	}
	return set, set.Validate(group)
}

func (s VerificationReceiptSet) Validate(group postypes.TaskGroup) error {
	if s.EpochID != group.EpochID {
		return errors.New("verification receipt set epoch mismatch")
	}
	seen := make(map[string]struct{}, len(s.Receipts))
	for i, receipt := range s.Receipts {
		if err := receipt.Validate(group); err != nil {
			return err
		}
		key := receipt.ValidatorAddress + "|" + receipt.VerifiedObjectHash
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate verification receipt %s", key)
		}
		seen[key] = struct{}{}
		if i > 0 && compareVerificationReceipts(s.Receipts[i-1], receipt) >= 0 {
			return errors.New("verification receipts must be sorted canonically")
		}
	}
	expectedRoot := ComputeVerificationReceiptRoot(group, s.Receipts)
	if s.Root != expectedRoot {
		return errors.New("verification receipt root mismatch")
	}
	return nil
}

func ComputeVerificationReceiptHash(receipt VerificationReceipt) string {
	h := sha256.New()
	writeHashPart(h, fmt.Sprintf("%d", receipt.EpochID))
	writeHashPart(h, receipt.TaskGroupID)
	writeHashPart(h, receipt.WorkloadID)
	writeHashPart(h, receipt.ValidatorAddress)
	writeHashPart(h, receipt.VerifiedObjectHash)
	writeHashPart(h, receipt.Result)
	writeHashPart(h, receipt.Signature)
	writeHashPart(h, receipt.GasOrCostOptional.String())
	writeHashPart(h, fmt.Sprintf("%d", receipt.CreatedHeight))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVerificationReceiptRoot(group postypes.TaskGroup, receipts []VerificationReceipt) string {
	h := sha256.New()
	writeHashPart(h, fmt.Sprintf("%d", group.EpochID))
	writeHashPart(h, group.TaskGroupID)
	writeHashPart(h, group.WorkloadID)
	writeHashPart(h, fmt.Sprintf("%d", len(receipts)))
	for _, receipt := range receipts {
		writeHashPart(h, ComputeVerificationReceiptHash(receipt))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func isVerificationResult(result string) bool {
	switch result {
	case VerificationResultValid, VerificationResultInvalid, VerificationResultAbstain, VerificationResultUnavailable:
		return true
	default:
		return false
	}
}

func sortVerificationReceipts(receipts []VerificationReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		return compareVerificationReceipts(receipts[i], receipts[j]) < 0
	})
}

func compareVerificationReceipts(left VerificationReceipt, right VerificationReceipt) int {
	if left.ValidatorAddress < right.ValidatorAddress {
		return -1
	}
	if left.ValidatorAddress > right.ValidatorAddress {
		return 1
	}
	if left.VerifiedObjectHash < right.VerifiedObjectHash {
		return -1
	}
	if left.VerifiedObjectHash > right.VerifiedObjectHash {
		return 1
	}
	return 0
}
