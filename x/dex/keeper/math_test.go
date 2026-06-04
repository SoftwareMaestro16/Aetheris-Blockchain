package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func TestCalcSwapOutAppliesFee(t *testing.T) {
	out := calcSwapOut(sdkmath.NewInt(1000), sdkmath.NewInt(1000), sdkmath.NewInt(100))
	if !out.Equal(sdkmath.NewInt(90)) {
		t.Fatalf("unexpected output amount: %s", out)
	}
}

func TestCalcSwapOutRoundsTinyInputToZero(t *testing.T) {
	out := calcSwapOut(sdkmath.NewInt(1000), sdkmath.NewInt(1000), sdkmath.NewInt(1))
	if !out.IsZero() {
		t.Fatalf("tiny input should round to zero, got %s", out)
	}
}

func TestCalcSwapOutIsMonotonicForLargerInput(t *testing.T) {
	small := calcSwapOut(sdkmath.NewInt(1000), sdkmath.NewInt(1000), sdkmath.NewInt(100))
	large := calcSwapOut(sdkmath.NewInt(1000), sdkmath.NewInt(1000), sdkmath.NewInt(200))
	if !large.GT(small) {
		t.Fatalf("larger input must produce larger output: small=%s large=%s", small, large)
	}
}
