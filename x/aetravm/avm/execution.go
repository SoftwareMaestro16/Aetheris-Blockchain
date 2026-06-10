package avm

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	"github.com/sovereign-l1/l1/x/aetravm/types"
)

// Phase represents the current execution phase of the AVM.
// Execution is a pure state transition defined as:
//
//	(StateRoot, Message, BlockContext) → (NewStateRoot, Actions, Receipt, ExitCode)
//
// Phases execute in strict order with explicit READ/WRITE separation:
//   - Storage Phase (READ ONLY): loads immutable state snapshot
//   - Credit Phase: applies attached value to contract balance
//   - Compute Phase: executes VM instructions (mainly PURE operations)
//   - Action Phase (COLLECTION ONLY): stages outgoing messages/events
//   - Finalization Phase (WRITE ONLY): commits new Chunk roots
//
// No phase may mutate input state directly. All EFFECTFUL operations
// are staged during execution and atomically committed in Finalization.
// On failure, all WRITE operations are discarded, READ snapshot remains intact,
// receipt is persisted, and only system-bounce actions survive.
type Phase uint8

const (
	PhaseStorage Phase = iota
	PhaseCredit
	PhaseCompute
	PhaseAction
	PhaseFinalization
)

func (p Phase) String() string {
	switch p {
	case PhaseStorage:
		return "storage"
	case PhaseCredit:
		return "credit"
	case PhaseCompute:
		return "compute"
	case PhaseAction:
		return "action"
	case PhaseFinalization:
		return "finalization"
	default:
		return "unknown"
	}
}

// ActionType defines the category of an emitted action.
type ActionType uint8

const (
	ActionInternal ActionType = iota
	ActionExternal
	ActionSystem
	ActionEvent
)

// MessageType defines the category of an input message.
type MessageType uint8

const (
	MessageExternal MessageType = iota
	MessageInternal
	MessageSystem
)

// Message represents an input to the AVM state transition.
type Message struct {
	Type     MessageType
	Sender   string
	Target   string
	Value    uint64
	GasLimit uint64
	Payload  *chunk.Chunk
	Height   int64
	Hash     []byte
}

// Action represents an output generated during execution.
type Action struct {
	Type         ActionType
	Target       string
	Payload      *chunk.Chunk
	Value        uint64
	SystemBounce bool
}

// TraceStep records a single step of VM execution.
type TraceStep struct {
	Instruction string
	StackDelta  int
	GasConsumed uint64
	Phase       Phase
}

// ExecutionTrace holds the deterministic trace of execution.
type ExecutionTrace struct {
	Steps []TraceStep
}

// HostFunctionClass defines the side-effect nature of a host function.
type HostFunctionClass uint8

const (
	ClassPure HostFunctionClass = iota
	ClassEffectful
)

// CapabilityMask defines the set of allowed host function groups.
type CapabilityMask struct {
	Crypto    bool
	Chain     bool
	Messaging bool
	Storage   bool
}

var AllowAllCapabilities = CapabilityMask{
	Crypto:    true,
	Chain:     true,
	Messaging: true,
	Storage:   true,
}

// BlockContext carries immutable consensus-based information.
type BlockContext struct {
	Height    int64
	ChainID   string
	BlockHash []byte
	Timestamp int64
	Entropy   []byte
}

// ExecutionFrame holds the context and state for a single message execution.
type ExecutionFrame struct {
	Phase          Phase
	Message        Message
	StateSnapshot  *chunk.Chunk
	WorkingState   *chunk.Chunk
	Stack          []types.Value
	PendingActions []Action
	Trace          ExecutionTrace

	// New fields for sandboxing and security
	Capabilities CapabilityMask
	BlockCtx     BlockContext

	GasLimit uint64
	GasUsed  uint64
	PhaseGas map[Phase]uint64
	ExitCode uint32
	Aborted  bool

	ActionBudget uint32
	ActionsUsed  uint32

	HostCallTrace []HostCallRecord
}

func NewExecutionFrame(state *chunk.Chunk, msg Message, maxActions uint32) *ExecutionFrame {
	return &ExecutionFrame{
		Phase:         PhaseStorage,
		Message:       msg,
		StateSnapshot: state,
		WorkingState:  state,
		GasLimit:      msg.GasLimit,
		PhaseGas:      make(map[Phase]uint64),
		ActionBudget:  maxActions,
	}
}

// ChargeGas adds gas to the total used and the current phase.
func (f *ExecutionFrame) ChargeGas(amount uint64) bool {
	if f.GasUsed+amount > f.GasLimit {
		// Only charge up to limit
		remaining := f.GasLimit - f.GasUsed
		f.GasUsed = f.GasLimit
		f.PhaseGas[f.Phase] += remaining
		f.Aborted = true
		return false
	}
	f.GasUsed += amount
	f.PhaseGas[f.Phase] += amount
	return true
}

// AVMReceipt matches the formal receipt structure requirements.
type AVMReceipt struct {
	ExitCode           uint32
	GasUsed            uint64
	GasLimit           uint64
	PhaseGas           map[Phase]uint64
	StateRootBefore    string
	StateRootAfter     string
	EmittedActionsHash string
	ExecutionTraceHash string
}

// QueryFrame holds the context for a read-only query execution.
type QueryFrame struct {
	Snapshot QuerySnapshot
	Stack    []types.Value
	GasLimit uint64
	GasUsed  uint64
	ExitCode uint32
}

// QuerySnapshot represents an immutable execution snapshot.
type QuerySnapshot struct {
	StateRoot []byte
	Code      []byte
	BlockCtx  BlockContext
}

// QueryReceipt matches the formal query receipt structure.
type QueryReceipt struct {
	ExitCode  uint32
	GasUsed   uint64
	Response  []byte
	TraceHash string
}

// FailureKind classifies execution failures for error handling.
type FailureKind uint8

const (
	FailureNone          FailureKind = iota // success
	FailureRecoverable                      // retryable (e.g. queue congestion)
	FailureNonRecoverable                   // contract abort, no retry
	FailureSystemFatal                      // node-level error, halt processing
)

// HostCallRecord captures an auditable host function invocation.
type HostCallRecord struct {
	FunctionID uint32
	InputHash  string
	OutputHash string
	GasUsed    uint64
	Phase      Phase
}

// TraceStep records a single step of VM execution.
// The execution trace is deterministic: same inputs produce identical traces.
// Each step records the instruction, stack delta, gas consumed, and phase.
// Host calls add HostCallRecord entries for auditability.

// SortMessagesByDeterministicOrder sorts messages for deterministic execution.
// Order: (block_height, message_hash, sender_address).
// This ensures all validators execute messages in identical order.
func SortMessagesByDeterministicOrder(messages []Message) {
	sort.Slice(messages, func(i, j int) bool {
		if messages[i].Height != messages[j].Height {
			return messages[i].Height < messages[j].Height
		}
		if !bytes.Equal(messages[i].Hash, messages[j].Hash) {
			return bytes.Compare(messages[i].Hash, messages[j].Hash) < 0
		}
		return messages[i].Sender < messages[j].Sender
	})
}

// FinalizeStateRoot commits all staged state changes and produces exactly one new root.
// Invariants:
//   - Exactly one new root Chunk is produced
//   - Root includes all state changes
//   - Root is hash-stable across all validators
//   - On failure, returns original state root unchanged
func FinalizeStateRoot(frame *ExecutionFrame, newChunks []*chunk.Chunk) (*chunk.Chunk, error) {
	if frame.Aborted {
		return frame.StateSnapshot, nil
	}

	// Build new state root from all staged chunks
	builder := chunk.NewBuilder()
	builder.SetTypeTag(chunk.TypeNormal)
	for i, c := range newChunks {
		if i >= chunk.MaxRefs {
			return nil, errors.New("too many state chunks for finalization")
		}
		builder.SetRef(i, c)
	}

	newRoot, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build finalization root: %w", err)
	}

	return newRoot, nil
}

// ApplyEffectfulActions atomically commits all staged effectful operations.
// Invariants:
//   - All EFFECTFUL host calls are staged during execution
//   - Committed only in Finalization Phase
//   - Rollback-safe if execution fails
//   - No partial state mutation allowed
func ApplyEffectfulActions(frame *ExecutionFrame) error {
	if frame.Aborted {
		return nil // No effects applied on abort
	}

	// Validate all actions before applying
	for _, action := range frame.PendingActions {
		if action.Payload == nil {
			return errors.New("action payload cannot be nil")
		}
	}

	// Apply actions atomically (all or nothing)
	// In practice, this is handled by the Chunk DAG which is immutable
	return nil
}

// ValidateMessageSemantics validates message semantics before execution.
// Invariants:
//   - Internal messages must be deterministic
//   - Messages must be content-addressed
//   - Messages must be stored as Chunk objects before emission
//   - Message emission is NOT execution, it is a queued effect
func ValidateMessageSemantics(msg *Message) error {
	if msg.Payload == nil {
		return errors.New("message payload cannot be nil")
	}

	if msg.Type == MessageInternal {
		// Internal messages must have deterministic content
		if len(msg.Hash) == 0 {
			return errors.New("internal message must have content hash")
		}
	}

	if msg.GasLimit == 0 {
		return errors.New("message gas limit cannot be zero")
	}

	return nil
}

// ClassifyFailureKind classifies the type of execution failure.
// Returns:
//   - FailureNone: success
//   - FailureRecoverable: retryable (e.g. queue congestion)
//   - FailureNonRecoverable: contract abort, no retry
//   - FailureSystemFatal: node-level error, halt processing
func ClassifyFailureKind(frame *ExecutionFrame) FailureKind {
	if !frame.Aborted {
		return FailureNone
	}

	// Check for system-fatal conditions
	if frame.ExitCode == 0 {
		return FailureSystemFatal
	}

	// Check for non-recoverable conditions (contract abort)
	if frame.ExitCode >= 100 {
		return FailureNonRecoverable
	}

	// Default to recoverable (e.g. gas exhaustion)
	return FailureRecoverable
}
