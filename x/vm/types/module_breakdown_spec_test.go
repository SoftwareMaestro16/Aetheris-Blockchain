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
	require.Len(t, registry.Modules, 6)

	byPath := map[AVMCosmosModulePath]AVMCosmosModuleBreakdown{}
	for _, module := range registry.Modules {
		require.NoError(t, module.Validate())
		byPath[module.ModulePath] = module
	}
	require.Contains(t, byPath, AVMModulePathActors)
	require.Contains(t, byPath, AVMModulePathAVM)
	require.Contains(t, byPath, AVMModulePathAVMContracts)
	require.Contains(t, byPath, AVMModulePathAVMInterfaces)
	require.Contains(t, byPath, AVMModulePathAsync)
	require.Contains(t, byPath, AVMModulePathContinuations)
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

func TestXActorsModuleBreakdownMatchesSection183(t *testing.T) {
	breakdown, err := DefaultXActorsModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, AVMModulePathActors, breakdown.ModulePath)

	require.Contains(t, breakdown.Purpose, "actor_records")
	require.Contains(t, breakdown.Purpose, "mailboxes")
	require.Contains(t, breakdown.Purpose, "actor_state")
	require.Contains(t, breakdown.Purpose, "permissions")
	require.Contains(t, breakdown.Purpose, "continuation_integration")

	require.ElementsMatch(t, []AVMModuleStateObject{
		AVMModuleStateActorRecord,
		AVMModuleStateActorMailbox,
		AVMModuleStateActorState,
		AVMModuleStateActorPermission,
	}, breakdown.StateObjects)
	require.ElementsMatch(t, []AVMModuleMessageName{
		AVMModuleMsgCreateActor,
		AVMModuleMsgSendActorMessage,
		AVMModuleMsgUpdateActor,
		AVMModuleMsgPauseActor,
	}, breakdown.Messages)
	require.ElementsMatch(t, []AVMModuleQueryName{
		AVMModuleQueryActor,
		AVMModuleQueryActorMailbox,
		AVMModuleQueryActorState,
	}, breakdown.Queries)
}

func TestXContinuationsModuleBreakdownMatchesSection184(t *testing.T) {
	breakdown, err := DefaultXContinuationsModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, AVMModulePathContinuations, breakdown.ModulePath)

	require.Contains(t, breakdown.Purpose, "async_workflow_state")
	require.Contains(t, breakdown.Purpose, "continuation_queues")
	require.Contains(t, breakdown.Purpose, "continuation_receipts")
	require.Contains(t, breakdown.Purpose, "resume")
	require.Contains(t, breakdown.Purpose, "expiry")

	require.ElementsMatch(t, []AVMModuleStateObject{
		AVMModuleStateContinuation,
		AVMModuleStateContinuationQueue,
		AVMModuleStateContinuationReceipt,
	}, breakdown.StateObjects)
	require.ElementsMatch(t, []AVMModuleMessageName{
		AVMModuleMsgCreateContinuation,
		AVMModuleMsgResumeContinuation,
		AVMModuleMsgCancelContinuation,
		AVMModuleMsgExpireContinuation,
	}, breakdown.Messages)
	require.ElementsMatch(t, []AVMModuleQueryName{
		AVMModuleQueryContinuation,
		AVMModuleQueryContinuationsByActor,
		AVMModuleQueryContinuationReceipt,
	}, breakdown.Queries)
}

func TestXAVMContractsModuleBreakdownMatchesSection185(t *testing.T) {
	breakdown, err := DefaultXAVMContractsModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, AVMModulePathAVMContracts, breakdown.ModulePath)

	require.Contains(t, breakdown.Purpose, "contract_code")
	require.Contains(t, breakdown.Purpose, "contract_instances")
	require.Contains(t, breakdown.Purpose, "contract_storage")
	require.Contains(t, breakdown.Purpose, "backend_adapters")

	require.ElementsMatch(t, []AVMModuleStateObject{
		AVMModuleStateCodeRecord,
		AVMModuleStateContractRecord,
		AVMModuleStateStorageValue,
		AVMModuleStateBackendConfig,
	}, breakdown.StateObjects)
	require.ElementsMatch(t, []AVMModuleMessageName{
		AVMModuleMsgStoreCode,
		AVMModuleMsgInstantiateContract,
		AVMModuleMsgExecuteContract,
		AVMModuleMsgMigrateContract,
	}, breakdown.Messages)
	require.ElementsMatch(t, []AVMModuleQueryName{
		AVMModuleQueryCode,
		AVMModuleQueryContract,
		AVMModuleQueryContractStorage,
		AVMModuleQueryContractProof,
	}, breakdown.Queries)
}

func TestXAVMInterfacesModuleBreakdownMatchesSection186(t *testing.T) {
	breakdown, err := DefaultXAVMInterfacesModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, AVMModulePathAVMInterfaces, breakdown.ModulePath)

	require.Contains(t, breakdown.Purpose, "contract_interface_schemas")
	require.Contains(t, breakdown.Purpose, "actor_interface_schemas")
	require.Contains(t, breakdown.Purpose, "service_interface_schemas")
	require.Contains(t, breakdown.Purpose, "interface_schema_registry")

	require.ElementsMatch(t, []AVMModuleStateObject{
		AVMModuleStateInterfaceDescriptor,
		AVMModuleStateMethodDescriptor,
		AVMModuleStateEventDescriptor,
		AVMModuleStateAsyncHandlerDescriptor,
	}, breakdown.StateObjects)
	require.ElementsMatch(t, []AVMModuleMessageName{
		AVMModuleMsgRegisterInterface,
		AVMModuleMsgUpdateInterface,
		AVMModuleMsgDeprecateInterface,
	}, breakdown.Messages)
	require.ElementsMatch(t, []AVMModuleQueryName{
		AVMModuleQueryInterface,
		AVMModuleQueryMethod,
		AVMModuleQueryInterfaceByTarget,
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

	actors, err := DefaultXActorsModuleBreakdown()
	require.NoError(t, err)
	actors.StateObjects = append(actors.StateObjects, AVMModuleStateContinuation)
	actors.BreakdownHash = ComputeAVMCosmosModuleBreakdownHash(actors)
	require.ErrorContains(t, actors.Validate(), "state object")

	continuations, err := DefaultXContinuationsModuleBreakdown()
	require.NoError(t, err)
	continuations.Messages = removeAVMModuleMessageForTest(continuations.Messages, AVMModuleMsgExpireContinuation)
	continuations.BreakdownHash = ComputeAVMCosmosModuleBreakdownHash(continuations)
	require.ErrorContains(t, continuations.Validate(), "message")

	contracts, err := DefaultXAVMContractsModuleBreakdown()
	require.NoError(t, err)
	contracts.StateObjects = append(contracts.StateObjects, AVMModuleStateInterfaceDescriptor)
	contracts.BreakdownHash = ComputeAVMCosmosModuleBreakdownHash(contracts)
	require.ErrorContains(t, contracts.Validate(), "state object")

	interfaces, err := DefaultXAVMInterfacesModuleBreakdown()
	require.NoError(t, err)
	interfaces.Messages = removeAVMModuleMessageForTest(interfaces.Messages, AVMModuleMsgDeprecateInterface)
	interfaces.BreakdownHash = ComputeAVMCosmosModuleBreakdownHash(interfaces)
	require.ErrorContains(t, interfaces.Validate(), "message")

	interfaces, err = DefaultXAVMInterfacesModuleBreakdown()
	require.NoError(t, err)
	interfaces.Queries = append(interfaces.Queries, AVMModuleQueryContractProof)
	interfaces.BreakdownHash = ComputeAVMCosmosModuleBreakdownHash(interfaces)
	require.ErrorContains(t, interfaces.Validate(), "query")
}

func TestAVMModuleRegistryRejectsMissingModuleAndHashMismatch(t *testing.T) {
	registry, err := DefaultAVMCosmosModuleRegistry()
	require.NoError(t, err)

	missing := registry
	missing.Modules = missing.Modules[:1]
	missing.RegistryHash = ComputeAVMCosmosModuleRegistryHash(missing)
	require.ErrorContains(t, missing.Validate(), "x/actors, x/avm, x/async, x/avmcontracts, x/avminterfaces, and x/continuations")

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
