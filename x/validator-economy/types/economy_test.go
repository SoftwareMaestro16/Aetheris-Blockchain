package types

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestBuildValidatorScoreRecordCapturesDeterministicComponents(t *testing.T) {
	params := testParams()
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	params.StakeSaturationCapFactorBps = 20_000
	params.StakeSaturationNaet = sdkmath.NewInt(10_000)
	candidate := testCandidate("val-record", 5_000)
	candidate.PerformanceScoreBps = 9_000
	candidate.UptimeFactorBps = 9_500
	candidate.LatencyFactorBps = 8_000
	candidate.ReliabilityIndexBps = 7_000

	record, err := BuildValidatorScoreRecord(12, params, candidate)
	require.NoError(t, err)
	require.Equal(t, uint64(12), record.EpochID)
	require.Equal(t, "val-record", record.ValidatorAddress)
	require.Equal(t, sdkmath.NewInt(5_000), record.RawStake)
	require.Equal(t, sdkmath.NewInt(2_000), record.EffectiveStake)
	require.Equal(t, sdkmath.NewInt(2_000), record.StakeWeight)
	require.Equal(t, uint32(9_000), record.PerformanceFactor)
	require.Equal(t, uint32(9_500), record.UptimeFactor)
	require.Equal(t, uint32(8_000), record.LatencyFactor)
	require.Equal(t, uint32(7_000), record.ReliabilityIndex)
	require.Equal(t, sdkmath.NewInt(957), record.ValidatorScore)
	require.Equal(t, SaturationStatusSaturated, record.SaturationStatus)
	require.Equal(t, DefaultScoreVersion, record.ScoreVersion)
}

func TestScoreComponentStateQueriesHistoricalRecords(t *testing.T) {
	params := testParams()
	a, err := BuildValidatorScoreRecord(3, params, testCandidate("val-b", 2_000))
	require.NoError(t, err)
	b, err := BuildValidatorScoreRecord(2, params, testCandidate("val-a", 1_000))
	require.NoError(t, err)

	state, err := NewScoreComponentState([]ValidatorScoreRecord{a, b})
	require.NoError(t, err)
	found, ok := state.GetScoreRecord(2, "val-a")
	require.True(t, ok)
	require.Equal(t, b, found)
	require.Equal(t, []ValidatorScoreRecord{b}, state.RecordsForEpoch(2))

	_, err = NewScoreComponentState([]ValidatorScoreRecord{b, b})
	require.ErrorContains(t, err, "duplicate score record")
}

func TestElectionRankingOrdersByScoreAndReportsRejectedCandidates(t *testing.T) {
	params := testParams()
	candidates := []postypes.Candidate{
		testCandidate("val-low", 1_000),
		testCandidate("val-high", 3_000),
		testCandidate("val-jailed", 9_000),
	}
	candidates[2].Jailed = true

	ranking, err := BuildElectionRanking(5, params, candidates, 2)
	require.NoError(t, err)
	require.Equal(t, []string{"val-high", "val-low"}, recordIDs(ranking.Records))
	require.Len(t, ranking.Rejected, 1)
	require.Equal(t, "val-jailed", ranking.Rejected[0].ValidatorAddress)
	require.Contains(t, ranking.Rejected[0].Reason, "jailed")
}

func TestValidatorSetTransitionLimitDefersExcessNewValidators(t *testing.T) {
	params := testParams()
	params.MaxValidatorSetChangeRateBps = 100
	previous := make([]string, 75)
	for i := range previous {
		previous[i] = fmt.Sprintf("old-%03d", i)
	}
	records := []ValidatorScoreRecord{
		testRecord(7, "new-a", 5_000),
		testRecord(7, "new-b", 4_000),
		testRecord(7, "new-c", 3_000),
	}
	ranking := ElectionRanking{EpochID: 7, Records: records, RequestedValidatorCount: 3}

	limited, err := ApplyValidatorSetTransitionLimit(params, previous, ranking)
	require.NoError(t, err)
	require.Equal(t, uint32(1), limited.MaxValidatorSetChanges)
	require.True(t, limited.TransitionLimited)
	require.Equal(t, []string{"new-a"}, recordIDs(limited.Records))
}

func TestScoreSimulationFlagsCentralizationAfterSaturation(t *testing.T) {
	params := testParams()
	params.MaxVotingPowerBps = 3_000
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	params.StakeSaturationCapFactorBps = 2_0000
	candidates := []postypes.Candidate{
		testCandidate("val-whale", 9_000),
		testCandidate("val-a", 1_000),
		testCandidate("val-b", 1_000),
		testCandidate("val-c", 1_000),
	}

	result, err := SimulateScores(ScoreSimulationInput{
		EpochID:      8,
		Params:       params,
		Candidates:   candidates,
		TargetActive: 4,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(7_500), result.MaxRawStakeShareBps)
	require.Equal(t, uint32(4_000), result.MaxEffectiveShareBps)
	require.True(t, result.CentralizationWarning)
	require.Equal(t, SaturationStatusSaturated, result.Ranking.Records[0].SaturationStatus)
}

func TestStakeSplittingImprovesEffectiveWeightOnlyThroughDistribution(t *testing.T) {
	params := testParams()
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	params.StakeSaturationCapFactorBps = 30_000
	single, err := SimulateScores(ScoreSimulationInput{
		EpochID:      9,
		Params:       params,
		Candidates:   []postypes.Candidate{testCandidate("val-whale", 9_000)},
		TargetActive: 1,
	})
	require.NoError(t, err)
	split, err := SimulateScores(ScoreSimulationInput{
		EpochID: 9,
		Params:  params,
		Candidates: []postypes.Candidate{
			testCandidate("val-a", 3_000),
			testCandidate("val-b", 3_000),
			testCandidate("val-c", 3_000),
		},
		TargetActive: 3,
	})
	require.NoError(t, err)

	require.Equal(t, sdkmath.NewInt(9_000), single.TotalRawStakeNaet)
	require.Equal(t, sdkmath.NewInt(3_000), single.TotalEffectiveNaet)
	require.Equal(t, sdkmath.NewInt(9_000), split.TotalRawStakeNaet)
	require.Equal(t, sdkmath.NewInt(9_000), split.TotalEffectiveNaet)
	require.Equal(t, []string{"val-a", "val-b", "val-c"}, split.ActiveValidatorIDs)
}

func testParams() postypes.Params {
	params := postypes.DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.StakeSaturationNaet = sdkmath.NewInt(100_000)
	return params
}

func testCandidate(id string, stake int64) postypes.Candidate {
	return postypes.Candidate{
		ValidatorID:         id,
		SelfStakeNaet:       sdkmath.NewInt(stake),
		DelegatedStakeNaet:  sdkmath.ZeroInt(),
		PerformanceScoreBps: postypes.BasisPoints,
		UptimeFactorBps:     postypes.BasisPoints,
		CommissionBps:       500,
	}
}

func testRecord(epochID uint64, id string, score int64) ValidatorScoreRecord {
	return ValidatorScoreRecord{
		EpochID:           epochID,
		ValidatorAddress:  id,
		RawStake:          sdkmath.NewInt(score),
		EffectiveStake:    sdkmath.NewInt(score),
		StakeWeight:       sdkmath.NewInt(score),
		PerformanceFactor: postypes.BasisPoints,
		UptimeFactor:      postypes.BasisPoints,
		LatencyFactor:     postypes.BasisPoints,
		ReliabilityIndex:  postypes.BasisPoints,
		ValidatorScore:    sdkmath.NewInt(score),
		SaturationStatus:  SaturationStatusNone,
		ScoreVersion:      DefaultScoreVersion,
	}
}

func recordIDs(records []ValidatorScoreRecord) []string {
	ids := make([]string, len(records))
	for i, record := range records {
		ids[i] = record.ValidatorAddress
	}
	return ids
}
