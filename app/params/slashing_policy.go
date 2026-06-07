package params

import "fmt"

const (
	SlashingEvidenceStandardCosmos = "cosmos_sdk_x_slashing_x_evidence"

	DoubleSignSlashMinBps     = int64(500)
	DoubleSignSlashMaxBps     = int64(1_000)
	DoubleSignSlashDefaultBps = DoubleSignSlashMinBps
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
	}
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
	if err := validateBpsRange("double_sign_slash", p.DoubleSignSlashBps, DoubleSignSlashMinBps, DoubleSignSlashMinBps, DoubleSignSlashMaxBps); err != nil {
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
	return nil
}
