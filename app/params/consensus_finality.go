package params

import "fmt"

type ConsensusFinalityReport struct {
	ValidatorCount              int
	BlocksObserved              int
	LocalnetStable              bool
	LoadProfileExecuted         bool
	ObservedBlockTimeMinSeconds int
	ObservedBlockTimeMaxSeconds int
	NormalFinalitySeconds       int
	StressFinalitySeconds       int
	WorstFinalitySeconds        int
	DegradedScenarioExecuted    bool
	HealthyVotingPowerBps       int64
	LivenessPreserved           bool
	IncludedInTestnetReport     bool
}

func ValidateConsensusFinalityReport(profile NetworkProfile, report ConsensusFinalityReport) error {
	if err := profile.Validate(); err != nil {
		return err
	}
	phase, err := profile.ValidatorSetPhase(report.ValidatorCount)
	if err != nil {
		return err
	}
	if report.BlocksObserved <= 0 {
		return fmt.Errorf("consensus finality report must include observed blocks")
	}
	if report.ValidatorCount >= AetraValidatorSetGenesisMin && report.ValidatorCount <= AetraValidatorSetGenesisMax && !report.LocalnetStable {
		return fmt.Errorf("100-128 validator localnet must remain stable")
	}
	if !report.LoadProfileExecuted {
		return fmt.Errorf("localnet/load profile must execute")
	}
	if report.ObservedBlockTimeMinSeconds <= 2 || report.ObservedBlockTimeMaxSeconds <= 2 {
		return fmt.Errorf("1-2 second block targets are not allowed")
	}
	if report.ObservedBlockTimeMinSeconds < phase.BlockTimeMinSeconds || report.ObservedBlockTimeMaxSeconds > phase.BlockTimeMaxSeconds || report.ObservedBlockTimeMinSeconds > report.ObservedBlockTimeMaxSeconds {
		return fmt.Errorf("observed block time must stay within %d-%d seconds for phase %q", phase.BlockTimeMinSeconds, phase.BlockTimeMaxSeconds, phase.Name)
	}
	if report.NormalFinalitySeconds < profile.NormalFinalityMinSeconds || report.NormalFinalitySeconds > profile.NormalFinalityMaxSeconds {
		return fmt.Errorf("normal finality must stay within %d-%d seconds", profile.NormalFinalityMinSeconds, profile.NormalFinalityMaxSeconds)
	}
	if report.StressFinalitySeconds < profile.StressFinalityMinSeconds || report.StressFinalitySeconds > profile.StressFinalityMaxSeconds {
		return fmt.Errorf("stress finality must stay within %d-%d seconds", profile.StressFinalityMinSeconds, profile.StressFinalityMaxSeconds)
	}
	if report.WorstFinalitySeconds <= 0 || report.WorstFinalitySeconds > profile.WorstFinalityTargetSeconds {
		return fmt.Errorf("worst finality target must be <= %d seconds", profile.WorstFinalityTargetSeconds)
	}
	if report.WorstFinalitySeconds < report.NormalFinalitySeconds || report.WorstFinalitySeconds < report.StressFinalitySeconds {
		return fmt.Errorf("worst finality must cover normal and stress observations")
	}
	if !report.DegradedScenarioExecuted {
		return fmt.Errorf("degraded validator scenario must execute")
	}
	if report.HealthyVotingPowerBps >= AetraHealthyVotingPowerBps && !report.LivenessPreserved {
		return fmt.Errorf("degraded scenario must preserve liveness when >= 2/3 voting power is healthy")
	}
	if !report.IncludedInTestnetReport {
		return fmt.Errorf("finality measurements must be included in testnet reports")
	}
	return nil
}
