package avm

import (
	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	"github.com/sovereign-l1/l1/x/aetravm/types"
)

const maxActionsPerExecution = 256

// Phase represents the current execution phase of the AVM.
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
}

func NewExecutionFrame(state *chunk.Chunk, msg Message) *ExecutionFrame {
	return &ExecutionFrame{
		Phase:         PhaseStorage,
		Message:       msg,
		StateSnapshot: state,
		WorkingState:  state,
		GasLimit:      msg.GasLimit,
		PhaseGas:      make(map[Phase]uint64),
		ActionBudget:  maxActionsPerExecution,
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
