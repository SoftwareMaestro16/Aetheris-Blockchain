package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusRejected = "rejected"
	StatusExpired  = "expired"

	EvidenceTypeConsensus   = "consensus"
	EvidenceTypeMissedBlock = "missed-block"
	EvidenceTypePerformance = "performance"
	EvidenceTypeFraud       = "fraud"

	VoteSupportAccept = "accept"
	VoteSupportReject = "reject"

	RegistryStatusJailed     = validatorregistrytypes.StatusJailed
	RegistryStatusTombstoned = validatorregistrytypes.StatusTombstoned

	MaxEvidenceV1          = uint32(100_000)
	MaxPendingEvidenceV1   = uint32(10_000)
	MaxProofHashBytesV1    = uint32(64)
	MaxPayloadBytesV1      = uint32(16_384)
	MaxVotesV1             = uint32(512)
	MaxSideEffectHistoryV1 = uint32(100_000)
	MaxBasisPoints         = uint32(10_000)

	DefaultEvidenceTTLBlocks        = uint64(10_000)
	DefaultReviewQuorumBps          = uint32(6_700)
	DefaultMinSlashFractionBps      = uint32(1)
	DefaultMaxSlashFractionBps      = uint32(2_000)
	DefaultCriticalSlashFractionBps = uint32(500)
	DefaultReporterRewardNaet       = uint64(1_000_000)
)

type State struct {
	Evidence             []EvidenceRecord
	SlashEvents          []SlashEvent
	ReporterRewards      []ReporterReward
	RegistryUpdates      []RegistryUpdate
	TombstonedValidators []string
}

type EvidenceRecord struct {
	EvidenceID       string
	Status           string
	EvidenceType     string
	AccusedValidator string
	Reporter         string
	ProofPayloadHash string
	PayloadSizeBytes uint32
	Votes            []EvidenceVote
	SlashDecision    SlashDecision
	RewardDecision   RewardDecision
	SubmittedHeight  uint64
	UpdatedHeight    uint64
	ExpirationHeight uint64
	FinalizedHeight  uint64
	RequiresReview   bool
	RejectionReason  string
}

type EvidenceVote struct {
	Voter          string
	Support        string
	VotingPowerBps uint32
	Height         uint64
}

type SlashDecision struct {
	FractionBps uint32
	Tombstone   bool
	Applied     bool
}

type RewardDecision struct {
	Reporter   string
	AmountNaet uint64
	Paid       bool
}

type SlashEvent struct {
	EvidenceID       string
	ValidatorAddress string
	FractionBps      uint32
	Tombstone        bool
	Height           uint64
}

type ReporterReward struct {
	EvidenceID string
	Reporter   string
	AmountNaet uint64
	Paid       bool
	Height     uint64
}

type RegistryUpdate struct {
	EvidenceID       string
	ValidatorAddress string
	Status           string
	Height           uint64
}

type MsgSubmitEvidence struct {
	Authority        string
	EvidenceID       string
	EvidenceType     string
	AccusedValidator string
	Reporter         string
	ProofPayloadHash string
	PayloadSizeBytes uint32
	RequiresReview   bool
	SlashFractionBps uint32
	RewardNaet       uint64
	Height           uint64
}

type MsgVoteEvidence struct {
	Authority      string
	EvidenceID     string
	Voter          string
	Accept         bool
	VotingPowerBps uint32
	Height         uint64
}

type MsgFinalizeEvidence struct {
	Authority  string
	EvidenceID string
	Height     uint64
}

type MsgCancelExpiredEvidence struct {
	Authority  string
	EvidenceID string
	Height     uint64
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("native evidence update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("native evidence update requires governance authority")
	}
	return nil
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Evidence)) > params.MaxEvidence {
		return errors.New("native evidence total evidence limit exceeded")
	}
	if uint32(len(s.SlashEvents)) > params.MaxSideEffectHistory ||
		uint32(len(s.ReporterRewards)) > params.MaxSideEffectHistory ||
		uint32(len(s.RegistryUpdates)) > params.MaxSideEffectHistory {
		return errors.New("native evidence side effect history limit exceeded")
	}
	ids := map[string]struct{}{}
	hashes := map[string]struct{}{}
	pending := uint32(0)
	tombstoned := map[string]struct{}{}
	for _, validator := range s.TombstonedValidators {
		if err := addressing.ValidateAuthorityAddress("native evidence tombstoned validator", validator); err != nil {
			return err
		}
		if _, found := tombstoned[validator]; found {
			return fmt.Errorf("native evidence duplicate tombstoned validator %s", validator)
		}
		tombstoned[validator] = struct{}{}
	}
	for _, evidence := range s.Evidence {
		if err := evidence.Validate(params); err != nil {
			return err
		}
		if _, found := ids[evidence.EvidenceID]; found {
			return fmt.Errorf("native evidence duplicate evidence id %s", evidence.EvidenceID)
		}
		ids[evidence.EvidenceID] = struct{}{}
		if _, found := hashes[evidence.ProofPayloadHash]; found {
			return fmt.Errorf("native evidence duplicate proof payload hash %s", evidence.ProofPayloadHash)
		}
		hashes[evidence.ProofPayloadHash] = struct{}{}
		if evidence.Status == StatusPending {
			pending++
		}
		if evidence.Status == StatusExpired && evidence.SlashDecision.Applied {
			return errors.New("native evidence expired evidence cannot slash")
		}
		if evidence.SlashDecision.Tombstone {
			if _, found := tombstoned[evidence.AccusedValidator]; evidence.Status == StatusAccepted && !found {
				return errors.New("native evidence tombstoned validator missing irreversible marker")
			}
		}
	}
	if pending > params.MaxPendingEvidence {
		return errors.New("native evidence pending evidence limit exceeded")
	}
	if err := validateSingleSideEffectPerEvidence(s.SlashEvents, func(e SlashEvent) string { return e.EvidenceID }); err != nil {
		return err
	}
	if err := validateSingleSideEffectPerEvidence(s.ReporterRewards, func(r ReporterReward) string { return r.EvidenceID }); err != nil {
		return err
	}
	for _, event := range s.SlashEvents {
		if event.FractionBps < params.MinSlashFractionBps || event.FractionBps > params.MaxSlashFractionBps {
			return errors.New("native evidence slash event fraction outside configured bounds")
		}
		if err := addressing.ValidateAuthorityAddress("native evidence slash validator", event.ValidatorAddress); err != nil {
			return err
		}
	}
	for _, reward := range s.ReporterRewards {
		if reward.AmountNaet > params.MaxReporterRewardNaet {
			return errors.New("native evidence reporter reward exceeds configured limit")
		}
		if err := addressing.ValidateAuthorityAddress("native evidence reward reporter", reward.Reporter); err != nil {
			return err
		}
	}
	for _, update := range s.RegistryUpdates {
		if err := addressing.ValidateAuthorityAddress("native evidence registry validator", update.ValidatorAddress); err != nil {
			return err
		}
		if update.Status != RegistryStatusJailed && update.Status != RegistryStatusTombstoned {
			return fmt.Errorf("native evidence unsupported registry update status %q", update.Status)
		}
	}
	return nil
}

func (e EvidenceRecord) Validate(params Params) error {
	if strings.TrimSpace(e.EvidenceID) == "" {
		return errors.New("native evidence id is required")
	}
	if e.EvidenceID != strings.TrimSpace(e.EvidenceID) || len(e.EvidenceID) > 96 {
		return errors.New("native evidence id must be trimmed and <= 96 bytes")
	}
	if !isStatus(e.Status) {
		return fmt.Errorf("native evidence unsupported status %q", e.Status)
	}
	if !isEvidenceType(e.EvidenceType) {
		return fmt.Errorf("native evidence unsupported evidence type %q", e.EvidenceType)
	}
	if err := addressing.ValidateAuthorityAddress("native evidence accused validator", e.AccusedValidator); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("native evidence reporter", e.Reporter); err != nil {
		return err
	}
	if err := validateProofPayloadHash(e.ProofPayloadHash, params.MaxProofHashBytes); err != nil {
		return err
	}
	if e.PayloadSizeBytes == 0 || e.PayloadSizeBytes > params.MaxPayloadBytes {
		return errors.New("native evidence payload size is outside configured bounds")
	}
	if e.SubmittedHeight == 0 || e.UpdatedHeight == 0 || e.ExpirationHeight <= e.SubmittedHeight {
		return errors.New("native evidence heights are invalid")
	}
	if e.FinalizedHeight > 0 && e.FinalizedHeight < e.SubmittedHeight {
		return errors.New("native evidence finalized height cannot precede submission")
	}
	if uint32(len(e.Votes)) > params.MaxVotes {
		return errors.New("native evidence vote limit exceeded")
	}
	voters := map[string]struct{}{}
	for _, vote := range e.Votes {
		if err := vote.Validate(); err != nil {
			return err
		}
		if _, found := voters[vote.Voter]; found {
			return fmt.Errorf("native evidence duplicate vote from %s", vote.Voter)
		}
		voters[vote.Voter] = struct{}{}
	}
	if err := e.SlashDecision.Validate(params); err != nil {
		return err
	}
	if err := e.RewardDecision.Validate(params, e.Reporter); err != nil {
		return err
	}
	if e.Status == StatusAccepted && !e.SlashDecision.Applied {
		return errors.New("native evidence accepted evidence must have applied slash decision")
	}
	if e.Status == StatusAccepted && !e.RewardDecision.Paid {
		return errors.New("native evidence accepted evidence must have paid reward decision")
	}
	if e.Status != StatusAccepted && (e.SlashDecision.Applied || e.RewardDecision.Paid) {
		return errors.New("native evidence non-accepted evidence cannot have applied slash or paid reward")
	}
	return nil
}

func (v EvidenceVote) Validate() error {
	if err := addressing.ValidateAuthorityAddress("native evidence voter", v.Voter); err != nil {
		return err
	}
	if v.Support != VoteSupportAccept && v.Support != VoteSupportReject {
		return fmt.Errorf("native evidence unsupported vote support %q", v.Support)
	}
	if v.VotingPowerBps == 0 || v.VotingPowerBps > MaxBasisPoints {
		return fmt.Errorf("native evidence voting power must be within 1..%d bps", MaxBasisPoints)
	}
	if v.Height == 0 {
		return errors.New("native evidence vote height must be positive")
	}
	return nil
}

func (d SlashDecision) Validate(params Params) error {
	if d.FractionBps < params.MinSlashFractionBps || d.FractionBps > params.MaxSlashFractionBps {
		return errors.New("native evidence slash fraction outside configured bounds")
	}
	return nil
}

func (d RewardDecision) Validate(params Params, reporter string) error {
	if d.Reporter != reporter {
		return errors.New("native evidence reward reporter must match evidence reporter")
	}
	if d.AmountNaet > params.MaxReporterRewardNaet {
		return errors.New("native evidence reporter reward exceeds configured limit")
	}
	return nil
}

func (s State) Normalize(params Params) State {
	s.Evidence = SortEvidence(s.Evidence)
	s.SlashEvents = SortSlashEvents(s.SlashEvents)
	s.ReporterRewards = SortReporterRewards(s.ReporterRewards)
	s.RegistryUpdates = SortRegistryUpdates(s.RegistryUpdates)
	s.TombstonedValidators = sortedUnique(s.TombstonedValidators)
	for idx := range s.Evidence {
		s.Evidence[idx].Votes = SortVotes(s.Evidence[idx].Votes)
	}
	return s
}

func SortEvidence(values []EvidenceRecord) []EvidenceRecord {
	out := append([]EvidenceRecord(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].EvidenceID < out[j].EvidenceID })
	return out
}

func SortVotes(values []EvidenceVote) []EvidenceVote {
	out := append([]EvidenceVote(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Voter < out[j].Voter })
	return out
}

func SortSlashEvents(values []SlashEvent) []SlashEvent {
	out := append([]SlashEvent(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].EvidenceID < out[j].EvidenceID })
	return out
}

func SortReporterRewards(values []ReporterReward) []ReporterReward {
	out := append([]ReporterReward(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].EvidenceID < out[j].EvidenceID })
	return out
}

func SortRegistryUpdates(values []RegistryUpdate) []RegistryUpdate {
	out := append([]RegistryUpdate(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].EvidenceID != out[j].EvidenceID {
			return out[i].EvidenceID < out[j].EvidenceID
		}
		return out[i].ValidatorAddress < out[j].ValidatorAddress
	})
	return out
}

func CanonicalSlashFraction(params Params, evidenceType string, requested uint32) uint32 {
	if requested != 0 {
		return requested
	}
	switch evidenceType {
	case EvidenceTypeConsensus, EvidenceTypeFraud:
		return params.CriticalFaultSlashFractionBps
	default:
		return params.MinSlashFractionBps
	}
}

func IsCriticalEvidenceType(evidenceType string) bool {
	return evidenceType == EvidenceTypeConsensus || evidenceType == EvidenceTypeFraud
}

func AcceptedVotingPowerBps(votes []EvidenceVote) uint32 {
	total := uint32(0)
	for _, vote := range votes {
		if vote.Support == VoteSupportAccept {
			total += vote.VotingPowerBps
		}
	}
	if total > MaxBasisPoints {
		return MaxBasisPoints
	}
	return total
}

func validateProofPayloadHash(value string, maxBytes uint32) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("native evidence proof payload hash is required")
	}
	if value != strings.TrimSpace(value) || uint32(len(value)) > maxBytes {
		return errors.New("native evidence proof payload hash must be trimmed and within configured length")
	}
	if len(value)%2 != 0 {
		return errors.New("native evidence proof payload hash must be even-length hex")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("native evidence proof payload hash must be hex: %w", err)
	}
	return nil
}

func isStatus(status string) bool {
	switch status {
	case StatusPending, StatusAccepted, StatusRejected, StatusExpired:
		return true
	default:
		return false
	}
}

func isEvidenceType(evidenceType string) bool {
	switch evidenceType {
	case EvidenceTypeConsensus, EvidenceTypeMissedBlock, EvidenceTypePerformance, EvidenceTypeFraud:
		return true
	default:
		return false
	}
}

func validateSingleSideEffectPerEvidence[T any](values []T, id func(T) string) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		evidenceID := id(value)
		if _, found := seen[evidenceID]; found {
			return fmt.Errorf("native evidence duplicate side effect for %s", evidenceID)
		}
		seen[evidenceID] = struct{}{}
	}
	return nil
}

func sortedUnique(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
