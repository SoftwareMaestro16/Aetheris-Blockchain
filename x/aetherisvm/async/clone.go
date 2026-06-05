package async

import sdk "github.com/cosmos/cosmos-sdk/types"

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
