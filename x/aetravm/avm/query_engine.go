package avm

import (
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
)

// QueryEngine handles read-only AVM queries.
type QueryEngine struct{}

func NewQueryEngine() *QueryEngine {
	return &QueryEngine{}
}

// ExecuteQuery runs a read-only query against a snapshot.
func (e *QueryEngine) ExecuteQuery(snapshot QuerySnapshot, method string, args []byte, gasLimit uint64) (QueryReceipt, error) {
	frame := &QueryFrame{
		Snapshot: snapshot,
		GasLimit: gasLimit,
	}

	// 1. Validate arguments (placeholder)
	// 2. Load code and state root (from snapshot)

	// 3. Execution loop (simulated)
	// In the execution loop, host calls MUST check:
	// if spec.Class == ClassEffectful { return Forbidden }

	// Simulation of success:
	frame.GasUsed = 100
	frame.ExitCode = contractstypes.ExitCodeOK

	return QueryReceipt{
		ExitCode:  frame.ExitCode,
		GasUsed:   frame.GasUsed,
		Response:  []byte("mock_response"),
		TraceHash: "mock_trace_hash",
	}, nil
}
