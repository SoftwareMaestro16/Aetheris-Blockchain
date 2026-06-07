package params

import (
	"fmt"
	"sort"
)

const (
	AetraValidatorScoreModuleName = "x/aetra-validator-score"

	AetraValidatorScorePurposePublicAccountability = "public_accountability_without_subjective_consensus_control"

	AetraValidatorScoreResponsibilityTrackUptime                  = "track_validator_uptime"
	AetraValidatorScoreResponsibilityTrackMissedBlockWindows      = "track_missed_block_windows"
	AetraValidatorScoreResponsibilityTrackJailHistory             = "track_jail_history"
	AetraValidatorScoreResponsibilityTrackSlashingHistory         = "track_slashing_history"
	AetraValidatorScoreResponsibilityTrackCommissionBehavior      = "track_commission_behavior"
	AetraValidatorScoreResponsibilityTrackSelfBondRatio           = "track_self_bond_ratio"
	AetraValidatorScoreResponsibilityTrackGovernanceParticipation = "track_governance_participation"
	AetraValidatorScoreResponsibilityTrackConcentrationStatus     = "track_concentration_status"
	AetraValidatorScoreResponsibilityProducePublicScore           = "produce_public_score"
	AetraValidatorScoreResponsibilityExplorerFriendlyQueries      = "expose_explorer_friendly_queries"

	AetraValidatorScoreGuardNoSubjectiveCensorship       = "score_must_not_be_subjective_censorship_mechanism"
	AetraValidatorScoreGuardInformationalFirst           = "score_informational_first"
	AetraValidatorScoreGuardObjectiveRewardOnly          = "reward_affecting_only_from_objective_chain_data"
	AetraValidatorScoreGuardConsensusOverrideDisabled    = "consensus_override_disabled_by_default"
	AetraValidatorScoreGuardObjectiveInputsDeterministic = "objective_inputs_must_be_deterministic"
)

type AetraValidatorScoreSpecEvidence struct {
	ModuleName string

	PublicAccountabilityWithoutSubjectiveConsensusControl bool
}

type AetraValidatorScoreSpecReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

type AetraValidatorScoreResponsibilitiesEvidence struct {
	ModuleName string

	TracksValidatorUptime          bool
	TracksMissedBlockWindows       bool
	TracksJailHistory              bool
	TracksSlashingHistory          bool
	TracksCommissionBehavior       bool
	TracksSelfBondRatio            bool
	TracksGovernanceParticipation  bool
	TracksConcentrationStatus      bool
	ProducesPublicScore            bool
	ExposesExplorerFriendlyQueries bool
}

type AetraValidatorScoreResponsibilitiesReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

type AetraValidatorScoreSubjectiveControlEvidence struct {
	ModuleName string

	NoSubjectiveCensorshipMechanism  bool
	InformationalFirst               bool
	RewardAffectingOnlyObjectiveData bool
	ConsensusOverrideDisabledDefault bool
	ObjectiveInputsDeterministic     bool
}

type AetraValidatorScoreSubjectiveControlReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

func DefaultAetraValidatorScoreSpecEvidence() AetraValidatorScoreSpecEvidence {
	return AetraValidatorScoreSpecEvidence{
		ModuleName: AetraValidatorScoreModuleName,

		PublicAccountabilityWithoutSubjectiveConsensusControl: true,
	}
}

func ValidateAetraValidatorScoreSpec(evidence AetraValidatorScoreSpecEvidence) error {
	report := BuildAetraValidatorScoreSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator score spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorScoreSpecReport(evidence AetraValidatorScoreSpecEvidence) AetraValidatorScoreSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraValidatorScoreModuleName {
		failed = append(failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	}

	checks := []requirementCheck{
		{AetraValidatorScorePurposePublicAccountability, evidence.PublicAccountabilityWithoutSubjectiveConsensusControl},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorScoreSpecReport{
		ModuleName: evidence.ModuleName,
		Required:   len(checks),
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}

func DefaultAetraValidatorScoreResponsibilitiesEvidence() AetraValidatorScoreResponsibilitiesEvidence {
	return AetraValidatorScoreResponsibilitiesEvidence{
		ModuleName: AetraValidatorScoreModuleName,

		TracksValidatorUptime:          true,
		TracksMissedBlockWindows:       true,
		TracksJailHistory:              true,
		TracksSlashingHistory:          true,
		TracksCommissionBehavior:       true,
		TracksSelfBondRatio:            true,
		TracksGovernanceParticipation:  true,
		TracksConcentrationStatus:      true,
		ProducesPublicScore:            true,
		ExposesExplorerFriendlyQueries: true,
	}
}

func ValidateAetraValidatorScoreResponsibilities(evidence AetraValidatorScoreResponsibilitiesEvidence) error {
	report := BuildAetraValidatorScoreResponsibilitiesReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator score responsibilities failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorScoreResponsibilitiesReport(evidence AetraValidatorScoreResponsibilitiesEvidence) AetraValidatorScoreResponsibilitiesReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraValidatorScoreModuleName {
		failed = append(failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	}

	checks := []requirementCheck{
		{AetraValidatorScoreResponsibilityTrackUptime, evidence.TracksValidatorUptime},
		{AetraValidatorScoreResponsibilityTrackMissedBlockWindows, evidence.TracksMissedBlockWindows},
		{AetraValidatorScoreResponsibilityTrackJailHistory, evidence.TracksJailHistory},
		{AetraValidatorScoreResponsibilityTrackSlashingHistory, evidence.TracksSlashingHistory},
		{AetraValidatorScoreResponsibilityTrackCommissionBehavior, evidence.TracksCommissionBehavior},
		{AetraValidatorScoreResponsibilityTrackSelfBondRatio, evidence.TracksSelfBondRatio},
		{AetraValidatorScoreResponsibilityTrackGovernanceParticipation, evidence.TracksGovernanceParticipation},
		{AetraValidatorScoreResponsibilityTrackConcentrationStatus, evidence.TracksConcentrationStatus},
		{AetraValidatorScoreResponsibilityProducePublicScore, evidence.ProducesPublicScore},
		{AetraValidatorScoreResponsibilityExplorerFriendlyQueries, evidence.ExposesExplorerFriendlyQueries},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorScoreResponsibilitiesReport{
		ModuleName: evidence.ModuleName,
		Required:   len(checks),
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}

func DefaultAetraValidatorScoreSubjectiveControlEvidence() AetraValidatorScoreSubjectiveControlEvidence {
	return AetraValidatorScoreSubjectiveControlEvidence{
		ModuleName: AetraValidatorScoreModuleName,

		NoSubjectiveCensorshipMechanism:  true,
		InformationalFirst:               true,
		RewardAffectingOnlyObjectiveData: true,
		ConsensusOverrideDisabledDefault: true,
		ObjectiveInputsDeterministic:     true,
	}
}

func ValidateAetraValidatorScoreSubjectiveControl(evidence AetraValidatorScoreSubjectiveControlEvidence) error {
	report := BuildAetraValidatorScoreSubjectiveControlReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator score subjective control failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorScoreSubjectiveControlReport(evidence AetraValidatorScoreSubjectiveControlEvidence) AetraValidatorScoreSubjectiveControlReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraValidatorScoreModuleName {
		failed = append(failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	}

	checks := []requirementCheck{
		{AetraValidatorScoreGuardNoSubjectiveCensorship, evidence.NoSubjectiveCensorshipMechanism},
		{AetraValidatorScoreGuardInformationalFirst, evidence.InformationalFirst},
		{AetraValidatorScoreGuardObjectiveRewardOnly, evidence.RewardAffectingOnlyObjectiveData},
		{AetraValidatorScoreGuardConsensusOverrideDisabled, evidence.ConsensusOverrideDisabledDefault},
		{AetraValidatorScoreGuardObjectiveInputsDeterministic, evidence.ObjectiveInputsDeterministic},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorScoreSubjectiveControlReport{
		ModuleName: evidence.ModuleName,
		Required:   len(checks),
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}
