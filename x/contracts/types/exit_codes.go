package types

// Exit code domains:
// 0-31   = VM execution errors
// 32-63  = action/message errors
// 64-95  = state/storage errors
// 96-127 = system/host errors

const (
	// VM Execution Errors (0-31)
	ExitCodeOK                     uint32 = iota
	ExitCodeValidationFailed              // 1
	ExitCodeUnauthorized                  // 2
	ExitCodeAccountInactive               // 3
	ExitCodeAccountFrozen                 // 4
	ExitCodeContractFrozen                // 5
	ExitCodeCodeRejected                  // 6
	ExitCodeTypeCheckFailed               // 7
	ExitCodeInvalidJump                   // 14 = invalid jump / branch
	ExitCodeCallStackOverflow             // 15 = call stack depth exceeded
	ExitCodeContinuationMissing           // 16 = continuation not found
	ExitCodeRecursionLimitExceeded        // 17 = recursive call limit exceeded

	// Memory / Type Safety Errors (18-21)
	ExitCodeInvalidMemoryAccess  // 18 = invalid memory access
	ExitCodeNullReference        // 19 = null reference / empty slice read
	ExitCodeInvalidChunkRef      // 20 = invalid chunk reference (content-addressing)
	ExitCodeCorruptedStateObject // 21 = corrupted state object

	// Arithmetic Edge Cases (22-24)
	ExitCodeDivisionByZero      // 22 = division by zero
	ExitCodeInvalidShift        // 23 = negative shift / invalid shift
	ExitCodeArithmeticUnderflow // 24 = arithmetic underflow (unsigned)

	// Execution Safety / Gas Edge Cases (25-27)
	ExitCodeGasLimitExceeded     // 25 = gas limit exceeded
	ExitCodeGasReservationFailed // 26 = gas reservation failed
	ExitCodeExecutionTimeout     // 27 = execution timeout

	// Action / Message Errors (32-63)
	ExitCodeMessageExpired   // 32 = message expired
	ExitCodeQueueLimit       // 33 = queue overflow
	ExitCodeMessageTooLarge  // 34 = message too large
	ExitCodeRoutingFailed    // 38 = message routing failure
	ExitCodeQueueOverflow    // 39 = queue overflow (full)
	ExitCodeShardUnavailable // 40 = shard unavailable / routing failure

	// State / Storage Errors (64-95)
	ExitCodeStorageLimit         // 64 = storage limit
	ExitCodeStorageRentDebt      // 65 = storage rent debt
	ExitCodeAccountStateTooBig   // 66 = account state too big
	ExitCodeStateCorruption      // 67 = state corruption
	ExitCodeStateVersionMismatch // 68 = state version mismatch
	ExitCodeSnapshotFailure      // 69 = snapshot failure

	// System / Host Errors (96-127)
	ExitCodeExecutionFailed     // 96 = execution failed
	ExitCodeInternalBounce      // 97 = internal bounce
	ExitCodeForbiddenHostCall   // 98 = forbidden host call
	ExitCodeContractAbort       // 99 = explicit contract abort
	ExitCodeAssertionFailed     // 100 = assertion failed
	ExitCodeInsufficientBalance // 101 = insufficient AET balance for gas
)

// IsVMExecutionError returns true if code is in VM execution domain (0-31)
func IsVMExecutionError(code uint32) bool {
	return code <= 31
}

// IsActionMessageError returns true if code is in action/message domain (32-63)
func IsActionMessageError(code uint32) bool {
	return code >= 32 && code <= 63
}

// IsStateStorageError returns true if code is in state/storage domain (64-95)
func IsStateStorageError(code uint32) bool {
	return code >= 64 && code <= 95
}

// IsSystemHostError returns true if code is in system/host domain (96+)
func IsSystemHostError(code uint32) bool {
	return code >= 96
}

func ExitCodeName(code uint32) string {
	switch code {
	case ExitCodeOK:
		return "ok"
	case ExitCodeValidationFailed:
		return "validation_failed"
	case ExitCodeUnauthorized:
		return "unauthorized"
	case ExitCodeAccountInactive:
		return "account_inactive"
	case ExitCodeAccountFrozen:
		return "account_frozen"
	case ExitCodeContractFrozen:
		return "contract_frozen"
	case ExitCodeCodeRejected:
		return "code_rejected"
	case ExitCodeTypeCheckFailed:
		return "type_check_failed"
	case ExitCodeInvalidJump:
		return "invalid_jump"
	case ExitCodeCallStackOverflow:
		return "call_stack_overflow"
	case ExitCodeContinuationMissing:
		return "continuation_missing"
	case ExitCodeRecursionLimitExceeded:
		return "recursion_limit_exceeded"
	case ExitCodeInvalidMemoryAccess:
		return "invalid_memory_access"
	case ExitCodeNullReference:
		return "null_reference"
	case ExitCodeInvalidChunkRef:
		return "invalid_chunk_reference"
	case ExitCodeCorruptedStateObject:
		return "corrupted_state_object"
	case ExitCodeDivisionByZero:
		return "division_by_zero"
	case ExitCodeInvalidShift:
		return "invalid_shift"
	case ExitCodeArithmeticUnderflow:
		return "arithmetic_underflow"
	case ExitCodeGasLimitExceeded:
		return "gas_limit_exceeded"
	case ExitCodeGasReservationFailed:
		return "gas_reservation_failed"
	case ExitCodeExecutionTimeout:
		return "execution_timeout"
	case ExitCodeMessageExpired:
		return "message_expired"
	case ExitCodeQueueLimit:
		return "queue_limit"
	case ExitCodeMessageTooLarge:
		return "message_too_large"
	case ExitCodeRoutingFailed:
		return "routing_failed"
	case ExitCodeQueueOverflow:
		return "queue_overflow"
	case ExitCodeShardUnavailable:
		return "shard_unavailable"
	case ExitCodeStorageLimit:
		return "storage_limit"
	case ExitCodeStorageRentDebt:
		return "storage_rent_debt"
	case ExitCodeAccountStateTooBig:
		return "account_state_too_big"
	case ExitCodeStateCorruption:
		return "state_corruption"
	case ExitCodeStateVersionMismatch:
		return "state_version_mismatch"
	case ExitCodeSnapshotFailure:
		return "snapshot_failure"
	case ExitCodeExecutionFailed:
		return "execution_failed"
	case ExitCodeInternalBounce:
		return "internal_bounce"
	case ExitCodeForbiddenHostCall:
		return "forbidden_host_call"
	case ExitCodeContractAbort:
		return "contract_abort"
	case ExitCodeAssertionFailed:
		return "assertion_failed"
	case ExitCodeInsufficientBalance:
		return "insufficient_balance"
	default:
		return "unknown"
	}
}
