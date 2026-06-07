package upgrades

import (
	"github.com/cosmos/cosmos-sdk/types/module"

	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
)

func NativeAccountVersionUpgradePlan() nativeaccounttypes.AccountVersionUpgradePlan {
	return nativeaccounttypes.NativeAccountVersionUpgradePlan()
}

func ValidateNativeAccountVersionUpgradeHandler(fromVM, currentVM module.VersionMap) error {
	plan := NativeAccountVersionUpgradePlan()
	if err := nativeaccounttypes.ValidateNativeAccountVersionUpgradePlan(plan); err != nil {
		return err
	}
	return ValidateVersionMap(fromVM, currentVM, plan.ModuleName)
}
