package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	ScoreMin = uint8(0)
	ScoreMax = uint8(100)

	LevelRestricted = "restricted"
	LevelNew        = "new"
	LevelNormal     = "normal"
	LevelTrusted    = "trusted"
	LevelElite      = "elite"

	MaxDomainScore   = uint16(10)
	MaxContractScore = uint16(15)
)

type ReputationRecord struct {
	Account          sdk.AccAddress
	Score            uint8
	AgeScore         uint16
	StakingScore     uint16
	TxSuccessScore   uint16
	VolumeScore      uint16
	DomainScore      uint16
	ContractScore    uint16
	SpamPenalty      uint16
	FailedTxPenalty  uint16
	SlashPenalty     uint16
	LastUpdatedEpoch uint64
}

type DecayParams struct {
	InactiveAfterEpochs uint64
	DecayRatePerEpoch   uint8
}

type ProgressiveLimits struct {
	MaxTxsPerBlock uint32
	MaxTxGas       uint64
	MaxQueueMsgs   uint32
}

func ComputeScore(record ReputationRecord) uint8 {
	positive := record.AgeScore +
		record.StakingScore +
		record.TxSuccessScore +
		record.VolumeScore +
		minU16(record.DomainScore, MaxDomainScore) +
		minU16(record.ContractScore, MaxContractScore)
	negative := record.SpamPenalty + record.FailedTxPenalty + record.SlashPenalty
	if negative >= positive {
		return ScoreMin
	}
	net := positive - negative
	if net > uint16(ScoreMax) {
		return ScoreMax
	}
	return uint8(net)
}

func ApplyComputedScore(record ReputationRecord) ReputationRecord {
	record.Score = ComputeScore(record)
	return record
}

func ApplyInactivityDecay(score uint8, inactiveEpochs uint64, params DecayParams) uint8 {
	if inactiveEpochs <= params.InactiveAfterEpochs || params.DecayRatePerEpoch == 0 {
		return score
	}
	decayEpochs := inactiveEpochs - params.InactiveAfterEpochs
	decay := decayEpochs * uint64(params.DecayRatePerEpoch)
	if decay >= uint64(score) {
		return ScoreMin
	}
	return uint8(uint64(score) - decay)
}

func LevelForScore(score uint8) string {
	switch {
	case score < 20:
		return LevelRestricted
	case score < 50:
		return LevelNew
	case score < 80:
		return LevelNormal
	case score < 95:
		return LevelTrusted
	default:
		return LevelElite
	}
}

func ValidateReputationRecord(record ReputationRecord) error {
	if len(record.Account) == 0 {
		return errors.New("reputation account is required")
	}
	if err := addressing.RejectZeroAddress("reputation account", record.Account); err != nil {
		return err
	}
	expected := ComputeScore(record)
	if record.Score != expected {
		return fmt.Errorf("reputation score mismatch: expected %d got %d", expected, record.Score)
	}
	if record.DomainScore > MaxDomainScore {
		return fmt.Errorf("domain score must not exceed %d", MaxDomainScore)
	}
	if record.ContractScore > MaxContractScore {
		return fmt.Errorf("contract score must not exceed %d", MaxContractScore)
	}
	return nil
}

func LimitsForScore(score uint8) ProgressiveLimits {
	switch LevelForScore(score) {
	case LevelRestricted:
		return ProgressiveLimits{MaxTxsPerBlock: 1, MaxTxGas: 100_000, MaxQueueMsgs: 1}
	case LevelNew:
		return ProgressiveLimits{MaxTxsPerBlock: 5, MaxTxGas: 250_000, MaxQueueMsgs: 4}
	case LevelNormal:
		return ProgressiveLimits{MaxTxsPerBlock: 25, MaxTxGas: 1_000_000, MaxQueueMsgs: 16}
	case LevelTrusted:
		return ProgressiveLimits{MaxTxsPerBlock: 100, MaxTxGas: 2_000_000, MaxQueueMsgs: 64}
	default:
		return ProgressiveLimits{MaxTxsPerBlock: 250, MaxTxGas: 5_000_000, MaxQueueMsgs: 128}
	}
}

func IsDirectReputationPurchaseAllowed() bool {
	return false
}

func minU16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}
