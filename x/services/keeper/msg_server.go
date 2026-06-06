package keeper

import (
	"context"
	"errors"

	servicestypes "github.com/sovereign-l1/l1/x/services/types"
)

var _ servicestypes.MsgServer = msgServer{}

type msgServer struct {
	keeper *Keeper
}

func NewMsgServerImpl(k *Keeper) servicestypes.MsgServer {
	return msgServer{keeper: k}
}

func (m msgServer) RegisterService(ctx context.Context, msg *servicestypes.MsgRegisterService) (*servicestypes.MsgRegisterServiceResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty register service request")
	}
	return &servicestypes.MsgRegisterServiceResponse{}, m.keeper.RegisterService(*msg)
}

func (m msgServer) UpdateService(ctx context.Context, msg *servicestypes.MsgUpdateService) (*servicestypes.MsgUpdateServiceResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty update service request")
	}
	return &servicestypes.MsgUpdateServiceResponse{}, m.keeper.UpdateService(*msg)
}

func (m msgServer) RegisterInterface(ctx context.Context, msg *servicestypes.MsgRegisterInterface) (*servicestypes.MsgRegisterInterfaceResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty register interface request")
	}
	return &servicestypes.MsgRegisterInterfaceResponse{}, m.keeper.RegisterInterface(*msg)
}

func (m msgServer) UpdateInterface(ctx context.Context, msg *servicestypes.MsgUpdateInterface) (*servicestypes.MsgUpdateInterfaceResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty update interface request")
	}
	return &servicestypes.MsgUpdateInterfaceResponse{}, m.keeper.UpdateInterface(*msg)
}

func (m msgServer) RenewService(ctx context.Context, msg *servicestypes.MsgRenewService) (*servicestypes.MsgRenewServiceResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty renew service request")
	}
	return &servicestypes.MsgRenewServiceResponse{}, m.keeper.RenewService(*msg)
}

func (m msgServer) DisableService(ctx context.Context, msg *servicestypes.MsgDisableService) (*servicestypes.MsgDisableServiceResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty disable service request")
	}
	return &servicestypes.MsgDisableServiceResponse{}, m.keeper.DisableService(*msg)
}

func (m msgServer) TransferService(ctx context.Context, msg *servicestypes.MsgTransferService) (*servicestypes.MsgTransferServiceResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty transfer service request")
	}
	return &servicestypes.MsgTransferServiceResponse{}, m.keeper.TransferService(*msg)
}

func (m msgServer) BindServiceIdentity(ctx context.Context, msg *servicestypes.MsgBindServiceIdentity) (*servicestypes.MsgBindServiceIdentityResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty bind identity request")
	}
	return &servicestypes.MsgBindServiceIdentityResponse{}, m.keeper.BindServiceIdentity(*msg)
}

func (m msgServer) UnbindServiceIdentity(ctx context.Context, msg *servicestypes.MsgUnbindServiceIdentity) (*servicestypes.MsgUnbindServiceIdentityResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty unbind identity request")
	}
	return &servicestypes.MsgUnbindServiceIdentityResponse{}, m.keeper.UnbindServiceIdentity(*msg)
}

func (m msgServer) RegisterProvider(ctx context.Context, msg *servicestypes.MsgRegisterProvider) (*servicestypes.MsgRegisterProviderResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty register provider request")
	}
	return &servicestypes.MsgRegisterProviderResponse{}, m.keeper.RegisterProvider(*msg)
}

func (m msgServer) UpdateProvider(ctx context.Context, msg *servicestypes.MsgUpdateProvider) (*servicestypes.MsgUpdateProviderResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty update provider request")
	}
	return &servicestypes.MsgUpdateProviderResponse{}, m.keeper.UpdateProvider(*msg)
}

func (m msgServer) StakeProviderCollateral(ctx context.Context, msg *servicestypes.MsgStakeProviderCollateral) (*servicestypes.MsgStakeProviderCollateralResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty stake provider collateral request")
	}
	return &servicestypes.MsgStakeProviderCollateralResponse{}, m.keeper.StakeProviderCollateral(*msg)
}

func (m msgServer) UnstakeProviderCollateral(ctx context.Context, msg *servicestypes.MsgUnstakeProviderCollateral) (*servicestypes.MsgUnstakeProviderCollateralResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty unstake provider collateral request")
	}
	return &servicestypes.MsgUnstakeProviderCollateralResponse{}, m.keeper.UnstakeProviderCollateral(*msg)
}

func (m msgServer) AnchorServiceReceipt(ctx context.Context, msg *servicestypes.MsgAnchorServiceReceipt) (*servicestypes.MsgAnchorServiceReceiptResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty anchor receipt request")
	}
	return &servicestypes.MsgAnchorServiceReceiptResponse{}, m.keeper.AnchorServiceReceipt(*msg)
}

func (m msgServer) SubmitServiceDispute(ctx context.Context, msg *servicestypes.MsgSubmitServiceDispute) (*servicestypes.MsgSubmitServiceDisputeResponse, error) {
	if msg == nil {
		return nil, errors.New("services empty submit dispute request")
	}
	return &servicestypes.MsgSubmitServiceDisputeResponse{}, m.keeper.SubmitServiceDispute(*msg)
}
