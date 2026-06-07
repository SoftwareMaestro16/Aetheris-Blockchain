package keeper

import "github.com/sovereign-l1/l1/x/aetra-economics/types"

type MsgServer struct {
	Keeper *Keeper
}

func NewMsgServerImpl(k *Keeper) MsgServer {
	return MsgServer{Keeper: k}
}

func (m MsgServer) UpdateEconomicsParams(msg types.MsgUpdateEconomicsParams) error {
	if err := m.requireAuthority(msg.Authority); err != nil {
		return err
	}
	return m.Keeper.SetParams(msg.Params)
}

func (m MsgServer) ApplyEpochEconomics(msg types.MsgApplyEpochEconomics) error {
	if err := m.requireAuthority(msg.Authority); err != nil {
		return err
	}
	_, err := m.Keeper.ApplyEpoch(msg.Input)
	return err
}

func (m MsgServer) requireAuthority(authority string) error {
	if authority != m.Keeper.Authority() {
		return types.ErrUnauthorized.Wrap("invalid authority")
	}
	return nil
}
