package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

const (
	ProposerStatusReady       = "ready"
	ProposerStatusUnavailable = "unavailable"
	ProposerStatusFallback    = "fallback"
)

type ProposerPriority struct {
	EpochID          uint64
	Slot             uint64
	TaskGroupID      string
	ValidatorAddress string
	PriorityScore    sdkmath.Int
	FallbackOrder    uint32
	ProposerStatus   string
}

type ProposerPriorityInput struct {
	ValidatorScore              sdkmath.Int
	PriorProposerPerformanceBps uint32
	MissedProposalCount         uint64
	TaskReliabilityBps          uint32
	StakeSaturationDampeningBps uint32
}

type ProposerSelectionInput struct {
	Group           postypes.TaskGroup
	ValidatorScores map[string]sdkmath.Int
	PriorityInputs  map[string]ProposerPriorityInput
	Unavailable     map[string]bool
}

type ProposerSelection struct {
	EpochID            uint64
	Slot               uint64
	TaskGroupID        string
	CanonicalProposer  string
	VerifierValidators []string
	Priorities         []ProposerPriority
	CanonicalPriority  ProposerPriority
	FallbackUsed       bool
}

func BuildProposerPriorities(input ProposerSelectionInput, slot uint64) ([]ProposerPriority, error) {
	if slot == 0 {
		return nil, errors.New("proposer slot is required")
	}
	if err := input.Group.Validate(); err != nil {
		return nil, err
	}
	if len(input.Group.ProposerOrder) == 0 {
		return nil, errors.New("task group proposer order is required")
	}
	priorities := make([]ProposerPriority, 0, len(input.Group.ProposerOrder))
	for fallbackOrder, validatorID := range input.Group.ProposerOrder {
		priorityInput := input.PriorityInputs[validatorID]
		if priorityInput.ValidatorScore.IsNil() {
			priorityInput.ValidatorScore = input.ValidatorScores[validatorID]
		}
		if priorityInput.PriorProposerPerformanceBps == 0 {
			priorityInput.PriorProposerPerformanceBps = postypes.BasisPoints
		}
		if priorityInput.TaskReliabilityBps == 0 {
			priorityInput.TaskReliabilityBps = postypes.BasisPoints
		}
		if priorityInput.StakeSaturationDampeningBps == 0 {
			priorityInput.StakeSaturationDampeningBps = postypes.BasisPoints
		}
		score, err := ComputeProposerPriorityScore(priorityInput)
		if err != nil {
			return nil, err
		}
		status := ProposerStatusReady
		if input.Unavailable[validatorID] {
			status = ProposerStatusUnavailable
		}
		priorities = append(priorities, ProposerPriority{
			EpochID:          input.Group.EpochID,
			Slot:             slot,
			TaskGroupID:      input.Group.TaskGroupID,
			ValidatorAddress: validatorID,
			PriorityScore:    score,
			FallbackOrder:    uint32(fallbackOrder),
			ProposerStatus:   status,
		})
	}
	sortProposerPriorities(priorities)
	for i := range priorities {
		priorities[i].FallbackOrder = uint32(i)
	}
	return priorities, nil
}

func SelectCanonicalProposer(input ProposerSelectionInput, slot uint64) (ProposerSelection, error) {
	priorities, err := BuildProposerPriorities(input, slot)
	if err != nil {
		return ProposerSelection{}, err
	}
	var selected ProposerPriority
	found := false
	for _, priority := range priorities {
		if priority.ProposerStatus == ProposerStatusUnavailable {
			continue
		}
		selected = priority
		found = true
		break
	}
	if !found {
		return ProposerSelection{}, errors.New("no available proposer for slot")
	}
	fallbackUsed := false
	if selected.FallbackOrder != 0 || priorities[0].ValidatorAddress != selected.ValidatorAddress {
		fallbackUsed = true
		selected.ProposerStatus = ProposerStatusFallback
	}
	for i := range priorities {
		if priorities[i].ValidatorAddress == selected.ValidatorAddress {
			priorities[i].ProposerStatus = selected.ProposerStatus
			break
		}
	}
	verifiers := make([]string, 0, len(input.Group.ValidatorMembers)-1)
	for _, validatorID := range input.Group.ValidatorMembers {
		if validatorID != selected.ValidatorAddress {
			verifiers = append(verifiers, validatorID)
		}
	}
	sort.Strings(verifiers)
	return ProposerSelection{
		EpochID:            input.Group.EpochID,
		Slot:               slot,
		TaskGroupID:        input.Group.TaskGroupID,
		CanonicalProposer:  selected.ValidatorAddress,
		VerifierValidators: verifiers,
		Priorities:         priorities,
		CanonicalPriority:  selected,
		FallbackUsed:       fallbackUsed,
	}, nil
}

func ComputeProposerPriorityScore(input ProposerPriorityInput) (sdkmath.Int, error) {
	if input.ValidatorScore.IsNil() || input.ValidatorScore.IsNegative() {
		return sdkmath.Int{}, errors.New("validator score must be non-negative")
	}
	if input.PriorProposerPerformanceBps > postypes.BasisPoints {
		return sdkmath.Int{}, fmt.Errorf("prior proposer performance must be <= %d bps", postypes.BasisPoints)
	}
	if input.TaskReliabilityBps > postypes.BasisPoints {
		return sdkmath.Int{}, fmt.Errorf("task reliability must be <= %d bps", postypes.BasisPoints)
	}
	if input.StakeSaturationDampeningBps > postypes.BasisPoints {
		return sdkmath.Int{}, fmt.Errorf("stake saturation dampening must be <= %d bps", postypes.BasisPoints)
	}
	score := mulIntBps(input.ValidatorScore, input.PriorProposerPerformanceBps)
	score = mulIntBps(score, input.TaskReliabilityBps)
	score = mulIntBps(score, input.StakeSaturationDampeningBps)
	if input.MissedProposalCount == 0 {
		return score, nil
	}
	penaltyBps := input.MissedProposalCount * 1_000
	if penaltyBps >= uint64(postypes.BasisPoints) {
		return sdkmath.ZeroInt(), nil
	}
	return mulIntBps(score, uint32(uint64(postypes.BasisPoints)-penaltyBps)), nil
}

func (p ProposerPriority) Validate() error {
	if p.EpochID == 0 {
		return errors.New("proposer priority epoch id is required")
	}
	if p.Slot == 0 {
		return errors.New("proposer priority slot is required")
	}
	if strings.TrimSpace(p.TaskGroupID) == "" {
		return errors.New("proposer priority task group id is required")
	}
	if strings.TrimSpace(p.ValidatorAddress) == "" {
		return errors.New("proposer priority validator address is required")
	}
	if p.PriorityScore.IsNil() || p.PriorityScore.IsNegative() {
		return errors.New("proposer priority score must be non-negative")
	}
	switch p.ProposerStatus {
	case ProposerStatusReady, ProposerStatusUnavailable, ProposerStatusFallback:
		return nil
	default:
		return fmt.Errorf("unsupported proposer status %q", p.ProposerStatus)
	}
}

func sortProposerPriorities(priorities []ProposerPriority) {
	sort.SliceStable(priorities, func(i, j int) bool {
		left := priorities[i]
		right := priorities[j]
		if !left.PriorityScore.Equal(right.PriorityScore) {
			return left.PriorityScore.GT(right.PriorityScore)
		}
		if left.FallbackOrder != right.FallbackOrder {
			return left.FallbackOrder < right.FallbackOrder
		}
		return left.ValidatorAddress < right.ValidatorAddress
	})
}

func mulIntBps(value sdkmath.Int, bps uint32) sdkmath.Int {
	return value.MulRaw(int64(bps)).QuoRaw(int64(postypes.BasisPoints))
}
