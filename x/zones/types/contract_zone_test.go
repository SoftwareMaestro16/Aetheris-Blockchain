package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContractZoneBoundaryAndSpecStateKeys(t *testing.T) {
	boundary := DefaultContractZoneBoundary()
	require.NoError(t, boundary.Validate())
	require.Contains(t, boundary.Messages, ContractMessageStoreCode)
	require.Contains(t, boundary.Messages, ContractMessageInstantiate)
	require.Contains(t, boundary.Messages, ContractMessageExecute)
	require.Contains(t, boundary.Messages, ContractMessageMigrate)
	require.Contains(t, boundary.Messages, ContractMessageCallback)
	require.Contains(t, boundary.Messages, ContractMessageProofVerify)
	require.Contains(t, boundary.ProofKinds, ContractProofCode)
	require.Contains(t, boundary.ProofKinds, ContractProofContract)
	require.Contains(t, boundary.ProofKinds, ContractProofState)
	require.Contains(t, boundary.ProofKinds, ContractProofABI)
	require.Contains(t, boundary.ProofKinds, ContractProofReceipt)

	codeKey, err := ContractCodeKey("code-1")
	require.NoError(t, err)
	require.Equal(t, "contract/code/code-1", codeKey)

	instanceKey, err := ContractInstanceKey("contract-1")
	require.NoError(t, err)
	require.Equal(t, "contract/instance/contract-1", instanceKey)

	storageKey, err := ContractStorageKey("contract-1", "balance")
	require.NoError(t, err)
	require.Equal(t, "contract/storage/contract-1/balance", storageKey)

	abiKey, err := ContractABIKey("code-1")
	require.NoError(t, err)
	require.Equal(t, "contract/abi/code-1", abiKey)

	inboxKey, err := ContractInboxKey("contract-1", "msg-1")
	require.NoError(t, err)
	require.Equal(t, "contract/inbox/contract-1/msg-1", inboxKey)

	receiptKey, err := ContractReceiptKey("contract-1", "receipt-1")
	require.NoError(t, err)
	require.Equal(t, "contract/receipts/contract-1/receipt-1", receiptKey)
}

func TestContractBytecodeAndCosmWasmBoundaryCommitments(t *testing.T) {
	iface, err := NewContractBytecodeInterface(ContractBytecodeInterface{
		Runtime:         ContractRuntimeAVM,
		InstructionSet:  "avm-v1",
		BytecodeHash:    hash("bytecode"),
		ABIHash:         hash("abi"),
		DeterminismHash: hash("determinism"),
		MaxCodeBytes:    1 << 20,
	})
	require.NoError(t, err)
	require.Equal(t, ComputeContractBytecodeInterfaceHash(iface), iface.InterfaceHash)
	require.NoError(t, iface.Validate())

	adapter, err := NewContractCosmWasmAdapterDescriptor(ContractCosmWasmAdapterDescriptor{
		AdapterID:      "cw-adapter",
		Version:        "v1",
		PolicyHash:     hash("adapter-policy"),
		CapabilityRoot: hash("adapter-capability"),
	})
	require.NoError(t, err)
	require.Equal(t, ComputeContractCosmWasmAdapterHash(adapter), adapter.DescriptorHash)
	require.NoError(t, adapter.Validate())

	rootA := ComputeContractBytecodeInterfaceRoot([]ContractBytecodeInterface{iface})
	rootB := ComputeContractBytecodeInterfaceRoot([]ContractBytecodeInterface{iface})
	require.Equal(t, rootA, rootB)
	require.NotEmpty(t, ComputeContractCosmWasmAdapterRoot([]ContractCosmWasmAdapterDescriptor{adapter}))
}

func TestContractStorageInboxAndReceiptsAreDeterministic(t *testing.T) {
	storage, err := UpsertContractStorage(nil, contractStorage("contract-b", "slot-2", "value-2"))
	require.NoError(t, err)
	storage, err = UpsertContractStorage(storage, contractStorage("contract-a", "slot-1", "value-1"))
	require.NoError(t, err)
	storage, err = UpsertContractStorage(storage, contractStorage("contract-b", "slot-2", "value-2b"))
	require.NoError(t, err)
	require.Equal(t, "contract-a", storage[0].ContractAddr)
	require.Len(t, storage, 2)

	inbox, err := EnqueueContractInbox(nil, contractInbox("contract-1", "msg-2", 2))
	require.NoError(t, err)
	inbox, err = EnqueueContractInbox(inbox, contractInbox("contract-1", "msg-1", 1))
	require.NoError(t, err)
	require.Equal(t, "msg-1", inbox[0].MsgID)
	_, err = EnqueueContractInbox(inbox, contractInbox("contract-1", "msg-dup", 1))
	require.ErrorContains(t, err, "sequence already exists")

	receipt, err := NewContractExecutionReceipt(contractReceipt("contract-1", "receipt-1", "msg-1", 1))
	require.NoError(t, err)
	zoneReceipt, err := receipt.ZoneReceipt()
	require.NoError(t, err)
	require.Equal(t, ZoneIDContract, zoneReceipt.ZoneID)
	require.Equal(t, ZoneReceiptStatusSuccess, zoneReceipt.Status)

	rootA := ComputeContractStorageRoot(storage)
	rootB := ComputeContractStorageRoot([]ContractStorageEntry{storage[1], storage[0]})
	require.Equal(t, rootA, rootB)
	require.Equal(t, ComputeContractInboxRoot(inbox), ComputeContractInboxRoot([]ContractInboxMessage{inbox[1], inbox[0]}))
}

func TestContractZoneStateRootAndProofRequest(t *testing.T) {
	iface, err := NewContractBytecodeInterface(ContractBytecodeInterface{
		Runtime:         ContractRuntimeAVM,
		InstructionSet:  "avm-v1",
		BytecodeHash:    hash("bytecode"),
		ABIHash:         hash("abi"),
		DeterminismHash: hash("determinism"),
		MaxCodeBytes:    1 << 20,
	})
	require.NoError(t, err)
	adapter, err := NewContractCosmWasmAdapterDescriptor(ContractCosmWasmAdapterDescriptor{
		AdapterID:      "cw-adapter",
		Version:        "v1",
		PolicyHash:     hash("adapter-policy"),
		CapabilityRoot: hash("adapter-capability"),
	})
	require.NoError(t, err)
	receipt, err := NewContractExecutionReceipt(contractReceipt("contract-1", "receipt-1", "msg-1", 1))
	require.NoError(t, err)

	state := ContractZoneState{
		Height: 77,
		Codes: []ContractCode{{
			CodeID:         "code-1",
			Runtime:        ContractRuntimeAVM,
			BytecodeHash:   hash("bytecode"),
			BytecodeSize:   4096,
			ABIHash:        hash("abi"),
			InterfaceHash:  iface.InterfaceHash,
			Uploader:       "alice",
			UploadedHeight: 77,
		}},
		Instances: []ContractInstance{{
			ContractAddr:  "contract-1",
			CodeID:        "code-1",
			Runtime:       ContractRuntimeAVM,
			Admin:         "alice",
			StorageRoot:   hash("storage"),
			CreatedHeight: 77,
			UpdatedHeight: 77,
		}},
		Storage:            []ContractStorageEntry{contractStorage("contract-1", "slot-1", "value-1")},
		ABIs:               []ContractABIDescriptor{{CodeID: "code-1", ABIHash: hash("abi"), InterfaceHash: iface.InterfaceHash, ExportedMethods: []string{"execute", "query"}, RegisteredHeight: 77}},
		Inbox:              []ContractInboxMessage{contractInbox("contract-1", "msg-1", 1)},
		Receipts:           []ContractExecutionReceipt{receipt},
		BytecodeInterfaces: []ContractBytecodeInterface{iface},
		CosmWasmAdapters:   []ContractCosmWasmAdapterDescriptor{adapter},
	}
	require.NoError(t, state.Validate())

	root, err := BuildContractZoneRootFromState(77, state, hash("proofs"))
	require.NoError(t, err)
	require.Equal(t, ZoneIDContract, root.ZoneID)
	require.Equal(t, ComputeContractZoneStateRoot(state), root.ZoneStateRoot)

	req, err := ContractProofRequest(ContractProofState, "contract-1/slot-1", 77, root.RootHash, 4)
	require.NoError(t, err)
	require.Equal(t, "QueryContractState/contract-1/slot-1", req.Key)
}

func contractStorage(contractAddr, key, value string) ContractStorageEntry {
	return ContractStorageEntry{
		ContractAddr: contractAddr,
		Key:          key,
		ValueHash:    hash(value),
	}
}

func contractInbox(contractAddr, msgID string, sequence uint64) ContractInboxMessage {
	return ContractInboxMessage{
		ContractAddr:   contractAddr,
		MsgID:          msgID,
		MessageKind:    ContractMessageExecute,
		Source:         "application-zone",
		PayloadHash:    hash(msgID),
		GasLimit:       1_000,
		Sequence:       sequence,
		ReceivedHeight: 77,
	}
}

func contractReceipt(contractAddr, receiptID, msgID string, sequence uint64) ContractExecutionReceipt {
	return ContractExecutionReceipt{
		ContractAddr:    contractAddr,
		ReceiptID:       receiptID,
		MsgID:           msgID,
		MessageKind:     ContractMessageExecute,
		Status:          ZoneReceiptStatusSuccess,
		GasUsed:         500,
		OutputHash:      hash(receiptID + "-output"),
		StorageRoot:     hash(receiptID + "-storage"),
		EmittedMessages: 1,
		Height:          77,
		Sequence:        sequence,
	}
}
