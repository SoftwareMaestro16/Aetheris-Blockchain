package async

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	AddressDerivationDomain = "aetheris/async-contract/v1"
	BounceOpcode            = uint32(0xffff_fffe)
	RefundOpcode            = uint32(0xffff_fffd)
	ResultOK                = uint32(0)
	ResultNoDestination     = uint32(1)
	ResultExpired           = uint32(2)
	ResultExecutionFailed   = uint32(3)
	ResultLimitExceeded     = uint32(4)
	CodeHashLength          = 32
)

type Params struct {
	MaxMessagesPerTx           uint32
	MaxMessagesPerBlock        uint32
	MaxRecursionDepth          uint32
	MaxBodySize                uint32
	MaxStateSize               uint32
	MaxContractDeploysPerTx    uint32
	MaxContractDeploysPerBlock uint32
	MaxEmittedMessagesPerExec  uint32
	MaxStorageWritesPerExec    uint32
	ExecutionGasPerMessage     uint64
	StorageFeePerByte          sdkmath.Int
	ForwardingFee              sdkmath.Int
	ContractDeploymentCost     sdkmath.Int
}

type ContractAccount struct {
	Address     sdk.AccAddress
	CodeHash    []byte
	State       []byte
	BalanceNaet sdkmath.Int
	LogicalTime uint64
}

type MessageEnvelope struct {
	Source             sdk.AccAddress
	Destination        sdk.AccAddress
	Value              sdk.Coin
	Opcode             uint32
	QueryID            uint64
	Body               []byte
	Bounce             bool
	Bounced            bool
	CreatedLogicalTime uint64
	DeadlineBlock      uint64
	GasLimit           uint64
	ForwardFee         sdk.Coin
	Depth              uint32
}

type QueuedMessage struct {
	TxIndex           uint64
	MessageIndex      uint32
	SourceLogicalTime uint64
	DestinationKey    string
	Sequence          uint64
	EnqueuedBlock     uint64
	Envelope          MessageEnvelope
}

type ExecutionResult struct {
	NewState      []byte
	Outgoing      []MessageEnvelope
	GasUsed       uint64
	StorageWrites uint32
	ResultCode    uint32
	Error         string
}

type ExecutionReceipt struct {
	Sequence       uint64
	Source         sdk.AccAddress
	Destination    sdk.AccAddress
	Opcode         uint32
	QueryID        uint64
	ResultCode     uint32
	GasUsed        uint64
	StorageFeeNaet sdkmath.Int
	ForwardFeeNaet sdkmath.Int
	Bounced        bool
	Error          string
}

type Observability struct {
	QueuedMessages    uint64
	ProcessedMessages uint64
	BouncedMessages   uint64
	RefundMessages    uint64
	FailedExecutions  uint64
	GasUsed           uint64
	QueueLag          uint64
}

type ExportedState struct {
	Params       Params
	Contracts    []ContractAccount
	Queue        []QueuedMessage
	Inbox        map[string][]QueuedMessage
	Outbox       map[string][]QueuedMessage
	Receipts     []ExecutionReceipt
	NextSequence uint64
	NextTxIndex  uint64
	BlockHeight  uint64
	Metrics      Observability
}

type Handler func(contract ContractAccount, msg MessageEnvelope) ExecutionResult

type Executor struct {
	params         Params
	contracts      map[string]ContractAccount
	queue          []QueuedMessage
	inbox          map[string][]QueuedMessage
	outbox         map[string][]QueuedMessage
	receipts       []ExecutionReceipt
	nextSequence   uint64
	nextTxIndex    uint64
	blockHeight    uint64
	metrics        Observability
	handlers       map[string]Handler
	deploysInBlock uint32
}

func DefaultParams() Params {
	return Params{
		MaxMessagesPerTx:           32,
		MaxMessagesPerBlock:        128,
		MaxRecursionDepth:          8,
		MaxBodySize:                4096,
		MaxStateSize:               64 * 1024,
		MaxContractDeploysPerTx:    4,
		MaxContractDeploysPerBlock: 16,
		MaxEmittedMessagesPerExec:  16,
		MaxStorageWritesPerExec:    64,
		ExecutionGasPerMessage:     10_000,
		StorageFeePerByte:          sdkmath.NewInt(1),
		ForwardingFee:              sdkmath.NewInt(1),
		ContractDeploymentCost:     sdkmath.NewInt(1_000),
	}
}

func NewExecutor(params Params) (*Executor, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &Executor{
		params:    params,
		contracts: make(map[string]ContractAccount),
		inbox:     make(map[string][]QueuedMessage),
		outbox:    make(map[string][]QueuedMessage),
		handlers:  make(map[string]Handler),
	}, nil
}

func DeriveContractAddress(deployer sdk.AccAddress, codeHash []byte, salt []byte) (sdk.AccAddress, error) {
	if err := aetherisaddress.RejectZeroAddress("contract deployer", deployer); err != nil {
		return nil, err
	}
	if len(codeHash) != CodeHashLength {
		return nil, fmt.Errorf("contract code hash must be %d bytes", CodeHashLength)
	}
	h := sha256.New()
	writePart(h.Write, []byte(AddressDerivationDomain))
	writePart(h.Write, deployer)
	writePart(h.Write, codeHash)
	writePart(h.Write, salt)
	return sdk.AccAddress(h.Sum(nil)), nil
}

func (p Params) Validate() error {
	if p.MaxMessagesPerTx == 0 {
		return errors.New("max messages per tx must be positive")
	}
	if p.MaxMessagesPerBlock == 0 {
		return errors.New("max messages per block must be positive")
	}
	if p.MaxRecursionDepth == 0 {
		return errors.New("max recursion depth must be positive")
	}
	if p.MaxBodySize == 0 {
		return errors.New("max body size must be positive")
	}
	if p.MaxStateSize == 0 {
		return errors.New("max state size must be positive")
	}
	if p.MaxContractDeploysPerTx == 0 {
		return errors.New("max contract deploys per tx must be positive")
	}
	if p.MaxContractDeploysPerBlock == 0 {
		return errors.New("max contract deploys per block must be positive")
	}
	if p.MaxEmittedMessagesPerExec == 0 {
		return errors.New("max emitted messages per execution must be positive")
	}
	if p.MaxStorageWritesPerExec == 0 {
		return errors.New("max storage writes per execution must be positive")
	}
	if p.ExecutionGasPerMessage == 0 {
		return errors.New("execution gas per message must be positive")
	}
	for name, value := range map[string]sdkmath.Int{
		"storage fee per byte":     p.StorageFeePerByte,
		"forwarding fee":           p.ForwardingFee,
		"contract deployment cost": p.ContractDeploymentCost,
	} {
		if value.IsNil() || value.IsNegative() {
			return fmt.Errorf("%s must be non-negative", name)
		}
	}
	return nil
}

func (c ContractAccount) Validate(params Params) error {
	if err := aetherisaddress.RejectZeroAddress("contract account", c.Address); err != nil {
		return err
	}
	if len(c.CodeHash) != CodeHashLength {
		return fmt.Errorf("contract code hash must be %d bytes", CodeHashLength)
	}
	if len(c.State) > int(params.MaxStateSize) {
		return fmt.Errorf("contract state size must be <= %d", params.MaxStateSize)
	}
	if c.BalanceNaet.IsNil() || c.BalanceNaet.IsNegative() {
		return errors.New("contract naet balance must be non-negative")
	}
	return nil
}

func (m MessageEnvelope) Validate(params Params) error {
	if err := aetherisaddress.RejectZeroAddress("message source", m.Source); err != nil {
		return err
	}
	if err := aetherisaddress.RejectZeroAddress("message destination", m.Destination); err != nil {
		return err
	}
	if m.Value.Denom != appparams.BaseDenom {
		return fmt.Errorf("message value denom must be %s", appparams.BaseDenom)
	}
	if !m.Value.IsValid() || m.Value.Amount.IsNegative() {
		return errors.New("message value must be valid and non-negative")
	}
	if len(m.Body) > int(params.MaxBodySize) {
		return fmt.Errorf("message body size must be <= %d", params.MaxBodySize)
	}
	if m.GasLimit == 0 {
		return errors.New("message gas limit must be positive")
	}
	if m.ForwardFee.Denom != appparams.BaseDenom {
		return fmt.Errorf("message forward fee denom must be %s", appparams.BaseDenom)
	}
	if !m.ForwardFee.IsValid() || m.ForwardFee.Amount.IsNegative() {
		return errors.New("message forward fee must be valid and non-negative")
	}
	if m.Depth > params.MaxRecursionDepth {
		return fmt.Errorf("message depth must be <= %d", params.MaxRecursionDepth)
	}
	return nil
}

func (e *Executor) RegisterHandler(address sdk.AccAddress, handler Handler) error {
	if err := aetherisaddress.RejectZeroAddress("contract handler", address); err != nil {
		return err
	}
	if _, ok := e.contracts[string(address)]; !ok {
		return errors.New("cannot register handler for missing contract")
	}
	e.handlers[string(address)] = handler
	return nil
}

func (e *Executor) DeployContract(deployer sdk.AccAddress, codeHash []byte, salt []byte, state []byte, balance sdkmath.Int) (sdk.AccAddress, error) {
	return e.DeployContracts(deployer, []DeploySpec{{CodeHash: codeHash, Salt: salt, State: state, BalanceNaet: balance}})
}

type DeploySpec struct {
	CodeHash    []byte
	Salt        []byte
	State       []byte
	BalanceNaet sdkmath.Int
}

func (e *Executor) DeployContracts(deployer sdk.AccAddress, specs []DeploySpec) (sdk.AccAddress, error) {
	if len(specs) == 0 {
		return nil, errors.New("deploy count must be positive")
	}
	if len(specs) > int(e.params.MaxContractDeploysPerTx) {
		return nil, fmt.Errorf("contract deploys per tx must be <= %d", e.params.MaxContractDeploysPerTx)
	}
	if e.deploysInBlock+uint32(len(specs)) > e.params.MaxContractDeploysPerBlock {
		return nil, fmt.Errorf("contract deploys per block must be <= %d", e.params.MaxContractDeploysPerBlock)
	}
	var first sdk.AccAddress
	for _, spec := range specs {
		if spec.BalanceNaet.IsNil() || spec.BalanceNaet.LT(e.params.ContractDeploymentCost) {
			return nil, errors.New("contract deployment requires naet deployment cost")
		}
		address, err := DeriveContractAddress(deployer, spec.CodeHash, spec.Salt)
		if err != nil {
			return nil, err
		}
		if _, exists := e.contracts[string(address)]; exists {
			return nil, errors.New("contract already deployed")
		}
		contract := ContractAccount{
			Address:     address,
			CodeHash:    append([]byte(nil), spec.CodeHash...),
			State:       append([]byte(nil), spec.State...),
			BalanceNaet: spec.BalanceNaet.Sub(e.params.ContractDeploymentCost),
		}
		if err := contract.Validate(e.params); err != nil {
			return nil, err
		}
		e.contracts[string(address)] = contract
		if first == nil {
			first = address
		}
	}
	e.deploysInBlock += uint32(len(specs))
	return first, nil
}

func (e *Executor) EnqueueTxMessages(messages []MessageEnvelope) error {
	if len(messages) == 0 {
		return errors.New("tx message count must be positive")
	}
	if len(messages) > int(e.params.MaxMessagesPerTx) {
		return fmt.Errorf("messages per tx must be <= %d", e.params.MaxMessagesPerTx)
	}
	txIndex := e.nextTxIndex
	e.nextTxIndex++
	for i, msg := range messages {
		if err := e.enqueueMessageWithOrder(msg, txIndex, uint32(i)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Executor) EnqueueMessage(msg MessageEnvelope) error {
	txIndex := e.nextTxIndex
	e.nextTxIndex++
	return e.enqueueMessageWithOrder(msg, txIndex, 0)
}

func (e *Executor) enqueueMessageWithOrder(msg MessageEnvelope, txIndex uint64, messageIndex uint32) error {
	if err := msg.Validate(e.params); err != nil {
		return err
	}
	queued := QueuedMessage{
		TxIndex:           txIndex,
		MessageIndex:      messageIndex,
		SourceLogicalTime: msg.CreatedLogicalTime,
		DestinationKey:    string(msg.Destination),
		Sequence:          e.nextSequence,
		EnqueuedBlock:     e.blockHeight,
		Envelope:          cloneMessage(msg),
	}
	e.nextSequence++
	e.queue = append(e.queue, queued)
	sort.SliceStable(e.queue, func(i, j int) bool {
		return queuedMessageLess(e.queue[i], e.queue[j])
	})
	e.inbox[string(msg.Destination)] = append(e.inbox[string(msg.Destination)], queued)
	e.outbox[string(msg.Source)] = append(e.outbox[string(msg.Source)], queued)
	e.metrics.QueuedMessages++
	return nil
}

func (e *Executor) ProcessBlock(height uint64) ([]ExecutionReceipt, error) {
	e.blockHeight = height
	e.deploysInBlock = 0
	if e.params.MaxMessagesPerBlock == 0 {
		return nil, errors.New("max messages per block must be positive")
	}
	count := uint32(0)
	receipts := make([]ExecutionReceipt, 0)
	for len(e.queue) > 0 && count < e.params.MaxMessagesPerBlock {
		receipt, err := e.processNext()
		if err != nil {
			return receipts, err
		}
		receipts = append(receipts, receipt)
		count++
	}
	e.updateQueueLag()
	return receipts, nil
}

func (e *Executor) ExportState() ExportedState {
	contracts := make([]ContractAccount, 0, len(e.contracts))
	for _, contract := range e.contracts {
		contracts = append(contracts, cloneContract(contract))
	}
	sort.Slice(contracts, func(i, j int) bool {
		return string(contracts[i].Address) < string(contracts[j].Address)
	})
	return ExportedState{
		Params:       e.params,
		Contracts:    contracts,
		Queue:        cloneQueuedMessages(e.queue),
		Inbox:        cloneQueuedMap(e.inbox),
		Outbox:       cloneQueuedMap(e.outbox),
		Receipts:     cloneReceipts(e.receipts),
		NextSequence: e.nextSequence,
		NextTxIndex:  e.nextTxIndex,
		BlockHeight:  e.blockHeight,
		Metrics:      e.metrics,
	}
}

func ImportState(exported ExportedState) (*Executor, error) {
	executor, err := NewExecutor(exported.Params)
	if err != nil {
		return nil, err
	}
	if err := ValidateExportedState(exported); err != nil {
		return nil, err
	}
	executor.queue = cloneQueuedMessages(exported.Queue)
	executor.inbox = cloneQueuedMap(exported.Inbox)
	executor.outbox = cloneQueuedMap(exported.Outbox)
	executor.receipts = cloneReceipts(exported.Receipts)
	executor.nextSequence = exported.NextSequence
	executor.nextTxIndex = exported.NextTxIndex
	executor.blockHeight = exported.BlockHeight
	executor.metrics = exported.Metrics
	for _, contract := range exported.Contracts {
		if err := contract.Validate(exported.Params); err != nil {
			return nil, err
		}
		executor.contracts[string(contract.Address)] = cloneContract(contract)
	}
	return executor, nil
}

func ValidateExportedState(exported ExportedState) error {
	if err := exported.Params.Validate(); err != nil {
		return err
	}
	seenContracts := make(map[string]struct{}, len(exported.Contracts))
	for _, contract := range exported.Contracts {
		if err := contract.Validate(exported.Params); err != nil {
			return err
		}
		key := string(contract.Address)
		if _, exists := seenContracts[key]; exists {
			return fmt.Errorf("duplicate contract address: %s", contract.Address.String())
		}
		seenContracts[key] = struct{}{}
	}
	seenSequences := make(map[uint64]struct{}, len(exported.Queue))
	for i, queued := range exported.Queue {
		if _, exists := seenSequences[queued.Sequence]; exists {
			return fmt.Errorf("duplicate queued message sequence: %d", queued.Sequence)
		}
		seenSequences[queued.Sequence] = struct{}{}
		if queued.Sequence >= exported.NextSequence {
			return fmt.Errorf("queued message sequence %d must be less than next_sequence %d", queued.Sequence, exported.NextSequence)
		}
		if queued.TxIndex >= exported.NextTxIndex {
			return fmt.Errorf("queued message tx_index %d must be less than next_tx_index %d", queued.TxIndex, exported.NextTxIndex)
		}
		if queued.SourceLogicalTime != queued.Envelope.CreatedLogicalTime {
			return fmt.Errorf("queued message %d source logical time drift", queued.Sequence)
		}
		if queued.DestinationKey != string(queued.Envelope.Destination) {
			return fmt.Errorf("queued message %d destination key drift", queued.Sequence)
		}
		if i > 0 && queuedMessageLess(queued, exported.Queue[i-1]) {
			return fmt.Errorf("queued messages must be sorted by tx/message/logical/destination/sequence order")
		}
		if err := queued.Envelope.Validate(exported.Params); err != nil {
			return fmt.Errorf("invalid queued message %d: %w", queued.Sequence, err)
		}
	}
	for owner, messages := range map[string][]QueuedMessage(exported.Inbox) {
		if err := validateQueuedView("inbox", owner, messages, exported.Params); err != nil {
			return err
		}
	}
	for owner, messages := range map[string][]QueuedMessage(exported.Outbox) {
		if err := validateQueuedView("outbox", owner, messages, exported.Params); err != nil {
			return err
		}
	}
	return nil
}

func (e *Executor) Contract(address sdk.AccAddress) (ContractAccount, bool) {
	contract, ok := e.contracts[string(address)]
	return cloneContract(contract), ok
}

func (e *Executor) Queue() []QueuedMessage {
	return cloneQueuedMessages(e.queue)
}

func (e *Executor) Receipts() []ExecutionReceipt {
	return cloneReceipts(e.receipts)
}

func (e *Executor) Metrics() Observability {
	return e.metrics
}

func (e *Executor) processNext() (ExecutionReceipt, error) {
	queued := e.queue[0]
	e.queue = e.queue[1:]
	msg := queued.Envelope
	receipt := ExecutionReceipt{
		Sequence:       queued.Sequence,
		Source:         append(sdk.AccAddress(nil), msg.Source...),
		Destination:    append(sdk.AccAddress(nil), msg.Destination...),
		Opcode:         msg.Opcode,
		QueryID:        msg.QueryID,
		GasUsed:        e.params.ExecutionGasPerMessage,
		StorageFeeNaet: sdkmath.ZeroInt(),
		ForwardFeeNaet: msg.ForwardFee.Amount,
	}
	e.metrics.ProcessedMessages++
	e.metrics.GasUsed += e.params.ExecutionGasPerMessage

	if msg.DeadlineBlock != 0 && e.blockHeight > msg.DeadlineBlock {
		receipt.ResultCode = ResultExpired
		receipt.Error = "message expired"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}

	contract, ok := e.contracts[string(msg.Destination)]
	if !ok {
		receipt.ResultCode = ResultNoDestination
		receipt.Error = "destination contract not found"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}

	handler := e.handlers[string(msg.Destination)]
	if handler == nil {
		receipt.ResultCode = ResultExecutionFailed
		receipt.Error = "destination contract has no handler"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}

	working := cloneContract(contract)
	working.BalanceNaet = working.BalanceNaet.Add(msg.Value.Amount)
	working.LogicalTime++
	result := handler(working, cloneMessage(msg))
	if result.GasUsed > 0 {
		receipt.GasUsed = result.GasUsed
	}
	if receipt.GasUsed > msg.GasLimit {
		receipt.ResultCode = ResultLimitExceeded
		receipt.Error = "message gas limit exceeded"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}
	receipt.ResultCode = result.ResultCode
	if result.ResultCode != ResultOK {
		if result.Error != "" {
			receipt.Error = result.Error
		} else {
			receipt.Error = "contract execution failed"
		}
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}
	if len(result.NewState) > int(e.params.MaxStateSize) {
		receipt.ResultCode = ResultLimitExceeded
		receipt.Error = "contract state limit exceeded"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}
	if len(result.Outgoing) > int(e.params.MaxEmittedMessagesPerExec) {
		receipt.ResultCode = ResultLimitExceeded
		receipt.Error = "emitted message limit exceeded"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}
	if result.StorageWrites > e.params.MaxStorageWritesPerExec {
		receipt.ResultCode = ResultLimitExceeded
		receipt.Error = "storage write limit exceeded"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}
	working.State = append([]byte(nil), result.NewState...)
	receipt.StorageFeeNaet = e.params.StorageFeePerByte.MulRaw(int64(len(working.State)))
	working.BalanceNaet = working.BalanceNaet.Sub(receipt.StorageFeeNaet)
	if working.BalanceNaet.IsNegative() {
		receipt.ResultCode = ResultExecutionFailed
		receipt.Error = "insufficient naet for storage fee"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}
	e.contracts[string(working.Address)] = working
	outgoingTxIndex := e.nextTxIndex
	if len(result.Outgoing) > 0 {
		e.nextTxIndex++
	}
	for i, out := range result.Outgoing {
		out.Source = append(sdk.AccAddress(nil), working.Address...)
		out.CreatedLogicalTime = working.LogicalTime
		out.Depth = msg.Depth + 1
		if err := e.enqueueMessageWithOrder(out, outgoingTxIndex, uint32(i)); err != nil {
			receipt.ResultCode = ResultLimitExceeded
			receipt.Error = err.Error()
			e.metrics.FailedExecutions++
			e.finalizeFailure(msg, receipt)
			e.receipts = append(e.receipts, receipt)
			return receipt, nil
		}
	}
	e.receipts = append(e.receipts, receipt)
	return receipt, nil
}

func (e *Executor) finalizeFailure(msg MessageEnvelope, receipt ExecutionReceipt) {
	if msg.Bounce && !msg.Bounced {
		bounce := MessageEnvelope{
			Source:             append(sdk.AccAddress(nil), msg.Destination...),
			Destination:        append(sdk.AccAddress(nil), msg.Source...),
			Value:              sdk.NewCoin(appparams.BaseDenom, msg.Value.Amount),
			Opcode:             BounceOpcode,
			QueryID:            msg.QueryID,
			Body:               append([]byte(nil), msg.Body...),
			Bounce:             false,
			Bounced:            true,
			CreatedLogicalTime: msg.CreatedLogicalTime,
			DeadlineBlock:      msg.DeadlineBlock,
			GasLimit:           msg.GasLimit,
			ForwardFee:         sdk.NewCoin(appparams.BaseDenom, e.params.ForwardingFee),
			Depth:              msg.Depth + 1,
		}
		if err := e.EnqueueMessage(bounce); err == nil {
			e.metrics.BouncedMessages++
		}
		return
	}
	if msg.Bounced || msg.Opcode == RefundOpcode {
		return
	}
	if msg.Value.Amount.IsPositive() {
		refund := MessageEnvelope{
			Source:             append(sdk.AccAddress(nil), msg.Destination...),
			Destination:        append(sdk.AccAddress(nil), msg.Source...),
			Value:              sdk.NewCoin(appparams.BaseDenom, msg.Value.Amount),
			Opcode:             RefundOpcode,
			QueryID:            msg.QueryID,
			Body:               []byte("refund"),
			Bounce:             false,
			Bounced:            false,
			CreatedLogicalTime: msg.CreatedLogicalTime,
			DeadlineBlock:      0,
			GasLimit:           msg.GasLimit,
			ForwardFee:         sdk.NewCoin(appparams.BaseDenom, e.params.ForwardingFee),
			Depth:              msg.Depth + 1,
		}
		if err := e.EnqueueMessage(refund); err == nil {
			e.metrics.RefundMessages++
		}
	}
}

func (e *Executor) updateQueueLag() {
	if len(e.queue) == 0 {
		e.metrics.QueueLag = 0
		return
	}
	oldest := e.queue[0].EnqueuedBlock
	if e.blockHeight > oldest {
		e.metrics.QueueLag = e.blockHeight - oldest
		return
	}
	e.metrics.QueueLag = 0
}

func validateQueuedView(name, owner string, messages []QueuedMessage, params Params) error {
	if len(owner) == 0 {
		return fmt.Errorf("%s owner key must not be empty", name)
	}
	for _, queued := range messages {
		if err := queued.Envelope.Validate(params); err != nil {
			return fmt.Errorf("invalid %s message %d: %w", name, queued.Sequence, err)
		}
	}
	return nil
}

func cloneContract(contract ContractAccount) ContractAccount {
	return ContractAccount{
		Address:     append(sdk.AccAddress(nil), contract.Address...),
		CodeHash:    append([]byte(nil), contract.CodeHash...),
		State:       append([]byte(nil), contract.State...),
		BalanceNaet: contract.BalanceNaet,
		LogicalTime: contract.LogicalTime,
	}
}

func cloneMessage(msg MessageEnvelope) MessageEnvelope {
	msg.Source = append(sdk.AccAddress(nil), msg.Source...)
	msg.Destination = append(sdk.AccAddress(nil), msg.Destination...)
	msg.Body = append([]byte(nil), msg.Body...)
	return msg
}

func cloneQueuedMessages(messages []QueuedMessage) []QueuedMessage {
	if len(messages) == 0 {
		return nil
	}
	out := make([]QueuedMessage, len(messages))
	for i, msg := range messages {
		out[i] = QueuedMessage{
			TxIndex:           msg.TxIndex,
			MessageIndex:      msg.MessageIndex,
			SourceLogicalTime: msg.SourceLogicalTime,
			DestinationKey:    msg.DestinationKey,
			Sequence:          msg.Sequence,
			EnqueuedBlock:     msg.EnqueuedBlock,
			Envelope:          cloneMessage(msg.Envelope),
		}
	}
	return out
}

func cloneQueuedMap(in map[string][]QueuedMessage) map[string][]QueuedMessage {
	out := make(map[string][]QueuedMessage, len(in))
	for key, value := range in {
		out[key] = cloneQueuedMessages(value)
	}
	return out
}

func cloneReceipts(receipts []ExecutionReceipt) []ExecutionReceipt {
	if len(receipts) == 0 {
		return nil
	}
	out := make([]ExecutionReceipt, len(receipts))
	for i, receipt := range receipts {
		receipt.Source = append(sdk.AccAddress(nil), receipt.Source...)
		receipt.Destination = append(sdk.AccAddress(nil), receipt.Destination...)
		out[i] = receipt
	}
	return out
}

func writePart(write func([]byte) (int, error), bz []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(bz)))
	_, _ = write(length[:])
	_, _ = write(bz)
}

func queuedMessageLess(a, b QueuedMessage) bool {
	if a.TxIndex != b.TxIndex {
		return a.TxIndex < b.TxIndex
	}
	if a.MessageIndex != b.MessageIndex {
		return a.MessageIndex < b.MessageIndex
	}
	if a.SourceLogicalTime != b.SourceLogicalTime {
		return a.SourceLogicalTime < b.SourceLogicalTime
	}
	if a.DestinationKey != b.DestinationKey {
		return a.DestinationKey < b.DestinationKey
	}
	return a.Sequence < b.Sequence
}
