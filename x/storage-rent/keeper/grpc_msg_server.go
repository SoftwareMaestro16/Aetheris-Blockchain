package keeper

import (
	"context"

	"github.com/sovereign-l1/l1/x/storage-rent/types"
)

type msgServer struct {
	k *Keeper
}

func NewMsgServer(k *Keeper) types.MsgServer {
	return &msgServer{k: k}
}

var _ types.MsgServer = &msgServer{}

func (m *msgServer) PayStorageRent(ctx context.Context, req *types.MsgPayStorageRent) (*types.MsgPayStorageRentResponse, error) {
	_, _, err := m.k.PayStorageRent(ctx, *req)
	if err != nil {
		return nil, err
	}
	return &types.MsgPayStorageRentResponse{}, nil
}

func (m *msgServer) UnfreezeContract(ctx context.Context, req *types.MsgUnfreezeContract) (*types.MsgUnfreezeContractResponse, error) {
	_, _, err := m.k.UnfreezeContract(ctx, *req)
	if err != nil {
		return nil, err
	}
	return &types.MsgUnfreezeContractResponse{}, nil
}
