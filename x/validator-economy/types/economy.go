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
	ModuleName          = "validator_economy"
	DefaultScoreVersion = uint32(1)

	SaturationStatusNone      = "none"
	SaturationStatusSaturated = "saturated"
)

type ValidatorScoreRecord struct {
	EpochID           uint64
	ValidatorAddress  string
	RawStake          sdkmath.Int
	EffectiveStake    sdkmath.Int
	StakeWeight       sdkmath.Int
	PerformanceFactor uint32
	UptimeFactor      uint32
	LatencyFactor     uint32
	ReliabilityIndex  uint32
	ValidatorScore    sdkmath.Int
	SaturationStatus  string
	ScoreVersion      uint32
}

type ScoreComponentState struct {
	Records []ValidatorScoreRecord
}

type ElectionRanking struct {
	EpochID                 uint64
	Records                 []ValidatorScoreRecord
	Rejected                []RejectedScoreCandidate
	MaxValidatorSetChanges  uint32
	TransitionLimited       bool
	RequestedValidatorCount uint32
}

type RejectedScoreCandidate struct {
	ValidatorAddress string
	Reason           string
}

type ScoreSimulationInput struct {
	EpochID        uint64
	Params         postypes.Params
	Candidates     []postypes.Candidate
	PreviousActive []string
	TargetActive   uint32
}

type ScoreSimulationResult struct {
	Ranking               ElectionRanking
	ActiveValidatorIDs    []string
	TotalRawStakeNaet     sdkmath.Int
	TotalEffectiveNaet    sdkmath.Int
	MaxRawStakeShareBps   uint32
	MaxEffectiveShareBps  uint32
	CentralizationWarning bool
}

func BuildValidatorScoreRecord(epochID uint64, params postypes.Params, candidate postypes.Candidate) (ValidatorScoreRecord, error) {
	scored, err := postypes.ScoreCandidate(params, candidate)
	if err != nil {
		return ValidatorScoreRecord{}, err
	}
	status := SaturationStatusNone
	if scored.ScoreComponents.SaturatedStakeNaet.IsPositive() {
		status = SaturationStatusSaturated
	}
	record := ValidatorScoreRecord{
		EpochID:           epochID,
		ValidatorAddress:  strings.TrimSpace(scored.ValidatorID),
		RawStake:          scored.TotalStakeNaet,
		EffectiveStake:    scored.EffectiveStakeNaet,
		StakeWeight:       scored.ScoreComponents.StakeWeightNaet,
		PerformanceFactor: scored.ScoreComponents.PerformanceFactorBps,
		UptimeFactor:      scored.ScoreComponents.UptimeFactorBps,
		LatencyFactor:     scored.ScoreComponents.LatencyFactorBps,
		ReliabilityIndex:  scored.ScoreComponents.ReliabilityIndexBps,
		ValidatorScore:    scored.Score,
		SaturationStatus:  status,
		ScoreVersion:      DefaultScoreVersion,
	}
	return record, record.Validate()
}

func (r ValidatorScoreRecord) Validate() error {
	if r.EpochID == 0 {
		return errors.New("score record epoch id is required")
	}
	if strings.TrimSpace(r.ValidatorAddress) == "" {
		return errors.New("score record validator address is required")
	}
	if r.RawStake.IsNegative() {
		return errors.New("score record raw stake cannot be negative")
	}
	if r.EffectiveStake.IsNegative() {
		return errors.New("score record effective stake cannot be negative")
	}
	if r.StakeWeight.IsNegative() {
		return errors.New("score record stake weight cannot be negative")
	}
	if r.EffectiveStake.GT(r.RawStake) {
		return errors.New("score record effective stake cannot exceed raw stake")
	}
	if !r.EffectiveStake.Equal(r.StakeWeight) {
		return errors.New("score record stake weight must equal effective stake")
	}
	if r.PerformanceFactor > postypes.BasisPoints {
		return fmt.Errorf("performance factor must be <= %d bps", postypes.BasisPoints)
	}
	if r.UptimeFactor > postypes.BasisPoints {
		return fmt.Errorf("uptime factor must be <= %d bps", postypes.BasisPoints)
	}
	if r.LatencyFactor > postypes.BasisPoints {
		return fmt.Errorf("latency factor must be <= %d bps", postypes.BasisPoints)
	}
	if r.ReliabilityIndex > postypes.BasisPoints {
		return fmt.Errorf("reliability index must be <= %d bps", postypes.BasisPoints)
	}
	if r.ValidatorScore.IsNegative() {
		return errors.New("validator score cannot be negative")
	}
	if r.SaturationStatus != SaturationStatusNone && r.SaturationStatus != SaturationStatusSaturated {
		return fmt.Errorf("unsupported saturation status %q", r.SaturationStatus)
	}
	if r.ScoreVersion == 0 {
		return errors.New("score version is required")
	}
	return nil
}

func NewScoreComponentState(records []ValidatorScoreRecord) (ScoreComponentState, error) {
	out := make([]ValidatorScoreRecord, len(records))
	seen := make(map[string]struct{}, len(records))
	for i, record := range records {
		record.ValidatorAddress = strings.TrimSpace(record.ValidatorAddress)
		if err := record.Validate(); err != nil {
			return ScoreComponentState{}, err
		}
		key := scoreRecordKey(record.EpochID, record.ValidatorAddress)
		if _, found := seen[key]; found {
			return ScoreComponentState{}, fmt.Errorf("duplicate score record %s", key)
		}
		seen[key] = struct{}{}
		out[i] = record
	}
	sortScoreRecords(out)
	return ScoreComponentState{Records: out}, nil
}

func (s ScoreComponentState) GetScoreRecord(epochID uint64, validatorAddress string) (ValidatorScoreRecord, bool) {
	validatorAddress = strings.TrimSpace(validatorAddress)
	for _, record := range s.Records {
		if record.EpochID == epochID && record.ValidatorAddress == validatorAddress {
			return record, true
		}
	}
	return ValidatorScoreRecord{}, false
}

func (s ScoreComponentState) RecordsForEpoch(epochID uint64) []ValidatorScoreRecord {
	records := make([]ValidatorScoreRecord, 0)
	for _, record := range s.Records {
		if record.EpochID == epochID {
			records = append(records, record)
		}
	}
	sortScoreRecords(records)
	return records
}

func BuildElectionRanking(epochID uint64, params postypes.Params, candidates []postypes.Candidate, targetActive uint32) (ElectionRanking, error) {
	if err := params.Validate(); err != nil {
		return ElectionRanking{}, err
	}
	if epochID == 0 {
		return ElectionRanking{}, errors.New("ranking epoch id is required")
	}
	if targetActive == 0 {
		targetActive = params.MinActiveValidators
	}
	ranking := ElectionRanking{EpochID: epochID, RequestedValidatorCount: targetActive}
	records := make([]ValidatorScoreRecord, 0, len(candidates))
	for _, candidate := range candidates {
		record, err := BuildValidatorScoreRecord(epochID, params, candidate)
		if err != nil {
			ranking.Rejected = append(ranking.Rejected, RejectedScoreCandidate{
				ValidatorAddress: strings.TrimSpace(candidate.ValidatorID),
				Reason:           err.Error(),
			})
			continue
		}
		records = append(records, record)
	}
	sortScoreRecords(records)
	limit := minUint32(targetActive, uint32(len(records)))
	ranking.Records = records[:limit]
	return ranking, nil
}

func ApplyValidatorSetTransitionLimit(params postypes.Params, previousActive []string, ranking ElectionRanking) (ElectionRanking, error) {
	if len(previousActive) == 0 {
		return ranking, nil
	}
	maxChanges, err := MaxValidatorSetChanges(params, uint32(len(previousActive)))
	if err != nil {
		return ElectionRanking{}, err
	}
	if maxChanges == 0 {
		ranking.MaxValidatorSetChanges = maxChanges
		return ranking, nil
	}
	previous := make(map[string]struct{}, len(previousActive))
	for _, validatorID := range previousActive {
		validatorID = strings.TrimSpace(validatorID)
		if validatorID != "" {
			previous[validatorID] = struct{}{}
		}
	}
	changes := uint32(0)
	limited := make([]ValidatorScoreRecord, 0, len(ranking.Records))
	deferred := make([]ValidatorScoreRecord, 0)
	for _, record := range ranking.Records {
		if _, wasActive := previous[record.ValidatorAddress]; wasActive || changes < maxChanges {
			limited = append(limited, record)
			if !wasActive {
				changes++
			}
			continue
		}
		deferred = append(deferred, record)
	}
	if len(deferred) > 0 {
		ranking.TransitionLimited = true
	}
	ranking.MaxValidatorSetChanges = maxChanges
	ranking.Records = limited
	return ranking, nil
}

func MaxValidatorSetChanges(params postypes.Params, activeValidatorCount uint32) (uint32, error) {
	return postypes.MaxValidatorSetChanges(params, activeValidatorCount)
}

func SimulateScores(input ScoreSimulationInput) (ScoreSimulationResult, error) {
	ranking, err := BuildElectionRanking(input.EpochID, input.Params, input.Candidates, input.TargetActive)
	if err != nil {
		return ScoreSimulationResult{}, err
	}
	ranking, err = ApplyValidatorSetTransitionLimit(input.Params, input.PreviousActive, ranking)
	if err != nil {
		return ScoreSimulationResult{}, err
	}
	result := ScoreSimulationResult{
		Ranking:            ranking,
		TotalRawStakeNaet:  sdkmath.ZeroInt(),
		TotalEffectiveNaet: sdkmath.ZeroInt(),
	}
	maxRaw := sdkmath.ZeroInt()
	maxEffective := sdkmath.ZeroInt()
	for _, record := range ranking.Records {
		result.ActiveValidatorIDs = append(result.ActiveValidatorIDs, record.ValidatorAddress)
		result.TotalRawStakeNaet = result.TotalRawStakeNaet.Add(record.RawStake)
		result.TotalEffectiveNaet = result.TotalEffectiveNaet.Add(record.EffectiveStake)
		if record.RawStake.GT(maxRaw) {
			maxRaw = record.RawStake
		}
		if record.EffectiveStake.GT(maxEffective) {
			maxEffective = record.EffectiveStake
		}
	}
	result.MaxRawStakeShareBps = shareBps(maxRaw, result.TotalRawStakeNaet)
	result.MaxEffectiveShareBps = shareBps(maxEffective, result.TotalEffectiveNaet)
	result.CentralizationWarning = result.MaxEffectiveShareBps > input.Params.MaxVotingPowerBps
	return result, nil
}

func sortScoreRecords(records []ValidatorScoreRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		left := records[i]
		right := records[j]
		if !left.ValidatorScore.Equal(right.ValidatorScore) {
			return left.ValidatorScore.GT(right.ValidatorScore)
		}
		if !left.EffectiveStake.Equal(right.EffectiveStake) {
			return left.EffectiveStake.GT(right.EffectiveStake)
		}
		if left.EpochID != right.EpochID {
			return left.EpochID < right.EpochID
		}
		return left.ValidatorAddress < right.ValidatorAddress
	})
}

func scoreRecordKey(epochID uint64, validatorAddress string) string {
	return fmt.Sprintf("%d/%s", epochID, validatorAddress)
}

func shareBps(part sdkmath.Int, total sdkmath.Int) uint32 {
	if !part.IsPositive() || !total.IsPositive() {
		return 0
	}
	if part.GTE(total) {
		return postypes.BasisPoints
	}
	return uint32(part.MulRaw(int64(postypes.BasisPoints)).Quo(total).Uint64())
}

func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
