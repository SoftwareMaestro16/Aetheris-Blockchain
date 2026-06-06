package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAVMCosmosModuleRegistryCoversSection18(t *testing.T) {
	registry, err := DefaultAVMCosmosModuleRegistry()
	require.NoError(t, err)
	require.NoError(t, registry.Validate())
	require.Equal(t, ComputeAVMCosmosModuleRegistryHash(registry), registry.RegistryHash)
	require.Len(t, registry.Modules, 2)

	byPath := map[AVMCosmosModulePath]AVMCosmosModuleBreakdown{}
	for _, module := range registry.Modules {
		require.NoError(t, module.Validate())
		byPath[module.ModulePath] = module
	}
	require.Contains(t, byPath, AVMModulePathAVM)
	require.Contains(t, byPath, AVMModulePathAsync)
}

func TestXAVMModuleBreakdownMatchesSection181(t *testing.T) {
	breakdown, err := DefaultXAVMModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, AVMModulePathAVM, breakdown.ModulePath)

	require.Contains(t, breakdown.Purpose, "runtime_parameters")
	require.Contains(t, breakdown.Purpose, "routing")
	require.Contains(t, breakdown.Purpose, "roots")
	require.Contains(t, breakdown.Purpose, "execution_receipts")
	require.Contains(t, breakdown.Purpose, "runtime_versions")

	require.ElementsMatch(t, []AVMModuleStateObject{
		AVMModuleStateAVMParams,
		AVMModuleStateRouteDescriptor,
		AVMModuleStateAVMRoot,
		AVMModuleStateExecutionReceipt,
		AVMModuleStateRuntimeVersion,
	}, breakdown.StateObjects)
	require.ElementsMatch(t, []AVMModuleMessageName{
		AVMModuleMsgSubmitAVMMessage,
		AVMModuleMsgRegisterRoute,
		AVMModuleMsgUpdateAVMParams,
		AVMModuleMsgScheduleRuntimeUpgrade,
	}, breakdown.Messages)
	require.ElementsMatch(t, []AVMModuleQueryName{
		AVMModuleQueryAVMParams,
		AVMModuleQueryAVMRoot,
		AVMModuleQueryRoute,
		AVMModuleQueryExecutionReceipt,
		AVMModuleQueryRuntimeVersion,
	}, breakdown.Queries)
}

func TestXAsyncModuleBreakdownMatchesSection182(t *testing.T) {
	breakdown, err := DefaultXAsyncModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, AVMModulePathAsync, breakdown.ModulePath)

	require.Contains(t, breakdown.Purpose, "async_message_queues")
	require.Contains(t, breakdown.Purpose, "retry_queue")
	require.Contains(t, breakdown.Purpose, "delayed_queue")
	require.Contains(t, breakdown.Purpose, "dead_letter_queue")
	require.Contains(t, breakdown.Purpose, "replay_tombstones")

	require.ElementsMatch(t, []AVMModuleStateObject{
		AVMModuleStateAsyncMessage,
		AVMModuleStateZoneQueue,
		AVMModuleStateRetryRecord,
		AVMModuleStateDeadLetter,
		AVMModuleStateReplayTombstone,
	}, breakdown.StateObjects)
	require.ElementsMatch(t, []AVMModuleMessageName{
		AVMModuleMsgSubmitAsyncMessage,
		AVMModuleMsgCancelAsyncMessage,
		AVMModuleMsgRetryAsyncMessage,
		AVMModuleMsgExpireAsyncMessage,
	}, breakdown.Messages)
	require.ElementsMatch(t, []AVMModuleQueryName{
		AVMModuleQueryAsyncMessage,
		AVMModuleQueryZoneQueue,
		AVMModuleQueryDeadLetter,
		AVMModuleQueryReplayTombstone,
	}, breakdown.Queries)
}

func TestAVMModuleBreakdownRejectsMissingAndCrossOwnedSurface(t *testing.T) {
	breakdown, err := DefaultXAVMModuleBreakdown()
	require.NoError(t, err)
	breakdown.Messages = removeAVMModuleMessageForTest(breakdown.Messages, AVMModuleMsgRegisterRoute)
	breakdown.BreakdownHash = ComputeAVMCosmosModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "message")

	breakdown, err = DefaultXAVMModuleBreakdown()
	require.NoError(t, err)
	breakdown.StateObjects = append(breakdown.StateObjects, AVMModuleStateZoneQueue)
	breakdown.BreakdownHash = ComputeAVMCosmosModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "state object")

	async, err := DefaultXAsyncModuleBreakdown()
	require.NoError(t, err)
	async.Queries = removeAVMModuleQueryForTest(async.Queries, AVMModuleQueryReplayTombstone)
	async.BreakdownHash = ComputeAVMCosmosModuleBreakdownHash(async)
	require.ErrorContains(t, async.Validate(), "query")
}

func TestAVMModuleRegistryRejectsMissingModuleAndHashMismatch(t *testing.T) {
	registry, err := DefaultAVMCosmosModuleRegistry()
	require.NoError(t, err)

	missing := registry
	missing.Modules = missing.Modules[:1]
	missing.RegistryHash = ComputeAVMCosmosModuleRegistryHash(missing)
	require.ErrorContains(t, missing.Validate(), "x/avm and x/async")

	mutated := registry
	mutated.Modules[0].Purpose[0] = "changed"
	require.ErrorContains(t, mutated.Validate(), "breakdown hash mismatch")
}

func removeAVMModuleMessageForTest(messages []AVMModuleMessageName, target AVMModuleMessageName) []AVMModuleMessageName {
	out := make([]AVMModuleMessageName, 0, len(messages))
	for _, message := range messages {
		if message != target {
			out = append(out, message)
		}
	}
	return out
}

func removeAVMModuleQueryForTest(queries []AVMModuleQueryName, target AVMModuleQueryName) []AVMModuleQueryName {
	out := make([]AVMModuleQueryName, 0, len(queries))
	for _, query := range queries {
		if query != target {
			out = append(out, query)
		}
	}
	return out
}
