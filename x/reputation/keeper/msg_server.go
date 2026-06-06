package keeper

import (
	"context"
	"encoding/json"

	"github.com/sovereign-l1/l1/x/reputation/types"
	reputationpb "github.com/sovereign-l1/l1/x/reputation/types/reputationpb"
)

var _ reputationpb.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) reputationpb.MsgServer { return msgServer{Keeper: k} }

func (m msgServer) UpdateReputationParams(ctx context.Context, msg *reputationpb.MsgUpdateReputationParams) (*reputationpb.MsgUpdateReputationParamsResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	var params types.ReputationParams
	if err := json.Unmarshal([]byte(msg.ParamsJson), &params); err != nil {
		return nil, err
	}
	next, err := types.ApplyUpdateReputationParams(state, types.MsgUpdateReputationParams{Authority: msg.Authority, Params: params})
	if err != nil {
		return nil, err
	}
	return &reputationpb.MsgUpdateReputationParamsResponse{}, m.SetState(ctx, next)
}

func (m msgServer) ApplyReputationPenalty(ctx context.Context, msg *reputationpb.MsgApplyReputationPenalty) (*reputationpb.MsgApplyReputationPenaltyResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	subject, err := parseAddress(msg.Subject)
	if err != nil {
		return nil, err
	}
	next, err := types.ApplyReputationPenalty(state, types.MsgApplyReputationPenalty{Authority: msg.Authority, SubjectType: msg.SubjectType, Subject: subject, Component: msg.Component, Amount: uint16(msg.Amount), Reason: msg.Reason, Epoch: msg.Epoch})
	if err != nil {
		return nil, err
	}
	record, _ := nextRecord(next, msg.SubjectType, subject)
	recordJSON, err := mustJSON(record)
	if err != nil {
		return nil, err
	}
	return &reputationpb.MsgApplyReputationPenaltyResponse{RecordJson: recordJSON}, m.SetState(ctx, next)
}

func (m msgServer) ApplyReputationReward(ctx context.Context, msg *reputationpb.MsgApplyReputationReward) (*reputationpb.MsgApplyReputationRewardResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	subject, err := parseAddress(msg.Subject)
	if err != nil {
		return nil, err
	}
	next, err := types.ApplyReputationReward(state, types.MsgApplyReputationReward{Authority: msg.Authority, SubjectType: msg.SubjectType, Subject: subject, Component: msg.Component, Amount: uint16(msg.Amount), Reason: msg.Reason, Epoch: msg.Epoch})
	if err != nil {
		return nil, err
	}
	record, _ := nextRecord(next, msg.SubjectType, subject)
	recordJSON, err := mustJSON(record)
	if err != nil {
		return nil, err
	}
	return &reputationpb.MsgApplyReputationRewardResponse{RecordJson: recordJSON}, m.SetState(ctx, next)
}

func (m msgServer) RecomputeReputation(ctx context.Context, msg *reputationpb.MsgRecomputeReputation) (*reputationpb.MsgRecomputeReputationResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	subject, err := parseAddress(msg.Subject)
	if err != nil {
		return nil, err
	}
	next, err := types.ApplyRecomputeReputation(state, types.MsgRecomputeReputation{Authority: msg.Authority, SubjectType: msg.SubjectType, Subject: subject, Epoch: msg.Epoch})
	if err != nil {
		return nil, err
	}
	record, _ := nextRecord(next, msg.SubjectType, subject)
	recordJSON, err := mustJSON(record)
	if err != nil {
		return nil, err
	}
	return &reputationpb.MsgRecomputeReputationResponse{RecordJson: recordJSON}, m.SetState(ctx, next)
}

func nextRecord(state types.ReputationState, subjectType string, subject []byte) (types.ReputationRecord, bool) {
	if subjectType == types.SubjectReporter {
		return types.QueryReporterReputation(state, subject)
	}
	return types.QueryValidatorReputation(state, subject)
}
