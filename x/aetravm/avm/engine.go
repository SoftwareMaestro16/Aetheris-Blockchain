package avm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
)

// Engine is the core AVM execution coordinator.
type Engine struct {
	// ... potentially some global state or config ...
}

func NewEngine() *Engine {
	return &Engine{}
}

// Execute performs a deterministic state transition.
// (StateChunk, Message, BlockContext) -> (NewStateChunk, Actions, Receipt, error)
func (e *Engine) Execute(state *chunk.Chunk, msg Message, blockCtx BlockContext, gasLimit uint64) (*chunk.Chunk, []Action, AVMReceipt, error) {
	msg.GasLimit = gasLimit // Override if provided explicitly
	frame := NewExecutionFrame(state, msg)

	// Set sandbox boundaries
	frame.BlockCtx = blockCtx
	frame.Capabilities = CapabilityMask{
		Crypto:    true,
		Chain:     true,
		Messaging: true,
		Storage:   true,
	}

	// Phase 1: Storage Phase - Load state Chunks (READ ONLY snapshot)
	frame.Phase = PhaseStorage
	// In a real implementation, this would track which chunks are touched.
	if !frame.ChargeGas(500) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}

	// Phase 2: Credit Phase - Apply attached value
	frame.Phase = PhaseCredit
	if !frame.ChargeGas(100) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}
	// TODO: Update working state balance based on message value

	// Phase 3: Compute Phase - Execute instructions
	frame.Phase = PhaseCompute
	// Simulation of VM execution:
	// We use the payload to simulate different execution paths.
	payloadData := frame.Message.Payload.Data()

	if string(payloadData) == "trigger_abort" {
		frame.Aborted = true
		return frame.finalize(contractstypes.ExitCodeContractAbort)
	}

	if string(payloadData) == "use_forbidden_opcode" {
		return frame.finalize(contractstypes.ExitCodeCodeRejected)
	}

	if string(payloadData) == "emit_actions" || string(payloadData) == "emit_with_bounce" {
		frame.PendingActions = append(frame.PendingActions, Action{
			Type:    ActionInternal,
			Target:  "contract_b",
			Payload: frame.Message.Payload, // Reuse payload for simplicity
		})

		if string(payloadData) == "emit_with_bounce" {
			frame.PendingActions = append(frame.PendingActions, Action{
				Type:         ActionSystem,
				Target:       "system_notifier",
				Payload:      frame.Message.Payload,
				SystemBounce: true,
			})
		} else {
			frame.PendingActions = append(frame.PendingActions, Action{
				Type:    ActionExternal,
				Target:  "user_a",
				Payload: frame.Message.Payload,
			})
		}
	}

	// We record a trace step for demonstration.
	frame.Trace.Steps = append(frame.Trace.Steps, TraceStep{
		Instruction: "LOAD_BAL",
		StackDelta:  1,
		GasConsumed: 10,
		Phase:       PhaseCompute,
	})

	if !frame.ChargeGas(1000) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}

	// Phase 4: Action Phase - Emit outgoing messages/events
	frame.Phase = PhaseAction
	if !frame.ChargeGas(200) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}
	// Actions are collected in f.PendingActions during PhaseCompute.

	if uint32(len(frame.PendingActions)) > frame.ActionBudget {
		frame.Aborted = true
		return frame.finalize(contractstypes.ExitCodeActionBudgetExceeded)
	}
	frame.ActionsUsed = uint32(len(frame.PendingActions))

	// Phase 5: Finalization Phase - Commit new Chunk roots
	frame.Phase = PhaseFinalization
	if !frame.ChargeGas(300) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}

	return frame.finalize(contractstypes.ExitCodeOK)
}

func (f *ExecutionFrame) finalize(exitCode uint32) (*chunk.Chunk, []Action, AVMReceipt, error) {
	f.ExitCode = exitCode

	receipt := AVMReceipt{
		ExitCode:        f.ExitCode,
		GasUsed:         f.GasUsed,
		GasLimit:        f.GasLimit,
		PhaseGas:        f.PhaseGas,
		StateRootBefore: hex.EncodeToString(f.StateSnapshot.Hash()),
	}

	// Action Determinism: Sort actions canonically
	sort.SliceStable(f.PendingActions, func(i, j int) bool {
		if f.PendingActions[i].Type != f.PendingActions[j].Type {
			return f.PendingActions[i].Type < f.PendingActions[j].Type
		}
		return f.PendingActions[i].Target < f.PendingActions[j].Target
	})

	// Revert Model
	if f.Aborted || f.ExitCode != contractstypes.ExitCodeOK {
		// Revert: return original state, discard non-system actions
		receipt.StateRootAfter = receipt.StateRootBefore

		var finalActions []Action
		for _, action := range f.PendingActions {
			if action.SystemBounce {
				finalActions = append(finalActions, action)
			}
		}
		receipt.EmittedActionsHash = f.computeActionsHash(finalActions)
		receipt.ExecutionTraceHash = f.computeTraceHash()

		return f.StateSnapshot, finalActions, receipt, nil
	}

	// Success: return working state and actions
	receipt.StateRootAfter = hex.EncodeToString(f.WorkingState.Hash())
	receipt.EmittedActionsHash = f.computeActionsHash(f.PendingActions)
	receipt.ExecutionTraceHash = f.computeTraceHash()

	return f.WorkingState, f.PendingActions, receipt, nil
}

func (f *ExecutionFrame) computeActionsHash(actions []Action) string {
	h := sha256.New()
	for _, a := range actions {
		h.Write([]byte(fmt.Sprintf("%d:%s:%v", a.Type, a.Target, a.Payload.Hash())))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (f *ExecutionFrame) computeTraceHash() string {
	h := sha256.New()
	for _, s := range f.Trace.Steps {
		h.Write([]byte(fmt.Sprintf("%s:%d:%d:%s", s.Instruction, s.StackDelta, s.GasConsumed, s.Phase)))
	}
	return hex.EncodeToString(h.Sum(nil))
}
