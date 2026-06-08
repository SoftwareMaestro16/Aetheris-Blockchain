package async

import (
	"errors"
	"fmt"
	"sort"
)

func (e *Executor) EnqueueTxMessages(messages []MessageEnvelope) error {
	if err := validateMessageBatch(e.params, messages); err != nil {
		return err
	}
	if err := e.validateQueueCapacity(messages); err != nil {
		return err
	}
	txIndex := e.nextTxIndex
	e.nextTxIndex++
	for i, msg := range messages {
		if _, err := e.enqueueMessageWithOrder(msg, e.blockHeight, txIndex, uint32(i)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Executor) EnqueueMessage(msg MessageEnvelope) error {
	_, err := e.enqueueSingleMessage(msg)
	return err
}

func (e *Executor) enqueueSingleMessage(msg MessageEnvelope) (QueuedMessage, error) {
	if err := validateMessageBatch(e.params, []MessageEnvelope{msg}); err != nil {
		return QueuedMessage{}, err
	}
	if err := e.validateQueueCapacity([]MessageEnvelope{msg}); err != nil {
		return QueuedMessage{}, err
	}
	txIndex := e.nextTxIndex
	e.nextTxIndex++
	return e.enqueueMessageWithOrder(msg, e.blockHeight, txIndex, 0)
}

func (e *Executor) enqueueMessageWithOrder(msg MessageEnvelope, txHeight, txIndex uint64, messageIndex uint32) (QueuedMessage, error) {
	queued := buildQueuedMessage(msg, txHeight, txIndex, messageIndex, e.nextSequence)
	if err := validateQueuedMessage(queued, e.params); err != nil {
		return QueuedMessage{}, err
	}
	e.nextSequence++
	e.queue = append(e.queue, queued)
	sort.SliceStable(e.queue, func(i, j int) bool {
		return queuedMessageLess(e.queue[i], e.queue[j])
	})
	destinationKey := inboxKey(queued.Envelope.Destination)
	sourceKey := outboxKey(queued.Envelope.Source)
	e.inbox[destinationKey] = append(e.inbox[destinationKey], queued)
	e.outbox[sourceKey] = append(e.outbox[sourceKey], queued)
	sort.SliceStable(e.inbox[destinationKey], func(i, j int) bool {
		return queuedMessageLess(e.inbox[destinationKey][i], e.inbox[destinationKey][j])
	})
	sort.SliceStable(e.outbox[sourceKey], func(i, j int) bool {
		return queuedMessageLess(e.outbox[sourceKey][i], e.outbox[sourceKey][j])
	})
	e.metrics.QueuedMessages++
	return queued, nil
}

func (e *Executor) validateQueueCapacity(messages []MessageEnvelope) error {
	counts := make(map[string]uint32)
	for _, queued := range e.queue {
		counts[queueAddressKey(queued.Envelope.Destination)]++
	}
	for _, msg := range messages {
		key := queueAddressKey(msg.Destination)
		counts[key]++
		if counts[key] > e.params.MaxQueuedMessagesPerContract {
			return fmt.Errorf("queued messages per contract must be <= %d", e.params.MaxQueuedMessagesPerContract)
		}
	}
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
		if readyBlock(e.queue[0]) > height {
			break
		}
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

func (e *Executor) updateQueueLag() {
	if len(e.queue) == 0 {
		e.metrics.QueueLag = 0
		return
	}
	oldest := e.queue[0].EnqueuedBlock
	if readyBlock(e.queue[0]) > e.blockHeight {
		e.metrics.QueueLag = 0
		return
	}
	if e.blockHeight > oldest {
		e.metrics.QueueLag = e.blockHeight - oldest
		return
	}
	e.metrics.QueueLag = 0
}
