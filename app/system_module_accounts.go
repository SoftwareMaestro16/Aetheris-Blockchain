package app

import (
	"fmt"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	delegatorprotectiontypes "github.com/sovereign-l1/l1/x/delegator-protection/types"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	systemregistrytypes "github.com/sovereign-l1/l1/x/system-registry/types"
	treasurytypes "github.com/sovereign-l1/l1/x/treasury/types"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
)

type ReservedSystemModuleAccount struct {
	Name                string
	ModuleName          string
	ModuleAccountName   string
	Raw                 string
	UserFriendly        string
	Core                bool
	CanHoldFunds        bool
	CanReceiveUserFunds bool
	CanSendFunds        bool
	Permissions         []string
}

var reservedSystemModuleAccountSpecs = []struct {
	addressName       string
	moduleName        string
	moduleAccountName string
	permissions       []string
}{
	{"AETMint", mintauthoritytypes.ModuleName, mintauthoritytypes.DefaultMintAuthorityModuleAccount, []string{authtypes.Minter}},
	{"AETFeeCollector", "fee-collector", feecollectortypes.CollectorModuleName, []string{authtypes.Burner}},
	{"AETTreasury", treasurytypes.ModuleName, treasurytypes.TreasuryModuleName, nil},
	{"AETBurn", burntypes.ModuleName, burntypes.ModuleName, []string{authtypes.Burner}},
	{"AETStorageRent", "storage-rent", storagerenttypes.ModuleName, nil},
	{"AETDelegatorProtection", delegatorprotectiontypes.ModuleName, delegatorprotectiontypes.ModuleName, nil},
	{"AETValidatorInsurance", validatorinsurancetypes.ModuleName, validatorinsurancetypes.ModuleName, nil},
	{"AETReporterRewards", "reporter-rewards", feecollectortypes.ReporterRewardsModuleName, nil},
	{"AETConfig", configtypes.ModuleName, configtypes.ModuleName, nil},
	{"AETSystemRegistry", systemregistrytypes.ModuleName, systemregistrytypes.ModuleName, nil},
	{"AETElector", validatorelectiontypes.ModuleName, validatorelectiontypes.ModuleName, nil},
}

func ReservedSystemModuleAccounts() []ReservedSystemModuleAccount {
	out := make([]ReservedSystemModuleAccount, 0, len(reservedSystemModuleAccountSpecs))
	for _, spec := range reservedSystemModuleAccountSpecs {
		address, found := aetherisaddress.SystemAddressByName(spec.addressName)
		if !found {
			panic(fmt.Sprintf("reserved system address %s is not registered", spec.addressName))
		}
		out = append(out, ReservedSystemModuleAccount{
			Name:                address.Name,
			ModuleName:          spec.moduleName,
			ModuleAccountName:   spec.moduleAccountName,
			Raw:                 address.Raw,
			UserFriendly:        address.UserFriendly,
			Core:                address.Core,
			CanHoldFunds:        address.CanHoldFunds,
			CanReceiveUserFunds: address.CanReceiveUserFunds,
			CanSendFunds:        address.CanSendFunds,
			Permissions:         append([]string(nil), spec.permissions...),
		})
	}
	return out
}

func ReservedSystemModuleAccountByName(name string) (ReservedSystemModuleAccount, bool) {
	for _, account := range ReservedSystemModuleAccounts() {
		if account.Name == name {
			return account, true
		}
	}
	return ReservedSystemModuleAccount{}, false
}

func ReservedSystemModuleAccountByModuleAccountName(moduleAccountName string) (ReservedSystemModuleAccount, bool) {
	for _, account := range ReservedSystemModuleAccounts() {
		if account.ModuleAccountName == moduleAccountName {
			return account, true
		}
	}
	return ReservedSystemModuleAccount{}, false
}

func IsReservedSystemModuleAccountName(moduleAccountName string) bool {
	_, found := ReservedSystemModuleAccountByModuleAccountName(moduleAccountName)
	return found
}

func ReservedSystemModuleAccountAddress(moduleAccountName string) (sdk.AccAddress, bool, error) {
	account, found := ReservedSystemModuleAccountByModuleAccountName(moduleAccountName)
	if !found {
		return nil, false, nil
	}
	bz, err := aetherisaddress.Parse(account.Raw)
	if err != nil {
		return nil, true, err
	}
	return sdk.AccAddress(bz), true, nil
}

func ValidateReservedSystemModuleAccountWiring(blocked map[string]bool) error {
	seen := map[string]string{}
	for _, address := range aetherisaddress.AllSystemAddresses() {
		bz, err := aetherisaddress.Parse(address.Raw)
		if err != nil {
			return fmt.Errorf("reserved system address %s raw address invalid: %w", address.Name, err)
		}
		if aetherisaddress.IsZero(bz) {
			return fmt.Errorf("reserved system address %s must not use zero address", address.Name)
		}
		key := sdk.AccAddress(bz).String()
		if other, found := seen[key]; found {
			return fmt.Errorf("reserved system address %s duplicates address with %s", address.Name, other)
		}
		seen[key] = address.Name
		if blocked[key] != !address.CanReceiveUserFunds {
			return fmt.Errorf("reserved system address %s blocked policy mismatch", address.Name)
		}
	}

	for _, account := range ReservedSystemModuleAccounts() {
		address, found := aetherisaddress.SystemAddressByName(account.Name)
		if !found {
			return fmt.Errorf("reserved module account %s is missing address catalog entry", account.Name)
		}
		if account.Raw != address.Raw || account.UserFriendly != address.UserFriendly {
			return fmt.Errorf("reserved module account %s address mismatch", account.Name)
		}
		if account.Core != address.Core || account.CanHoldFunds != address.CanHoldFunds ||
			account.CanReceiveUserFunds != address.CanReceiveUserFunds || account.CanSendFunds != address.CanSendFunds {
			return fmt.Errorf("reserved module account %s policy mismatch", account.Name)
		}
		if permissions, found := maccPerms[account.ModuleAccountName]; !found {
			return fmt.Errorf("reserved module account %s missing macc permission entry %s", account.Name, account.ModuleAccountName)
		} else if !sameStringSet(permissions, account.Permissions) {
			return fmt.Errorf("reserved module account %s permission mismatch", account.Name)
		}
		bz, err := aetherisaddress.Parse(account.Raw)
		if err != nil {
			return fmt.Errorf("reserved module account %s raw address invalid: %w", account.Name, err)
		}
		if aetherisaddress.IsZero(bz) {
			return fmt.Errorf("reserved module account %s must not use zero address", account.Name)
		}
		key := sdk.AccAddress(bz).String()
		if blocked[key] != !account.CanReceiveUserFunds {
			return fmt.Errorf("reserved module account %s blocked policy mismatch", account.Name)
		}
	}
	if mint, found := ReservedSystemModuleAccountByName("AETMint"); !found ||
		mint.ModuleAccountName != mintauthoritytypes.DefaultMintAuthorityModuleAccount ||
		mint.Raw != "4:030c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c308353" {
		return fmt.Errorf("mint authority address must be AETMint")
	}
	if burn, found := ReservedSystemModuleAccountByName("AETBurn"); !found ||
		burn.ModuleAccountName != burntypes.ModuleName ||
		burn.Raw != "4:004104104104104104104104104104104104104104104104104104104105444d" {
		return fmt.Errorf("burn sink address must be AETBurn")
	}
	return nil
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for _, value := range left {
		if !slices.Contains(right, value) {
			return false
		}
	}
	return true
}
