package async

import (
	"testing"

	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	"github.com/stretchr/testify/require"
)

func TestRuntimeExitCodesHaveCanonicalContractMapping(t *testing.T) {
	seen := map[uint32]struct{}{}
	for _, spec := range RuntimeExitCodes() {
		require.Less(t, spec.ContractExitCode, uint32(128))
		require.NotEqual(t, "unknown", RuntimeExitCodeName(spec.Code))
		require.True(t, contractstypes.KnownExitCode(spec.ContractExitCode))
		require.Equal(t, contractstypes.ExitCodeName(spec.ContractExitCode), spec.ContractExitCodeName)
		_, duplicate := seen[spec.Code]
		require.False(t, duplicate, "duplicate runtime exit code %d", spec.Code)
		seen[spec.Code] = struct{}{}
	}
}

func TestRuntimeLimitMappingUsesFailurePhase(t *testing.T) {
	require.Equal(t, uint32(contractstypes.ExitCodeOutOfGas), ContractExitCodeForRuntime(ResultLimitExceeded, FailedPhaseExecution))
	require.Equal(t, "out_of_gas", ContractExitCodeNameForRuntime(ResultLimitExceeded, FailedPhaseExecution))
	require.Equal(t, uint32(contractstypes.ExitCodeStorageLimit), ContractExitCodeForRuntime(ResultLimitExceeded, FailedPhaseStorage))
	require.Equal(t, uint32(contractstypes.ExitCodeQueueLimit), ContractExitCodeForRuntime(ResultLimitExceeded, FailedPhaseQueue))
	require.Equal(t, uint32(contractstypes.ExitCodeValidationFailed), ContractExitCodeForRuntime(ResultLimitExceeded, FailedPhaseValidation))
}

func TestRuntimeSpecialMappings(t *testing.T) {
	require.Equal(t, uint32(contractstypes.ExitCodeMessageExpired), ContractExitCodeForRuntime(ResultExpired, FailedPhaseQueue))
	require.Equal(t, uint32(contractstypes.ExitCodeInternalBounce), ContractExitCodeForRuntime(ResultBounceSuppressed, FailedPhaseQueue))
	require.Equal(t, uint32(contractstypes.ExitCodeForbiddenHostCall), ContractExitCodeForRuntime(ResultForbiddenHostCall, FailedPhaseExecution))
	require.Equal(t, uint32(contractstypes.ExitCodeContractAbort), ContractExitCodeForRuntime(142, FailedPhaseExecution))
	require.Equal(t, "unknown", RuntimeExitCodeName(142))
}
