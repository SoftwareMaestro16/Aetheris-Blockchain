package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestExpensiveContractChargedMore(t *testing.T) {
	params := DefaultParams()
	cheap := Operation{Contract: addr(0x11), Op: OpStorageRead, Count: 1}
	expensive := Operation{Contract: addr(0x11), Op: OpHeavyCompute, Count: 1}

	cheapCost, err := ComputeCost(params, []Operation{cheap})
	require.NoError(t, err)
	expensiveCost, err := ComputeCost(params, []Operation{expensive})
	require.NoError(t, err)

	require.Greater(t, expensiveCost, cheapCost)
}

func TestComputeCapsAreEnforced(t *testing.T) {
	params := DefaultParams()
	params.MaxBlockComputeUnits = 100
	params.MaxContractComputeUnits = 50
	params.OpCosts = map[string]uint64{OpContractCall: 30}
	meter, err := NewBlockMeter(params)
	require.NoError(t, err)

	require.NoError(t, meter.Charge(Operation{Contract: addr(0x22), Op: OpContractCall, Count: 1}))
	require.ErrorContains(t, meter.Charge(Operation{Contract: addr(0x22), Op: OpContractCall, Count: 1}), "contract compute cap")

	require.NoError(t, meter.Charge(Operation{Contract: addr(0x33), Op: OpContractCall, Count: 1}))
	require.NoError(t, meter.Charge(Operation{Contract: addr(0x44), Op: OpContractCall, Count: 1}))
	require.ErrorContains(t, meter.Charge(Operation{Contract: addr(0x55), Op: OpContractCall, Count: 1}), "block compute cap")
}

func TestComputeAccountingDeterministic(t *testing.T) {
	meter, err := NewBlockMeter(DefaultParams())
	require.NoError(t, err)
	require.NoError(t, meter.Charge(Operation{Contract: addr(0x44), Op: OpStorageWrite, Count: 2}))
	require.NoError(t, meter.Charge(Operation{Contract: addr(0x11), Op: OpStorageRead, Count: 1}))

	stats := meter.Stats()
	require.Len(t, stats, 2)
	require.Equal(t, addr(0x11), stats[0].Contract)
	require.Equal(t, addr(0x44), stats[1].Contract)
	require.Equal(t, uint64(45), meter.Used())
}

func TestComputeRejectsZeroContract(t *testing.T) {
	_, err := ComputeCost(DefaultParams(), []Operation{{Contract: addr(0), Op: OpNoop, Count: 1}})
	require.ErrorContains(t, err, "must not be zero address")
}

func addr(fill byte) sdk.AccAddress {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return sdk.AccAddress(out)
}
