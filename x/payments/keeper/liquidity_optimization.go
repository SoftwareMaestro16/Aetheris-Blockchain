package keeper

import paymentstypes "github.com/sovereign-l1/l1/x/payments/types"

func (k Keeper) LiquidityOptimizationState() (paymentstypes.LiquidityOptimizationState, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.LiquidityOptimizationState{}, err
	}
	return k.genesis.Liquidity.Export(), nil
}

func (k *Keeper) HandleLiquidityOptimizationMessage(msg paymentstypes.LiquidityOptimizationMessage) (paymentstypes.LiquidityOptimizationState, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.LiquidityOptimizationState{}, err
	}
	next, err := paymentstypes.ApplyLiquidityOptimizationMessage(k.genesis.State, k.genesis.Liquidity, msg)
	if err != nil {
		return paymentstypes.LiquidityOptimizationState{}, err
	}
	k.genesis.Liquidity = next
	return next.Export(), nil
}

func (k *Keeper) ExpireLiquidityReservations(currentHeight uint64) ([]paymentstypes.Reservation, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return nil, err
	}
	next, expired, err := paymentstypes.ExpireLiquidityReservations(k.genesis.State, k.genesis.Liquidity, currentHeight)
	if err != nil {
		return nil, err
	}
	k.genesis.Liquidity = next
	return expired, nil
}

func (k *Keeper) DecayLiquidityScores(currentHeight, halfLife uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.DecayLiquidityScores(k.genesis.Liquidity, currentHeight, halfLife)
	if err != nil {
		return err
	}
	k.genesis.Liquidity = next
	return nil
}

func (k Keeper) CapacityForecast(channelID, from, to string, currentHeight, ttl uint64) (paymentstypes.CapacityForecast, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.CapacityForecast{}, err
	}
	return paymentstypes.BuildCapacityForecast(k.genesis.State, k.genesis.Liquidity, channelID, from, to, currentHeight, ttl)
}
