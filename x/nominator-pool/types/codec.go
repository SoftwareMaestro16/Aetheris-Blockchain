package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateNominatorPool{}, "l1/nominatorpool/MsgCreateNominatorPool", nil)
	cdc.RegisterConcrete(&MsgDepositToPool{}, "l1/nominatorpool/MsgDepositToPool", nil)
	cdc.RegisterConcrete(&MsgRequestPoolWithdrawal{}, "l1/nominatorpool/MsgRequestPoolWithdrawal", nil)
	cdc.RegisterConcrete(&MsgCancelPoolWithdrawal{}, "l1/nominatorpool/MsgCancelPoolWithdrawal", nil)
	cdc.RegisterConcrete(&MsgDepositToStakingPool{}, "l1/nominatorpool/MsgDepositToStakingPool", nil)
	cdc.RegisterConcrete(&MsgRequestPoolUnbond{}, "l1/nominatorpool/MsgRequestPoolUnbond", nil)
	cdc.RegisterConcrete(&MsgClaimPoolRewards{}, "l1/nominatorpool/MsgClaimPoolRewards", nil)
	cdc.RegisterConcrete(&MsgSyncPoolRewards{}, "l1/nominatorpool/MsgSyncPoolRewards", nil)
	cdc.RegisterConcrete(&MsgClaimStakingRewards{}, "l1/nominatorpool/MsgClaimStakingRewards", nil)
	cdc.RegisterConcrete(&MsgClaimStakeReputation{}, "l1/nominatorpool/MsgClaimStakeReputation", nil)
	cdc.RegisterConcrete(&MsgDelegateToValidator{}, "l1/nominatorpool/MsgDelegateToValidator", nil)
	cdc.RegisterConcrete(&MsgTopUpPoolReserve{}, "l1/nominatorpool/MsgTopUpPoolReserve", nil)
	cdc.RegisterConcrete(&MsgUpdatePoolCommission{}, "l1/nominatorpool/MsgUpdatePoolCommission", nil)
	cdc.RegisterConcrete(&MsgChangePoolValidator{}, "l1/nominatorpool/MsgChangePoolValidator", nil)
	cdc.RegisterConcrete(&MsgRegisterValidator{}, "l1/nominatorpool/MsgRegisterValidator", nil)
	cdc.RegisterConcrete(&MsgUpdateValidator{}, "l1/nominatorpool/MsgUpdateValidator", nil)
	cdc.RegisterConcrete(&MsgCreateOfficialLiquidStakingPool{}, "l1/nominatorpool/MsgCreateOfficialLiquidStakingPool", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &Msg_serviceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateNominatorPool{},
		&MsgDepositToPool{},
		&MsgRequestPoolWithdrawal{},
		&MsgCancelPoolWithdrawal{},
		&MsgSyncPoolRewards{},
		&MsgClaimStakingRewards{},
		&MsgUpdatePoolCommission{},
		&MsgChangePoolValidator{},
		&MsgDepositToStakingPool{},
		&MsgRequestPoolUnbond{},
		&MsgWithdrawPoolStake{},
		&MsgTopUpPoolReserve{},
		&MsgClaimPoolRewards{},
		&MsgClaimStakeReputation{},
		&MsgDelegateToValidator{},
		&MsgRegisterValidator{},
		&MsgUpdateValidator{},
		&MsgUpdateStakingParams{},
		&MsgCreateOfficialLiquidStakingPool{},
	)
	registry.RegisterImplementations(
		(*txtypes.MsgResponse)(nil),
		&MsgCreateNominatorPoolResponse{},
		&MsgDepositToPoolResponse{},
		&MsgRequestPoolWithdrawalResponse{},
		&MsgCancelPoolWithdrawalResponse{},
		&MsgSyncPoolRewardsResponse{},
		&MsgClaimStakingRewardsResponse{},
		&MsgUpdatePoolCommissionResponse{},
		&MsgChangePoolValidatorResponse{},
		&MsgDepositToStakingPoolResponse{},
		&MsgRequestPoolUnbondResponse{},
		&MsgWithdrawPoolStakeResponse{},
		&MsgTopUpPoolReserveResponse{},
		&MsgClaimPoolRewardsResponse{},
		&MsgClaimStakeReputationResponse{},
		&MsgDelegateToValidatorResponse{},
		&MsgRegisterValidatorResponse{},
		&MsgUpdateValidatorResponse{},
		&MsgUpdateStakingParamsResponse{},
		&MsgCreateOfficialLiquidStakingPoolResponse{},
	)
}
