package types

const (
	ExitCodeOK uint32 = iota
	ExitCodeValidationFailed
	ExitCodeUnauthorized
	ExitCodeAccountInactive
	ExitCodeAccountFrozen
	ExitCodeContractFrozen
	ExitCodeCodeRejected
	ExitCodeOutOfGas
	ExitCodeStorageLimit
	ExitCodeStorageRentDebt
	ExitCodeMessageExpired
	ExitCodeQueueLimit
	ExitCodeExecutionFailed
	ExitCodeInternalBounce
	ExitCodeForbiddenHostCall
	ExitCodeInvalidJump
	ExitCodeCallStackOverflow
	ExitCodeContinuationNotFound
	ExitCodeRecursionLimitExceeded
	ExitCodeInvalidMemoryAccess
	ExitCodeNullReference
	ExitCodeInvalidChunkReference
	ExitCodeCorruptedStateObject
	ExitCodeDivisionByZero
	ExitCodeInvalidShift
	ExitCodeArithmeticUnderflow
	ExitCodeGasLimitExceeded
	ExitCodeGasReservationFailed
	ExitCodeExecutionTimeout
	ExitCodeStackOverflow
	ExitCodeStackUnderflow
	ExitCodeTypeCheckError
	ExitCodeMessageRoutingFailed
	ExitCodeQueueOverflow
	ExitCodeShardUnavailable
	ExitCodeInsufficientBalance
	ExitCodeInsufficientGas
	ExitCodeStateCorruption
	ExitCodeStateVersionMismatch
	ExitCodeSnapshotFailure
	ExitCodeExplicitAbort
	ExitCodeAssertionFailed
	ExitCodeAccountStateTooBig
	ExitCodeInactiveFrozen
	ExitCodeContractAbort
)

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
	case ExitCodeOutOfGas:
		return "out_of_gas"
	case ExitCodeStorageLimit:
		return "storage_limit"
	case ExitCodeStorageRentDebt:
		return "storage_rent_debt"
	case ExitCodeMessageExpired:
		return "message_expired"
	case ExitCodeQueueLimit:
		return "queue_limit"
	case ExitCodeExecutionFailed:
		return "execution_failed"
	case ExitCodeInternalBounce:
		return "internal_bounce"
	case ExitCodeForbiddenHostCall:
		return "forbidden_host_call"
	case ExitCodeInvalidJump:
		return "invalid_jump"
	case ExitCodeCallStackOverflow:
		return "call_stack_overflow"
	case ExitCodeContinuationNotFound:
		return "continuation_not_found"
	case ExitCodeRecursionLimitExceeded:
		return "recursion_limit_exceeded"
	case ExitCodeInvalidMemoryAccess:
		return "invalid_memory_access"
	case ExitCodeNullReference:
		return "null_reference"
	case ExitCodeInvalidChunkReference:
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
	case ExitCodeStackOverflow:
		return "stack_overflow"
	case ExitCodeStackUnderflow:
		return "stack_underflow"
	case ExitCodeTypeCheckError:
		return "type_check_error"
	case ExitCodeMessageRoutingFailed:
		return "message_routing_failed"
	case ExitCodeQueueOverflow:
		return "queue_overflow"
	case ExitCodeShardUnavailable:
		return "shard_unavailable"
	case ExitCodeInsufficientBalance:
		return "insufficient_balance"
	case ExitCodeInsufficientGas:
		return "insufficient_gas"
	case ExitCodeStateCorruption:
		return "state_corruption"
	case ExitCodeStateVersionMismatch:
		return "state_version_mismatch"
	case ExitCodeSnapshotFailure:
		return "snapshot_failure"
	case ExitCodeExplicitAbort:
		return "explicit_abort"
	case ExitCodeAssertionFailed:
		return "assertion_failed"
	case ExitCodeAccountStateTooBig:
		return "account_state_too_big"
	case ExitCodeInactiveFrozen:
		return "inactive_frozen"
	case ExitCodeContractAbort:
		return "contract_abort"
	default:
		return "unknown"
	}
}
