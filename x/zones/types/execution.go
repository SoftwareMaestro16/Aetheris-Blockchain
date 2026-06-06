package types

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

type ExecutionResult = ZoneReceipt
type MessageReceipt = ZoneReceipt

type ZoneExecutionMachine interface {
	ZoneID() ZoneID
	ExecuteTx(context.Context, ZoneTransaction) (ExecutionResult, error)
	ApplyMessage(context.Context, ZoneMessage) (MessageReceipt, error)
	BeginZoneBlock(context.Context) error
	EndZoneBlock(context.Context) (ZoneExecutionSummary, error)
	ExportZone(context.Context) (ZoneExport, error)
	ImportZone(context.Context, ZoneExport) error
	StateRoot(context.Context) (string, error)
}

type ZoneExecutionSummary struct {
	ZoneID                 ZoneID
	Height                 uint64
	TransactionsExecuted   uint32
	InboundMessagesApplied uint32
	ReceiptsProduced       uint32
	GasConsumed            uint64
	StateRoot              string
	InboxRoot              string
	OutboxRoot             string
	ReceiptRoot            string
	ExecutionResultRoot    string
	SummaryHash            string
}

func BeginZoneBlock(ctx context.Context, machine ZoneExecutionMachine) error {
	if machine == nil {
		return errors.New("zone execution machine is required")
	}
	return machine.BeginZoneBlock(ctx)
}

func ExecuteTx(ctx context.Context, machine ZoneExecutionMachine, tx ZoneTransaction) (ExecutionResult, error) {
	if machine == nil {
		return ExecutionResult{}, errors.New("zone execution machine is required")
	}
	if err := tx.Validate(machine.ZoneID()); err != nil {
		return ExecutionResult{}, err
	}
	result, err := machine.ExecuteTx(ctx, tx)
	if err != nil {
		return ExecutionResult{}, err
	}
	if result.ZoneID != machine.ZoneID() {
		return ExecutionResult{}, errors.New("zone execution result route mismatch")
	}
	if result.ItemHash != tx.TxHash {
		return ExecutionResult{}, errors.New("zone execution result item hash mismatch")
	}
	return result, result.Validate()
}

func ApplyMessage(ctx context.Context, machine ZoneExecutionMachine, msg ZoneMessage) (MessageReceipt, error) {
	if machine == nil {
		return MessageReceipt{}, errors.New("zone execution machine is required")
	}
	if err := msg.Validate(machine.ZoneID()); err != nil {
		return MessageReceipt{}, err
	}
	receipt, err := machine.ApplyMessage(ctx, msg)
	if err != nil {
		return MessageReceipt{}, err
	}
	if receipt.ZoneID != machine.ZoneID() {
		return MessageReceipt{}, errors.New("zone message receipt route mismatch")
	}
	if receipt.ItemHash != msg.PayloadHash {
		return MessageReceipt{}, errors.New("zone message receipt item hash mismatch")
	}
	return receipt, receipt.Validate()
}

func EndZoneBlock(ctx context.Context, machine ZoneExecutionMachine) (ZoneExecutionSummary, error) {
	if machine == nil {
		return ZoneExecutionSummary{}, errors.New("zone execution machine is required")
	}
	summary, err := machine.EndZoneBlock(ctx)
	if err != nil {
		return ZoneExecutionSummary{}, err
	}
	if summary.ZoneID != machine.ZoneID() {
		return ZoneExecutionSummary{}, errors.New("zone execution summary route mismatch")
	}
	return summary, summary.Validate()
}

func StateRoot(ctx context.Context, machine ZoneExecutionMachine) (string, error) {
	if machine == nil {
		return "", errors.New("zone execution machine is required")
	}
	root, err := machine.StateRoot(ctx)
	if err != nil {
		return "", err
	}
	if err := ValidateHash("zone state root", root); err != nil {
		return "", err
	}
	return root, nil
}

func NewZoneExecutionSummary(summary ZoneExecutionSummary) (ZoneExecutionSummary, error) {
	if summary.SummaryHash != "" {
		return ZoneExecutionSummary{}, errors.New("zone execution summary hash must be empty before construction")
	}
	if err := summary.ValidateFormat(); err != nil {
		return ZoneExecutionSummary{}, err
	}
	summary.SummaryHash = ComputeZoneExecutionSummaryHash(summary)
	return summary, summary.Validate()
}

func (s ZoneExecutionSummary) ValidateFormat() error {
	if err := ValidateZoneID(s.ZoneID); err != nil {
		return err
	}
	if s.Height == 0 {
		return errors.New("zone execution summary height must be positive")
	}
	for _, item := range []struct {
		name  string
		value string
	}{
		{name: "zone execution summary state root", value: s.StateRoot},
		{name: "zone execution summary inbox root", value: s.InboxRoot},
		{name: "zone execution summary outbox root", value: s.OutboxRoot},
		{name: "zone execution summary receipt root", value: s.ReceiptRoot},
		{name: "zone execution summary result root", value: s.ExecutionResultRoot},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if s.SummaryHash != "" {
		return ValidateHash("zone execution summary hash", s.SummaryHash)
	}
	return nil
}

func (s ZoneExecutionSummary) Validate() error {
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.SummaryHash == "" {
		return errors.New("zone execution summary hash is required")
	}
	if s.SummaryHash != ComputeZoneExecutionSummaryHash(s) {
		return errors.New("zone execution summary hash mismatch")
	}
	return nil
}

func ComputeZoneExecutionSummaryHash(summary ZoneExecutionSummary) string {
	h := sha256.New()
	writeRuntimePart(h, "aetheris-zone-execution-summary-v1")
	writeRuntimePart(h, string(summary.ZoneID))
	writeRuntimeUint64(h, summary.Height)
	writeRuntimeUint64(h, uint64(summary.TransactionsExecuted))
	writeRuntimeUint64(h, uint64(summary.InboundMessagesApplied))
	writeRuntimeUint64(h, uint64(summary.ReceiptsProduced))
	writeRuntimeUint64(h, summary.GasConsumed)
	writeRuntimePart(h, summary.StateRoot)
	writeRuntimePart(h, summary.InboxRoot)
	writeRuntimePart(h, summary.OutboxRoot)
	writeRuntimePart(h, summary.ReceiptRoot)
	writeRuntimePart(h, summary.ExecutionResultRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func BuildZoneExecutionSummary(height uint64, runtime ZoneRuntimeState, queues ZoneMessageQueues, transactionsExecuted uint32, inboundMessagesApplied uint32, receipts []ZoneReceipt, gasConsumed uint64) (ZoneExecutionSummary, error) {
	if err := runtime.Validate(); err != nil {
		return ZoneExecutionSummary{}, err
	}
	if err := queues.Validate(); err != nil {
		return ZoneExecutionSummary{}, err
	}
	if runtime.ZoneID != queues.ZoneID {
		return ZoneExecutionSummary{}, errors.New("zone execution summary queue route mismatch")
	}
	for _, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return ZoneExecutionSummary{}, err
		}
		if receipt.ZoneID != runtime.ZoneID || receipt.Height != height {
			return ZoneExecutionSummary{}, fmt.Errorf("zone execution summary receipt route mismatch for %s", receipt.ReceiptHash)
		}
	}
	return NewZoneExecutionSummary(ZoneExecutionSummary{
		ZoneID:                 runtime.ZoneID,
		Height:                 height,
		TransactionsExecuted:   transactionsExecuted,
		InboundMessagesApplied: inboundMessagesApplied,
		ReceiptsProduced:       uint32(len(receipts)),
		GasConsumed:            gasConsumed,
		StateRoot:              runtime.StateRoot,
		InboxRoot:              queues.InboxRoot(),
		OutboxRoot:             queues.OutboxRoot(),
		ReceiptRoot:            ComputeZoneReceiptRoot(receipts),
		ExecutionResultRoot:    ComputeZoneExecutionResultRoot(receipts),
	})
}
