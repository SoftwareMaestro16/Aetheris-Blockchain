package keeper

import paymentstypes "github.com/sovereign-l1/l1/x/payments/types"

func (k Keeper) SnapshotRoutingEngine(store paymentstypes.TopologyStore, policy paymentstypes.RoutePolicy, rateLimit paymentstypes.GossipRateLimitPolicy, failurePolicy paymentstypes.RouteFailureScoringPolicy) (paymentstypes.RoutingEngineState, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.RoutingEngineState{}, err
	}
	return paymentstypes.SnapshotRoutingEngineState(store, policy, rateLimit, failurePolicy)
}

func (k Keeper) ApplyRoutingGossip(engine paymentstypes.RoutingEngineState, msg paymentstypes.RoutingEngineMessage, currentHeight uint64) (paymentstypes.RoutingEngineState, paymentstypes.GossipRateLimitDecision, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.RoutingEngineState{}, paymentstypes.GossipRateLimitDecision{}, err
	}
	return paymentstypes.ApplyRoutingEngineMessage(engine, k.genesis.State, msg, currentHeight)
}

func (k Keeper) PruneRoutingTopology(engine paymentstypes.RoutingEngineState, currentHeight uint64) (paymentstypes.RoutingEngineState, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.RoutingEngineState{}, err
	}
	return paymentstypes.PruneRoutingEngineTopology(engine, currentHeight)
}

func (k Keeper) SelectRoutingPath(engine paymentstypes.RoutingEngineState, req paymentstypes.RouteSelectionRequest) (paymentstypes.RoutingEngineState, paymentstypes.ScoredRoute, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.RoutingEngineState{}, paymentstypes.ScoredRoute{}, err
	}
	return paymentstypes.SelectRoutingEnginePath(engine, k.genesis.State, req)
}

func (k Keeper) RetryRoutingPath(engine paymentstypes.RoutingEngineState, req paymentstypes.RouteRetryRequest) (paymentstypes.RoutingEngineState, paymentstypes.RouteRetryResult, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.RoutingEngineState{}, paymentstypes.RouteRetryResult{}, err
	}
	return paymentstypes.RetryRoutingEnginePath(engine, k.genesis.State, req)
}

func (k Keeper) ApplyRoutingFailures(engine paymentstypes.RoutingEngineState, reports []paymentstypes.RouteFailureReport) (paymentstypes.RoutingEngineState, []paymentstypes.RouteFailureScore, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.RoutingEngineState{}, nil, err
	}
	return paymentstypes.ApplyRoutingEngineFailures(engine, reports)
}

func (k Keeper) HandleCapacityProbe(engine paymentstypes.RoutingEngineState, req paymentstypes.CapacityProbeRequest, responder string) (paymentstypes.CapacityProbeResponse, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.CapacityProbeResponse{}, err
	}
	return paymentstypes.HandleCapacityProbe(engine, k.genesis.State, req, responder)
}

func (k Keeper) SimulateRoutingSpam(engine paymentstypes.RoutingEngineState, envelopes []paymentstypes.SignedGossipEnvelope, currentHeight uint64) (paymentstypes.RoutingEngineState, paymentstypes.TopologySpamSimulation, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.RoutingEngineState{}, paymentstypes.TopologySpamSimulation{}, err
	}
	store, sim, err := paymentstypes.SimulateTopologySpam(k.genesis.State, engine.Topology, envelopes, currentHeight, engine.RateLimit)
	if err != nil {
		return paymentstypes.RoutingEngineState{}, paymentstypes.TopologySpamSimulation{}, err
	}
	next, err := paymentstypes.SnapshotRoutingEngineState(store, engine.Policy, engine.RateLimit, engine.FailureScoring)
	if err != nil {
		return paymentstypes.RoutingEngineState{}, paymentstypes.TopologySpamSimulation{}, err
	}
	return next, sim, nil
}
