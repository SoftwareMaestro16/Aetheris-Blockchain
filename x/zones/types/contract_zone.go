package types

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

const (
	ContractZonePrefix     = "contract"
	ContractCodePrefix     = ContractZonePrefix + "/code"
	ContractInstancePrefix = ContractZonePrefix + "/instance"
	ContractStoragePrefix  = ContractZonePrefix + "/storage"
	ContractABIPrefix      = ContractZonePrefix + "/abi"
	ContractInboxPrefix    = ContractZonePrefix + "/inbox"
	ContractReceiptPrefix  = ContractZonePrefix + "/receipts"
)

type ContractRuntimeKind string
type ContractMessageKind string
type ContractProofKind string

const (
	ContractRuntimeAVM      ContractRuntimeKind = "AVM"
	ContractRuntimeCosmWasm ContractRuntimeKind = "COSMWASM"

	ContractMessageStoreCode   ContractMessageKind = "MsgStoreCode"
	ContractMessageInstantiate ContractMessageKind = "MsgInstantiateContract"
	ContractMessageExecute     ContractMessageKind = "MsgExecuteContract"
	ContractMessageMigrate     ContractMessageKind = "MsgMigrateContract"
	ContractMessageCallback    ContractMessageKind = "MsgContractCallback"
	ContractMessageProofVerify ContractMessageKind = "MsgContractProofVerify"

	ContractProofCode     ContractProofKind = "QueryCode"
	ContractProofContract ContractProofKind = "QueryContract"
	ContractProofState    ContractProofKind = "QueryContractState"
	ContractProofABI      ContractProofKind = "QueryContractABI"
	ContractProofReceipt  ContractProofKind = "QueryContractReceipt"
)

type ContractBytecodeRuntime interface {
	RuntimeID() string
	ValidateBytecode(context.Context, ContractCode) error
	ExecuteContractMessage(context.Context, ContractInstance, ContractInboxMessage) (ContractExecutionReceipt, []ContractStorageEntry, []ContractInboxMessage, error)
	ComputeBytecodeRoot(context.Context, []ContractCode) (string, error)
	ComputeContractStateRoot(context.Context, ContractInstance, []ContractStorageEntry) (string, error)
}

type AVMBytecodeRuntime = ContractBytecodeRuntime

type CosmWasmAdapterBoundary interface {
	AdapterID() string
	ValidateCosmWasmCode(context.Context, ContractCode) error
	TranslateCosmWasmMessage(context.Context, ContractInboxMessage) (ContractInboxMessage, error)
	AdapterStateRoot(context.Context) (string, error)
}

type ContractZoneBoundary struct {
	ZoneID       ZoneID
	OwnsPrefixes []string
	Messages     []ContractMessageKind
	ProofKinds   []ContractProofKind
}

type ContractBytecodeInterface struct {
	Runtime         ContractRuntimeKind
	InstructionSet  string
	BytecodeHash    string
	ABIHash         string
	DeterminismHash string
	MaxCodeBytes    uint64
	InterfaceHash   string
}

type ContractCosmWasmAdapterDescriptor struct {
	AdapterID      string
	Version        string
	PolicyHash     string
	CapabilityRoot string
	DescriptorHash string
}

type ContractCode struct {
	CodeID         string
	Runtime        ContractRuntimeKind
	BytecodeHash   string
	BytecodeSize   uint64
	ABIHash        string
	InterfaceHash  string
	Uploader       string
	UploadedHeight uint64
}

type ContractInstance struct {
	ContractAddr  string
	CodeID        string
	Runtime       ContractRuntimeKind
	Admin         string
	StorageRoot   string
	CreatedHeight uint64
	UpdatedHeight uint64
}

type ContractStorageEntry struct {
	ContractAddr string
	Key          string
	ValueHash    string
}

type ContractABIDescriptor struct {
	CodeID           string
	ABIHash          string
	InterfaceHash    string
	ExportedMethods  []string
	RegisteredHeight uint64
}

type ContractInboxMessage struct {
	ContractAddr   string
	MsgID          string
	MessageKind    ContractMessageKind
	Source         string
	PayloadHash    string
	GasLimit       uint64
	Sequence       uint64
	ReceivedHeight uint64
}

type ContractExecutionReceipt struct {
	ContractAddr    string
	ReceiptID       string
	MsgID           string
	MessageKind     ContractMessageKind
	Status          ZoneReceiptStatus
	GasUsed         uint64
	OutputHash      string
	StorageRoot     string
	EmittedMessages uint32
	Height          uint64
	Sequence        uint64
	ReceiptHash     string
}

type ContractZoneState struct {
	Height             uint64
	Codes              []ContractCode
	Instances          []ContractInstance
	Storage            []ContractStorageEntry
	ABIs               []ContractABIDescriptor
	Inbox              []ContractInboxMessage
	Receipts           []ContractExecutionReceipt
	BytecodeInterfaces []ContractBytecodeInterface
	CosmWasmAdapters   []ContractCosmWasmAdapterDescriptor
}

type ContractZoneRoots struct {
	Height                uint64
	CodeRoot              string
	InstanceRoot          string
	StorageRoot           string
	ABIRoot               string
	InboxRoot             string
	ReceiptRoot           string
	BytecodeInterfaceRoot string
	CosmWasmAdapterRoot   string
	ExecutionRoot         string
	ProofRoot             string
	StateRoot             string
}

func DefaultContractZoneBoundary() ContractZoneBoundary {
	return ContractZoneBoundary{
		ZoneID: ZoneIDContract,
		OwnsPrefixes: []string{
			ContractABIPrefix,
			ContractCodePrefix,
			ContractInboxPrefix,
			ContractInstancePrefix,
			ContractReceiptPrefix,
			ContractStoragePrefix,
		},
		Messages: []ContractMessageKind{
			ContractMessageStoreCode,
			ContractMessageInstantiate,
			ContractMessageExecute,
			ContractMessageMigrate,
			ContractMessageCallback,
			ContractMessageProofVerify,
		},
		ProofKinds: []ContractProofKind{
			ContractProofCode,
			ContractProofContract,
			ContractProofState,
			ContractProofABI,
			ContractProofReceipt,
		},
	}
}

func (b ContractZoneBoundary) Validate() error {
	if b.ZoneID != ZoneIDContract {
		return errors.New("contract zone boundary must use CONTRACT_ZONE")
	}
	if len(b.OwnsPrefixes) == 0 || len(b.Messages) == 0 || len(b.ProofKinds) == 0 {
		return errors.New("contract zone boundary requires prefixes, messages, and proof kinds")
	}
	for i, prefix := range b.OwnsPrefixes {
		if err := validateRuntimeToken("contract zone prefix", prefix, MaxZoneNamespaceLength); err != nil {
			return err
		}
		if i > 0 && b.OwnsPrefixes[i-1] >= prefix {
			return errors.New("contract zone prefixes must be sorted canonically")
		}
	}
	for _, msg := range b.Messages {
		if !IsContractMessageKind(msg) {
			return fmt.Errorf("unknown contract message kind %q", msg)
		}
	}
	for _, proof := range b.ProofKinds {
		if !IsContractProofKind(proof) {
			return fmt.Errorf("unknown contract proof kind %q", proof)
		}
	}
	return nil
}

func ContractCodeKey(codeID string) (string, error) {
	if err := validateRuntimeToken("contract code id", codeID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractCodePrefix + "/" + codeID, nil
}

func ContractInstanceKey(contractAddr string) (string, error) {
	if err := validateRuntimeToken("contract address", contractAddr, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractInstancePrefix + "/" + contractAddr, nil
}

func ContractStorageKey(contractAddr, key string) (string, error) {
	if _, err := ContractInstanceKey(contractAddr); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("contract storage key", key, MaxZoneProofKeyLength); err != nil {
		return "", err
	}
	return ContractStoragePrefix + "/" + contractAddr + "/" + key, nil
}

func ContractABIKey(codeID string) (string, error) {
	if _, err := ContractCodeKey(codeID); err != nil {
		return "", err
	}
	return ContractABIPrefix + "/" + codeID, nil
}

func ContractInboxKey(contractAddr, msgID string) (string, error) {
	if _, err := ContractInstanceKey(contractAddr); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("contract inbox message id", msgID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractInboxPrefix + "/" + contractAddr + "/" + msgID, nil
}

func ContractReceiptKey(contractAddr, receiptID string) (string, error) {
	if _, err := ContractInstanceKey(contractAddr); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("contract receipt id", receiptID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractReceiptPrefix + "/" + contractAddr + "/" + receiptID, nil
}

func NewContractBytecodeInterface(iface ContractBytecodeInterface) (ContractBytecodeInterface, error) {
	if iface.InterfaceHash != "" {
		return ContractBytecodeInterface{}, errors.New("contract bytecode interface hash must be empty before construction")
	}
	if err := iface.ValidateFormat(); err != nil {
		return ContractBytecodeInterface{}, err
	}
	iface.InterfaceHash = ComputeContractBytecodeInterfaceHash(iface)
	return iface, iface.Validate()
}

func NewContractCosmWasmAdapterDescriptor(adapter ContractCosmWasmAdapterDescriptor) (ContractCosmWasmAdapterDescriptor, error) {
	if adapter.DescriptorHash != "" {
		return ContractCosmWasmAdapterDescriptor{}, errors.New("contract CosmWasm adapter descriptor hash must be empty before construction")
	}
	if err := adapter.ValidateFormat(); err != nil {
		return ContractCosmWasmAdapterDescriptor{}, err
	}
	adapter.DescriptorHash = ComputeContractCosmWasmAdapterHash(adapter)
	return adapter, adapter.Validate()
}

func NewContractExecutionReceipt(receipt ContractExecutionReceipt) (ContractExecutionReceipt, error) {
	if receipt.ReceiptHash != "" {
		return ContractExecutionReceipt{}, errors.New("contract receipt hash must be empty before construction")
	}
	if err := receipt.ValidateFormat(); err != nil {
		return ContractExecutionReceipt{}, err
	}
	receipt.ReceiptHash = ComputeContractExecutionReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func UpsertContractStorage(entries []ContractStorageEntry, update ContractStorageEntry) ([]ContractStorageEntry, error) {
	if err := update.Validate(); err != nil {
		return nil, err
	}
	next := append([]ContractStorageEntry(nil), entries...)
	for i, entry := range next {
		if entry.ContractAddr == update.ContractAddr && entry.Key == update.Key {
			next[i] = update
			return normalizeContractStorage(next), nil
		}
	}
	next = append(next, update)
	return normalizeContractStorage(next), nil
}

func EnqueueContractInbox(inbox []ContractInboxMessage, msg ContractInboxMessage) ([]ContractInboxMessage, error) {
	if err := msg.Validate(); err != nil {
		return nil, err
	}
	next := append([]ContractInboxMessage(nil), inbox...)
	for _, existing := range next {
		if existing.ContractAddr != msg.ContractAddr {
			continue
		}
		if existing.MsgID == msg.MsgID {
			return nil, errors.New("contract inbox message id already exists")
		}
		if existing.Sequence == msg.Sequence {
			return nil, errors.New("contract inbox sequence already exists")
		}
	}
	next = append(next, msg)
	return normalizeContractInbox(next), nil
}

func ContractProofRequest(kind ContractProofKind, key string, height uint64, root string, limit uint32) (ZoneProofRequest, error) {
	if !IsContractProofKind(kind) {
		return ZoneProofRequest{}, fmt.Errorf("unknown contract proof kind %q", kind)
	}
	if err := validateRuntimeToken("contract proof key", key, MaxZoneProofKeyLength); err != nil {
		return ZoneProofRequest{}, err
	}
	req := ZoneProofRequest{
		ZoneID: ZoneIDContract,
		Height: height,
		Kind:   ZoneProofKindState,
		Key:    string(kind) + "/" + key,
		Root:   root,
		Limit:  limit,
	}
	return req, req.Validate()
}

func BuildContractZoneRoot(roots ContractZoneRoots) (ZoneRoot, error) {
	if err := roots.Validate(); err != nil {
		return ZoneRoot{}, err
	}
	stateRoot := roots.StateRoot
	if stateRoot == "" {
		stateRoot = hashRuntimeParts(
			"aetheris-contract-zone-state-v1",
			roots.CodeRoot,
			roots.InstanceRoot,
			roots.StorageRoot,
			roots.ABIRoot,
			roots.InboxRoot,
			roots.ReceiptRoot,
			roots.BytecodeInterfaceRoot,
			roots.CosmWasmAdapterRoot,
		)
	}
	root := ZoneRoot{
		ZoneID:              ZoneIDContract,
		Height:              roots.Height,
		ZoneStateRoot:       stateRoot,
		InboxRoot:           roots.InboxRoot,
		OutboxRoot:          EmptyRootHash(),
		ReceiptRoot:         roots.ReceiptRoot,
		ExecutionResultRoot: roots.ExecutionRoot,
		ProofRoot:           roots.ProofRoot,
	}
	root.RootHash = ComputeZoneRootHash(root)
	return root, root.Validate()
}

func BuildContractZoneRootFromState(height uint64, state ContractZoneState, proofRoot string) (ZoneRoot, error) {
	normalized := state.Normalize()
	roots := ContractZoneRoots{
		Height:                height,
		CodeRoot:              ComputeContractCodeRoot(normalized.Codes),
		InstanceRoot:          ComputeContractInstanceRoot(normalized.Instances),
		StorageRoot:           ComputeContractStorageRoot(normalized.Storage),
		ABIRoot:               ComputeContractABIRoot(normalized.ABIs),
		InboxRoot:             ComputeContractInboxRoot(normalized.Inbox),
		ReceiptRoot:           ComputeContractReceiptRoot(normalized.Receipts),
		BytecodeInterfaceRoot: ComputeContractBytecodeInterfaceRoot(normalized.BytecodeInterfaces),
		CosmWasmAdapterRoot:   ComputeContractCosmWasmAdapterRoot(normalized.CosmWasmAdapters),
		ExecutionRoot:         ComputeContractExecutionRoot(normalized.Receipts),
		ProofRoot:             proofRoot,
		StateRoot:             ComputeContractZoneStateRoot(normalized),
	}
	return BuildContractZoneRoot(roots)
}

func (s ContractZoneState) Normalize() ContractZoneState {
	s.Codes = normalizeContractCodes(s.Codes)
	s.Instances = normalizeContractInstances(s.Instances)
	s.Storage = normalizeContractStorage(s.Storage)
	s.ABIs = normalizeContractABIs(s.ABIs)
	s.Inbox = normalizeContractInbox(s.Inbox)
	s.Receipts = normalizeContractReceipts(s.Receipts)
	s.BytecodeInterfaces = normalizeContractBytecodeInterfaces(s.BytecodeInterfaces)
	s.CosmWasmAdapters = normalizeContractCosmWasmAdapters(s.CosmWasmAdapters)
	return s
}

func (s ContractZoneState) Validate() error {
	normalized := s.Normalize()
	for _, code := range normalized.Codes {
		if err := code.Validate(); err != nil {
			return err
		}
	}
	for _, instance := range normalized.Instances {
		if err := instance.Validate(); err != nil {
			return err
		}
	}
	for _, entry := range normalized.Storage {
		if err := entry.Validate(); err != nil {
			return err
		}
	}
	for _, abi := range normalized.ABIs {
		if err := abi.Validate(); err != nil {
			return err
		}
	}
	for _, msg := range normalized.Inbox {
		if err := msg.Validate(); err != nil {
			return err
		}
	}
	for _, receipt := range normalized.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
	}
	for _, iface := range normalized.BytecodeInterfaces {
		if err := iface.Validate(); err != nil {
			return err
		}
	}
	for _, adapter := range normalized.CosmWasmAdapters {
		if err := adapter.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (i ContractBytecodeInterface) ValidateFormat() error {
	if !IsContractRuntimeKind(i.Runtime) {
		return fmt.Errorf("unknown contract runtime kind %q", i.Runtime)
	}
	if err := validateRuntimeToken("contract instruction set", i.InstructionSet, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("contract bytecode interface bytecode hash", i.BytecodeHash); err != nil {
		return err
	}
	if err := ValidateHash("contract bytecode interface ABI hash", i.ABIHash); err != nil {
		return err
	}
	if err := ValidateHash("contract bytecode interface determinism hash", i.DeterminismHash); err != nil {
		return err
	}
	if i.MaxCodeBytes == 0 {
		return errors.New("contract bytecode interface max code bytes must be positive")
	}
	if i.InterfaceHash != "" {
		return ValidateHash("contract bytecode interface hash", i.InterfaceHash)
	}
	return nil
}

func (i ContractBytecodeInterface) Validate() error {
	if err := i.ValidateFormat(); err != nil {
		return err
	}
	if i.InterfaceHash == "" {
		return errors.New("contract bytecode interface hash is required")
	}
	if i.InterfaceHash != ComputeContractBytecodeInterfaceHash(i) {
		return errors.New("contract bytecode interface hash mismatch")
	}
	return nil
}

func (a ContractCosmWasmAdapterDescriptor) ValidateFormat() error {
	if err := validateRuntimeToken("contract CosmWasm adapter id", a.AdapterID, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("contract CosmWasm adapter version", a.Version, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("contract CosmWasm adapter policy hash", a.PolicyHash); err != nil {
		return err
	}
	if err := ValidateHash("contract CosmWasm adapter capability root", a.CapabilityRoot); err != nil {
		return err
	}
	if a.DescriptorHash != "" {
		return ValidateHash("contract CosmWasm adapter descriptor hash", a.DescriptorHash)
	}
	return nil
}

func (a ContractCosmWasmAdapterDescriptor) Validate() error {
	if err := a.ValidateFormat(); err != nil {
		return err
	}
	if a.DescriptorHash == "" {
		return errors.New("contract CosmWasm adapter descriptor hash is required")
	}
	if a.DescriptorHash != ComputeContractCosmWasmAdapterHash(a) {
		return errors.New("contract CosmWasm adapter descriptor hash mismatch")
	}
	return nil
}

func (c ContractCode) Validate() error {
	if _, err := ContractCodeKey(c.CodeID); err != nil {
		return err
	}
	if !IsContractRuntimeKind(c.Runtime) {
		return fmt.Errorf("unknown contract runtime kind %q", c.Runtime)
	}
	if err := ValidateHash("contract bytecode hash", c.BytecodeHash); err != nil {
		return err
	}
	if c.BytecodeSize == 0 {
		return errors.New("contract bytecode size must be positive")
	}
	if err := ValidateHash("contract ABI hash", c.ABIHash); err != nil {
		return err
	}
	if err := ValidateHash("contract interface hash", c.InterfaceHash); err != nil {
		return err
	}
	if err := validateRuntimeToken("contract code uploader", c.Uploader, MaxZoneEndpointLength); err != nil {
		return err
	}
	if c.UploadedHeight == 0 {
		return errors.New("contract code uploaded height must be positive")
	}
	return nil
}

func (i ContractInstance) Validate() error {
	if _, err := ContractInstanceKey(i.ContractAddr); err != nil {
		return err
	}
	if _, err := ContractCodeKey(i.CodeID); err != nil {
		return err
	}
	if !IsContractRuntimeKind(i.Runtime) {
		return fmt.Errorf("unknown contract runtime kind %q", i.Runtime)
	}
	if err := validateRuntimeToken("contract admin", i.Admin, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("contract instance storage root", i.StorageRoot); err != nil {
		return err
	}
	if i.CreatedHeight == 0 || i.UpdatedHeight == 0 {
		return errors.New("contract instance heights must be positive")
	}
	if i.UpdatedHeight < i.CreatedHeight {
		return errors.New("contract instance updated height must not precede created height")
	}
	return nil
}

func (e ContractStorageEntry) Validate() error {
	if _, err := ContractStorageKey(e.ContractAddr, e.Key); err != nil {
		return err
	}
	return ValidateHash("contract storage value hash", e.ValueHash)
}

func (a ContractABIDescriptor) Validate() error {
	if _, err := ContractABIKey(a.CodeID); err != nil {
		return err
	}
	if err := ValidateHash("contract ABI hash", a.ABIHash); err != nil {
		return err
	}
	if err := ValidateHash("contract interface hash", a.InterfaceHash); err != nil {
		return err
	}
	if a.RegisteredHeight == 0 {
		return errors.New("contract ABI registered height must be positive")
	}
	methods := append([]string(nil), a.ExportedMethods...)
	sort.Strings(methods)
	for i, method := range methods {
		if err := validateRuntimeToken("contract ABI method", method, MaxZoneEndpointLength); err != nil {
			return err
		}
		if i > 0 && methods[i-1] == method {
			return errors.New("contract ABI methods must be unique")
		}
	}
	return nil
}

func (m ContractInboxMessage) Validate() error {
	if _, err := ContractInboxKey(m.ContractAddr, m.MsgID); err != nil {
		return err
	}
	if !IsContractMessageKind(m.MessageKind) {
		return fmt.Errorf("unknown contract inbox message kind %q", m.MessageKind)
	}
	if err := validateRuntimeToken("contract inbox source", m.Source, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("contract inbox payload hash", m.PayloadHash); err != nil {
		return err
	}
	if m.GasLimit == 0 {
		return errors.New("contract inbox gas limit must be positive")
	}
	if m.Sequence == 0 || m.ReceivedHeight == 0 {
		return errors.New("contract inbox sequence and height must be positive")
	}
	return nil
}

func (r ContractExecutionReceipt) ValidateFormat() error {
	if _, err := ContractReceiptKey(r.ContractAddr, r.ReceiptID); err != nil {
		return err
	}
	if err := validateRuntimeToken("contract receipt message id", r.MsgID, MaxZoneEndpointLength); err != nil {
		return err
	}
	if !IsContractMessageKind(r.MessageKind) {
		return fmt.Errorf("unknown contract receipt message kind %q", r.MessageKind)
	}
	if !IsZoneReceiptStatus(r.Status) {
		return fmt.Errorf("unknown contract receipt status %q", r.Status)
	}
	if err := ValidateHash("contract receipt output hash", r.OutputHash); err != nil {
		return err
	}
	if err := ValidateHash("contract receipt storage root", r.StorageRoot); err != nil {
		return err
	}
	if r.Height == 0 || r.Sequence == 0 {
		return errors.New("contract receipt height and sequence must be positive")
	}
	if r.ReceiptHash != "" {
		return ValidateHash("contract receipt hash", r.ReceiptHash)
	}
	return nil
}

func (r ContractExecutionReceipt) Validate() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.ReceiptHash == "" {
		return errors.New("contract receipt hash is required")
	}
	if r.ReceiptHash != ComputeContractExecutionReceiptHash(r) {
		return errors.New("contract receipt hash mismatch")
	}
	return nil
}

func (r ContractExecutionReceipt) ZoneReceipt() (ZoneReceipt, error) {
	return NewZoneReceipt(ZoneReceipt{
		ZoneID:     ZoneIDContract,
		Height:     r.Height,
		ItemHash:   hashRuntimeParts("contract-zone-receipt-item-v1", r.ContractAddr, r.ReceiptID, r.MsgID),
		Status:     r.Status,
		GasUsed:    r.GasUsed,
		ResultHash: r.OutputHash,
		Sequence:   r.Sequence,
	})
}

func (r ContractZoneRoots) Validate() error {
	if r.Height == 0 {
		return errors.New("contract zone root height must be positive")
	}
	for _, item := range []struct {
		name  string
		value string
	}{
		{name: "contract code root", value: r.CodeRoot},
		{name: "contract instance root", value: r.InstanceRoot},
		{name: "contract storage root", value: r.StorageRoot},
		{name: "contract ABI root", value: r.ABIRoot},
		{name: "contract inbox root", value: r.InboxRoot},
		{name: "contract receipt root", value: r.ReceiptRoot},
		{name: "contract bytecode interface root", value: r.BytecodeInterfaceRoot},
		{name: "contract CosmWasm adapter root", value: r.CosmWasmAdapterRoot},
		{name: "contract execution root", value: r.ExecutionRoot},
		{name: "contract proof root", value: r.ProofRoot},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.StateRoot != "" {
		return ValidateHash("contract state root", r.StateRoot)
	}
	return nil
}

func ComputeContractBytecodeInterfaceHash(iface ContractBytecodeInterface) string {
	return hashRuntimeParts(
		"aetheris-contract-bytecode-interface-v1",
		string(iface.Runtime),
		iface.InstructionSet,
		iface.BytecodeHash,
		iface.ABIHash,
		iface.DeterminismHash,
		fmt.Sprint(iface.MaxCodeBytes),
	)
}

func ComputeContractCosmWasmAdapterHash(adapter ContractCosmWasmAdapterDescriptor) string {
	return hashRuntimeParts(
		"aetheris-contract-cosmwasm-adapter-v1",
		adapter.AdapterID,
		adapter.Version,
		adapter.PolicyHash,
		adapter.CapabilityRoot,
	)
}

func ComputeContractExecutionReceiptHash(receipt ContractExecutionReceipt) string {
	return hashRuntimeParts(
		"aetheris-contract-zone-receipt-v1",
		receipt.ContractAddr,
		receipt.ReceiptID,
		receipt.MsgID,
		string(receipt.MessageKind),
		string(receipt.Status),
		fmt.Sprint(receipt.GasUsed),
		receipt.OutputHash,
		receipt.StorageRoot,
		fmt.Sprint(receipt.EmittedMessages),
		fmt.Sprint(receipt.Height),
		fmt.Sprint(receipt.Sequence),
	)
}

func ComputeContractZoneStateRoot(state ContractZoneState) string {
	normalized := state.Normalize()
	return hashRuntimeParts(
		"aetheris-contract-zone-state-v1",
		ComputeContractCodeRoot(normalized.Codes),
		ComputeContractInstanceRoot(normalized.Instances),
		ComputeContractStorageRoot(normalized.Storage),
		ComputeContractABIRoot(normalized.ABIs),
		ComputeContractInboxRoot(normalized.Inbox),
		ComputeContractReceiptRoot(normalized.Receipts),
		ComputeContractBytecodeInterfaceRoot(normalized.BytecodeInterfaces),
		ComputeContractCosmWasmAdapterRoot(normalized.CosmWasmAdapters),
	)
}

func ComputeContractCodeRoot(codes []ContractCode) string {
	ordered := normalizeContractCodes(codes)
	parts := []string{"aetheris-contract-code-root-v1", fmt.Sprint(len(ordered))}
	for _, code := range ordered {
		parts = append(parts, code.CodeID, string(code.Runtime), code.BytecodeHash, fmt.Sprint(code.BytecodeSize), code.ABIHash, code.InterfaceHash, code.Uploader, fmt.Sprint(code.UploadedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractInstanceRoot(instances []ContractInstance) string {
	ordered := normalizeContractInstances(instances)
	parts := []string{"aetheris-contract-instance-root-v1", fmt.Sprint(len(ordered))}
	for _, instance := range ordered {
		parts = append(parts, instance.ContractAddr, instance.CodeID, string(instance.Runtime), instance.Admin, instance.StorageRoot, fmt.Sprint(instance.CreatedHeight), fmt.Sprint(instance.UpdatedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractStorageRoot(entries []ContractStorageEntry) string {
	ordered := normalizeContractStorage(entries)
	parts := []string{"aetheris-contract-storage-root-v1", fmt.Sprint(len(ordered))}
	for _, entry := range ordered {
		key, _ := ContractStorageKey(entry.ContractAddr, entry.Key)
		parts = append(parts, key, entry.ValueHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractABIRoot(abis []ContractABIDescriptor) string {
	ordered := normalizeContractABIs(abis)
	parts := []string{"aetheris-contract-abi-root-v1", fmt.Sprint(len(ordered))}
	for _, abi := range ordered {
		methods := append([]string(nil), abi.ExportedMethods...)
		sort.Strings(methods)
		parts = append(parts, abi.CodeID, abi.ABIHash, abi.InterfaceHash, fmt.Sprint(abi.RegisteredHeight))
		parts = append(parts, methods...)
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractInboxRoot(inbox []ContractInboxMessage) string {
	ordered := normalizeContractInbox(inbox)
	parts := []string{"aetheris-contract-inbox-root-v1", fmt.Sprint(len(ordered))}
	for _, msg := range ordered {
		key, _ := ContractInboxKey(msg.ContractAddr, msg.MsgID)
		parts = append(parts, key, string(msg.MessageKind), msg.Source, msg.PayloadHash, fmt.Sprint(msg.GasLimit), fmt.Sprint(msg.Sequence), fmt.Sprint(msg.ReceivedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractReceiptRoot(receipts []ContractExecutionReceipt) string {
	ordered := normalizeContractReceipts(receipts)
	parts := []string{"aetheris-contract-receipt-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ReceiptHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractExecutionRoot(receipts []ContractExecutionReceipt) string {
	ordered := normalizeContractReceipts(receipts)
	parts := []string{"aetheris-contract-execution-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ContractAddr, receipt.MsgID, string(receipt.Status), receipt.OutputHash, receipt.StorageRoot, fmt.Sprint(receipt.GasUsed))
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractBytecodeInterfaceRoot(interfaces []ContractBytecodeInterface) string {
	ordered := normalizeContractBytecodeInterfaces(interfaces)
	parts := []string{"aetheris-contract-bytecode-interface-root-v1", fmt.Sprint(len(ordered))}
	for _, iface := range ordered {
		parts = append(parts, string(iface.Runtime), iface.InstructionSet, iface.BytecodeHash, iface.ABIHash, iface.DeterminismHash, fmt.Sprint(iface.MaxCodeBytes), iface.InterfaceHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractCosmWasmAdapterRoot(adapters []ContractCosmWasmAdapterDescriptor) string {
	ordered := normalizeContractCosmWasmAdapters(adapters)
	parts := []string{"aetheris-contract-cosmwasm-adapter-root-v1", fmt.Sprint(len(ordered))}
	for _, adapter := range ordered {
		parts = append(parts, adapter.AdapterID, adapter.Version, adapter.PolicyHash, adapter.CapabilityRoot, adapter.DescriptorHash)
	}
	return hashRuntimeParts(parts...)
}

func IsContractRuntimeKind(kind ContractRuntimeKind) bool {
	switch kind {
	case ContractRuntimeAVM, ContractRuntimeCosmWasm:
		return true
	default:
		return false
	}
}

func IsContractMessageKind(kind ContractMessageKind) bool {
	switch kind {
	case ContractMessageStoreCode,
		ContractMessageInstantiate,
		ContractMessageExecute,
		ContractMessageMigrate,
		ContractMessageCallback,
		ContractMessageProofVerify:
		return true
	default:
		return false
	}
}

func IsContractProofKind(kind ContractProofKind) bool {
	switch kind {
	case ContractProofCode,
		ContractProofContract,
		ContractProofState,
		ContractProofABI,
		ContractProofReceipt:
		return true
	default:
		return false
	}
}

func normalizeContractCodes(codes []ContractCode) []ContractCode {
	out := append([]ContractCode(nil), codes...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CodeID < out[j].CodeID })
	return out
}

func normalizeContractInstances(instances []ContractInstance) []ContractInstance {
	out := append([]ContractInstance(nil), instances...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ContractAddr < out[j].ContractAddr })
	return out
}

func normalizeContractStorage(entries []ContractStorageEntry) []ContractStorageEntry {
	out := append([]ContractStorageEntry(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ContractAddr != out[j].ContractAddr {
			return out[i].ContractAddr < out[j].ContractAddr
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func normalizeContractABIs(abis []ContractABIDescriptor) []ContractABIDescriptor {
	out := append([]ContractABIDescriptor(nil), abis...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CodeID < out[j].CodeID })
	return out
}

func normalizeContractInbox(inbox []ContractInboxMessage) []ContractInboxMessage {
	out := append([]ContractInboxMessage(nil), inbox...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ContractAddr != out[j].ContractAddr {
			return out[i].ContractAddr < out[j].ContractAddr
		}
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		return out[i].MsgID < out[j].MsgID
	})
	return out
}

func normalizeContractReceipts(receipts []ContractExecutionReceipt) []ContractExecutionReceipt {
	out := append([]ContractExecutionReceipt(nil), receipts...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ContractAddr != out[j].ContractAddr {
			return out[i].ContractAddr < out[j].ContractAddr
		}
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		return out[i].ReceiptID < out[j].ReceiptID
	})
	return out
}

func normalizeContractBytecodeInterfaces(interfaces []ContractBytecodeInterface) []ContractBytecodeInterface {
	out := append([]ContractBytecodeInterface(nil), interfaces...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Runtime != out[j].Runtime {
			return out[i].Runtime < out[j].Runtime
		}
		return out[i].InstructionSet < out[j].InstructionSet
	})
	return out
}

func normalizeContractCosmWasmAdapters(adapters []ContractCosmWasmAdapterDescriptor) []ContractCosmWasmAdapterDescriptor {
	out := append([]ContractCosmWasmAdapterDescriptor(nil), adapters...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].AdapterID != out[j].AdapterID {
			return out[i].AdapterID < out[j].AdapterID
		}
		return out[i].Version < out[j].Version
	})
	return out
}
