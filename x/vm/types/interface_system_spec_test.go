package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAVMInterfaceDescriptorCommitsMethodsEventsAsyncHandlersAndGets(t *testing.T) {
	descriptor := testAVMInterfaceDescriptor(t)

	require.NoError(t, descriptor.Validate())
	require.Equal(t, ComputeAVMInterfaceHash(descriptor), descriptor.InterfaceHash)
	require.Equal(t, "v1.0.0", descriptor.InterfaceVersion)
	require.Equal(t, AVMInterfaceTargetActor, descriptor.TargetType)
	require.Equal(t, AVMInterfaceSchemaJSONSchema, descriptor.SchemaEncoding)
	require.Len(t, descriptor.MethodDescriptors, 2)
	require.Len(t, descriptor.EventDescriptors, 1)
	require.Len(t, descriptor.AsyncHandlerDescriptors, 1)
	require.Len(t, descriptor.GetMethodDescriptors, 1)
	require.Equal(t, "avm/interfaces/"+descriptor.InterfaceHash, AVMInterfaceDescriptorKey(descriptor.InterfaceHash))

	mutated := descriptor
	mutated.MethodDescriptors = append([]AVMMethodDescriptor(nil), descriptor.MethodDescriptors...)
	mutated.MethodDescriptors[0].GasHint++
	require.NotEqual(t, descriptor.InterfaceHash, ComputeAVMInterfaceHash(mutated))
}

func TestAVMMethodDescriptorSupportsExecutionModesAndOptionalRequirements(t *testing.T) {
	for _, mode := range []AVMInterfaceExecutionMode{
		AVMInterfaceExecutionSync,
		AVMInterfaceExecutionAsync,
		AVMInterfaceExecutionScheduled,
		AVMInterfaceExecutionGet,
	} {
		method := AVMMethodDescriptor{
			MethodID:                   "method-" + string(mode),
			Name:                       "method_" + string(mode),
			InputSchemaHash:            engineHash("input-" + string(mode)),
			OutputSchemaHash:           engineHash("output-" + string(mode)),
			ExecutionMode:              mode,
			GasHint:                    10,
			PaymentRequirementOptional: "naet",
			ProofRequirementOptional:   "state-proof",
		}
		require.NoError(t, method.Validate())
	}

	badMode := AVMMethodDescriptor{
		MethodID:         "bad",
		Name:             "bad",
		InputSchemaHash:  engineHash("input"),
		OutputSchemaHash: engineHash("output"),
		ExecutionMode:    AVMInterfaceExecutionMode("streaming"),
		GasHint:          1,
	}
	require.ErrorContains(t, badMode.Validate(), "execution mode")

	noGas := badMode
	noGas.ExecutionMode = AVMInterfaceExecutionSync
	noGas.GasHint = 0
	require.ErrorContains(t, noGas.Validate(), "gas hint")
}

func TestAVMInterfaceRegistryRootAndDuplicateRejection(t *testing.T) {
	descriptor := testAVMInterfaceDescriptor(t)
	second, err := NewAVMInterfaceDescriptor(AVMInterfaceDescriptor{
		InterfaceVersion: "v1.0.1",
		Owner:            "native-bank",
		TargetType:       AVMInterfaceTargetNativeModule,
		MethodDescriptors: []AVMMethodDescriptor{{
			MethodID:         "bank.send",
			Name:             "MsgSend",
			InputSchemaHash:  engineHash("bank-send-input"),
			OutputSchemaHash: engineHash("bank-send-output"),
			ExecutionMode:    AVMInterfaceExecutionSync,
			GasHint:          20,
		}},
		SchemaEncoding:       AVMInterfaceSchemaProtobuf,
		MetadataHashOptional: engineHash("bank-metadata"),
	})
	require.NoError(t, err)

	registry, err := NewAVMInterfaceRegistry(AVMInterfaceRegistry{Interfaces: []AVMInterfaceDescriptor{second, descriptor}})
	require.NoError(t, err)
	require.NoError(t, registry.Validate())
	require.Equal(t, ComputeAVMInterfaceRegistryRoot(registry), registry.Root)

	duplicate := registry
	duplicate.Interfaces = append([]AVMInterfaceDescriptor(nil), registry.Interfaces...)
	duplicate.Interfaces = append(duplicate.Interfaces, descriptor)
	duplicate.Root = ComputeAVMInterfaceRegistryRoot(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate")
}

func TestAVMInterfaceDescriptorRejectsMalformedDescriptorSets(t *testing.T) {
	descriptor := testAVMInterfaceDescriptor(t)

	empty := descriptor
	empty.MethodDescriptors = nil
	empty.EventDescriptors = nil
	empty.AsyncHandlerDescriptors = nil
	empty.GetMethodDescriptors = nil
	empty.InterfaceHash = ComputeAVMInterfaceHash(empty)
	require.ErrorContains(t, empty.Validate(), "at least one")

	duplicateMethod := descriptor
	duplicateMethod.MethodDescriptors = append([]AVMMethodDescriptor(nil), descriptor.MethodDescriptors...)
	duplicateMethod.MethodDescriptors = append(duplicateMethod.MethodDescriptors, descriptor.MethodDescriptors[0])
	duplicateMethod.InterfaceHash = ComputeAVMInterfaceHash(duplicateMethod)
	require.ErrorContains(t, duplicateMethod.Validate(), "duplicate AVM method")

	badMetadata := descriptor
	badMetadata.MetadataHashOptional = "not-a-hash"
	badMetadata.InterfaceHash = ComputeAVMInterfaceHash(badMetadata)
	require.ErrorContains(t, badMetadata.Validate(), "metadata")

	badEncoding := descriptor
	badEncoding.SchemaEncoding = AVMInterfaceSchemaEncoding("yaml")
	badEncoding.InterfaceHash = ComputeAVMInterfaceHash(badEncoding)
	require.ErrorContains(t, badEncoding.Validate(), "schema encoding")
}

func testAVMInterfaceDescriptor(t *testing.T) AVMInterfaceDescriptor {
	t.Helper()
	descriptor, err := NewAVMInterfaceDescriptor(AVMInterfaceDescriptor{
		InterfaceVersion: "v1.0.0",
		Owner:            "actor-contract-1",
		TargetType:       AVMInterfaceTargetActor,
		MethodDescriptors: []AVMMethodDescriptor{
			{
				MethodID:                   "actor.execute",
				Name:                       "execute",
				InputSchemaHash:            engineHash("execute-input"),
				OutputSchemaHash:           engineHash("execute-output"),
				ExecutionMode:              AVMInterfaceExecutionAsync,
				GasHint:                    100,
				PaymentRequirementOptional: "naet",
				ProofRequirementOptional:   "state-proof",
			},
			{
				MethodID:         "actor.schedule",
				Name:             "schedule",
				InputSchemaHash:  engineHash("schedule-input"),
				OutputSchemaHash: engineHash("schedule-output"),
				ExecutionMode:    AVMInterfaceExecutionScheduled,
				GasHint:          120,
			},
		},
		EventDescriptors: []AVMEventDescriptor{{
			EventID:    "actor.executed",
			Name:       "ActorExecuted",
			SchemaHash: engineHash("event-schema"),
		}},
		AsyncHandlerDescriptors: []AVMAsyncHandlerDescriptor{{
			HandlerID:           "actor.handle",
			Name:                "handle",
			InputSchemaHash:     engineHash("handler-input"),
			OutputSchemaHash:    engineHash("handler-output"),
			GasHint:             80,
			RetryPolicyOptional: "bounded",
		}},
		GetMethodDescriptors: []AVMGetMethodDescriptor{{
			MethodID:         "actor.balance",
			Name:             "balance",
			InputSchemaHash:  engineHash("get-input"),
			OutputSchemaHash: engineHash("get-output"),
			GasHint:          5,
		}},
		SchemaEncoding:       AVMInterfaceSchemaJSONSchema,
		MetadataHashOptional: engineHash("interface-metadata"),
	})
	require.NoError(t, err)
	return descriptor
}
