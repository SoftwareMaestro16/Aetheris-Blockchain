package types

import "context"

// ReputationKeeper defines the interface for reading validator reputation scores.
type ReputationKeeper interface {
	// GetValidatorTotalScore returns the total score and jailed/slashed status.
	GetValidatorTotalScore(ctx context.Context, addr string) (totalScore uint32, jailed bool, err error)
}
