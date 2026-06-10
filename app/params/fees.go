package params

import "fmt"

func ValidateNativeFeeDenomsV1(denoms []string, maxAllowed int) error {
	if len(denoms) == 0 {
		return fmt.Errorf("allowed_fee_denoms must contain %s", BaseDenom)
	}
	if len(denoms) > maxAllowed {
		return fmt.Errorf("v1 only allows %d fee denom: %s", maxAllowed, BaseDenom)
	}
	if denoms[0] != BaseDenom {
		return fmt.Errorf("v1 only accepts fee denom %s", BaseDenom)
	}
	return nil
}
