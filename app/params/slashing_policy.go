package params

import (
	"fmt"
	"time"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

const (
	SlashingEvidenceStandardCosmos = "cosmos_sdk_x_slashing_x_evidence"

	DoubleSignSlashMinBps     = int64(500)
	DoubleSignSlashMaxBps     = int64(1_000)
	DoubleSignSlashDefaultBps = DoubleSignSlashMinBps

	DowntimeFirstSlashMinBps     = int64(5)
	DowntimeFirstSlashMaxBps     = int64(10)
	DowntimeFirstSlashDefaultBps = DowntimeFirstSlashMinBps

	DowntimeRepeatSlashMinBps     = int64(25)
	DowntimeRepeatSlashMaxBps     = int64(50)
	DowntimeRepeatSlashDefaultBps = DowntimeRepeatSlashMinBps

	DowntimeChronicSlashMaxBps     = int64(100)
	DowntimeChronicSlashDefaultBps = DowntimeChronicSlashMaxBps

	DowntimeFirstJailMinMinutes     = int64(60)
	DowntimeFirstJailMaxMinutes     = int64(360)
	DowntimeFirstJailDefaultMinutes = DowntimeFirstJailMinMinutes
	DowntimeRepeatJailMinutes       = int64(24 * 60)
	DowntimeChronicJailMinutes      = int64(72 * 60)
)

type SlashingAccountabilityPolicy struct {
	EvidenceStandard                   string
	ObjectiveCryptographicEvidenceOnly bool
	SubjectiveSlashingAllowed          bool
	DoubleSignSlashBps                 int64
	DoubleSignJailImmediate            bool
	DoubleSignPermanentTombstone       bool
	ConsensusKeyReuseForbidden         bool
	UsesCosmosSlashingAndEvidence      bool
	ProgressiveDowntimeEnabled         bool
	StandardDowntimeStatePreserved     bool
	CustomDowntimeOverlayRequired      bool
	DowntimeFirstSlashBps              int64
	DowntimeFirstJailMinutes           int64
	DowntimeRepeatSlashBps             int64
	DowntimeRepeatJailMinutes          int64
	DowntimeChronicSlashBps            int64
	DowntimeChronicJailMinutes         int64
	DowntimeGovernanceReputationFlag   bool
}

func DefaultSlashingAccountabilityPolicy() SlashingAccountabilityPolicy {
	return SlashingAccountabilityPolicy{
		EvidenceStandard:                   SlashingEvidenceStandardCosmos,
		ObjectiveCryptographicEvidenceOnly: true,
		SubjectiveSlashingAllowed:          false,
		DoubleSignSlashBps:                 DoubleSignSlashDefaultBps,
		DoubleSignJailImmediate:            true,
		DoubleSignPermanentTombstone:       true,
		ConsensusKeyReuseForbidden:         true,
		UsesCosmosSlashingAndEvidence:      true,
		ProgressiveDowntimeEnabled:         true,
		StandardDowntimeStatePreserved:     true,
		CustomDowntimeOverlayRequired:      true,
		DowntimeFirstSlashBps:              DowntimeFirstSlashDefaultBps,
		DowntimeFirstJailMinutes:           DowntimeFirstJailDefaultMinutes,
		DowntimeRepeatSlashBps:             DowntimeRepeatSlashDefaultBps,
		DowntimeRepeatJailMinutes:          DowntimeRepeatJailMinutes,
		DowntimeChronicSlashBps:            DowntimeChronicSlashDefaultBps,
		DowntimeChronicJailMinutes:         DowntimeChronicJailMinutes,
		DowntimeGovernanceReputationFlag:   true,
	}
}

func AetraSlashingParams() slashingtypes.Params {
	params := slashingtypes.DefaultParams()
	params.SlashFractionDoubleSign = BpsToLegacyDec(DoubleSignSlashDefaultBps)
	params.SlashFractionDowntime = BpsToLegacyDec(DowntimeFirstSlashDefaultBps)
	params.DowntimeJailDuration = time.Duration(DowntimeFirstJailDefaultMinutes) * time.Minute
	return params
}

func (p SlashingAccountabilityPolicy) Validate() error {
	if p.EvidenceStandard != SlashingEvidenceStandardCosmos {
		return fmt.Errorf("slashing evidence standard must use Cosmos SDK x/slashing and x/evidence")
	}
	if !p.ObjectiveCryptographicEvidenceOnly {
		return fmt.Errorf("slashing must require objective cryptographic evidence")
	}
	if p.SubjectiveSlashingAllowed {
		return fmt.Errorf("subjective slashing must not be enabled")
	}
	if err := validateSlashingBpsValue("double_sign_slash", p.DoubleSignSlashBps, DoubleSignSlashMinBps, DoubleSignSlashMaxBps); err != nil {
		return err
	}
	if !p.DoubleSignJailImmediate {
		return fmt.Errorf("double-sign evidence must jail immediately")
	}
	if !p.DoubleSignPermanentTombstone {
		return fmt.Errorf("double-sign evidence must permanently tombstone the validator")
	}
	if !p.ConsensusKeyReuseForbidden {
		return fmt.Errorf("double-sign tombstone must forbid consensus key reuse")
	}
	if !p.UsesCosmosSlashingAndEvidence {
		return fmt.Errorf("slashing policy must use Cosmos SDK slashing and evidence modules")
	}
	if !p.ProgressiveDowntimeEnabled {
		return fmt.Errorf("progressive downtime penalties must be enabled")
	}
	if !p.StandardDowntimeStatePreserved {
		return fmt.Errorf("progressive downtime must preserve standard x/slashing signing state")
	}
	if !p.CustomDowntimeOverlayRequired {
		return fmt.Errorf("progressive downtime requires custom overlay when x/slashing is insufficient")
	}
	if err := validateSlashingBpsValue("downtime_first_slash", p.DowntimeFirstSlashBps, DowntimeFirstSlashMinBps, DowntimeFirstSlashMaxBps); err != nil {
		return err
	}
	if p.DowntimeFirstJailMinutes < DowntimeFirstJailMinMinutes || p.DowntimeFirstJailMinutes > DowntimeFirstJailMaxMinutes {
		return fmt.Errorf("downtime first jail must stay within 1-6 hours")
	}
	if err := validateSlashingBpsValue("downtime_repeat_slash", p.DowntimeRepeatSlashBps, DowntimeRepeatSlashMinBps, DowntimeRepeatSlashMaxBps); err != nil {
		return err
	}
	if p.DowntimeRepeatJailMinutes != DowntimeRepeatJailMinutes {
		return fmt.Errorf("downtime repeat jail must be 24 hours")
	}
	if p.DowntimeChronicSlashBps <= p.DowntimeRepeatSlashBps || p.DowntimeChronicSlashBps > DowntimeChronicSlashMaxBps {
		return fmt.Errorf("downtime chronic slash must be above repeat slash and <= 1 percent")
	}
	if p.DowntimeChronicJailMinutes <= p.DowntimeRepeatJailMinutes {
		return fmt.Errorf("downtime chronic jail must be longer than repeat jail")
	}
	if !p.DowntimeGovernanceReputationFlag {
		return fmt.Errorf("chronic downtime must expose governance or reputation flag")
	}
	return nil
}

func validateSlashingBpsValue(name string, value, allowedMin, allowedMax int64) error {
	if value < allowedMin || value > allowedMax {
		return fmt.Errorf("%s must stay within %d-%d bps", name, allowedMin, allowedMax)
	}
	return nil
}
