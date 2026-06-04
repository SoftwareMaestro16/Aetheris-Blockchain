package types

import (
	"testing"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestDefaultParamsValidate(t *testing.T) {
	params := DefaultParams()
	if len(params.AllowedFeeDenoms) != 1 || params.AllowedFeeDenoms[0] != appparams.BaseDenom {
		t.Fatalf("expected only norb as default fee denom: %v", params.AllowedFeeDenoms)
	}
	if err := params.Validate(); err != nil {
		t.Fatalf("default params should validate: %v", err)
	}
}

func TestParamsRejectInvalidAllowedFeeDenoms(t *testing.T) {
	tests := map[string][]string{
		"empty list":       {},
		"non native denom": {"uatom"},
		"duplicate native": {appparams.BaseDenom, appparams.BaseDenom},
		"mixed denoms":     {appparams.BaseDenom, "testtoken"},
	}

	for name, denoms := range tests {
		t.Run(name, func(t *testing.T) {
			params := DefaultParams()
			params.AllowedFeeDenoms = denoms
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid allowed fee denoms to fail")
			}
		})
	}
}

func TestParamsRejectInvalidFeeSplitRatios(t *testing.T) {
	tests := map[string]func(*Params){
		"malformed validator ratio": func(params *Params) {
			params.ValidatorRewardsRatio = "not-a-decimal"
		},
		"malformed community ratio": func(params *Params) {
			params.CommunityPoolRatio = "not-a-decimal"
		},
		"negative ratio": func(params *Params) {
			params.ValidatorRewardsRatio = "-0.1"
			params.CommunityPoolRatio = "1.1"
		},
		"sum not one": func(params *Params) {
			params.ValidatorRewardsRatio = "0.80"
			params.CommunityPoolRatio = "0.10"
		},
		"ratio greater than one": func(params *Params) {
			params.ValidatorRewardsRatio = "1.01"
			params.CommunityPoolRatio = "-0.01"
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			params := DefaultParams()
			mutate(&params)
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid fee split params to fail")
			}
		})
	}
}
