package types

import (
	"encoding/hex"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestCrossZoneQueueStateKeysMatchSpec(t *testing.T) {
	params := testCrossZoneParams()
	msg := testCrossZoneMessage(t, params, 1, 7, []byte("payload"))
	sender := hex.EncodeToString(msg.Sender)
	msgID := hex.EncodeToString(msg.MessageID)

	outboxKey, err := CrossZoneOutboxKey(msg.SourceZone, msg.Sender, msg.SourceSequence)
	require.NoError(t, err)
	require.Equal(t, "messages/outbox/FINANCIAL_ZONE/"+sender+"/7", outboxKey)

	inboxKey, err := CrossZoneInboxKey(msg.DestinationZone, msg.Sender, msg.SourceSequence)
	require.NoError(t, err)
	require.Equal(t, "messages/inbox/CONTRACT_ZONE/"+sender+"/7", inboxKey)

	receiptKey, err := CrossZoneReceiptKey(msg.MessageID)
	require.NoError(t, err)
	require.Equal(t, "messages/receipts/"+msgID, receiptKey)

	nonceKey, err := CrossZoneNonceKey(msg.SourceZone, msg.Sender)
	require.NoError(t, err)
	require.Equal(t, "messages/nonces/FINANCIAL_ZONE/"+sender, nonceKey)

	replayKey, err := CrossZoneReplayKey(msg.MessageID)
	require.NoError(t, err)
	require.Equal(t, "messages/replay/"+msgID, replayKey)

	expiryKey, err := CrossZoneExpiryKey(msg.Deadline, msg.MessageID)
	require.NoError(t, err)
	require.Equal(t, "messages/expiry/100/"+msgID, expiryKey)
}

func TestCrossZoneQueueLifecycleAndReplayTombstone(t *testing.T) {
	params := testCrossZoneParams()
	first := testCrossZoneMessage(t, params, 1, 1, []byte("payload-1"))
	second := testCrossZoneMessage(t, params, 2, 2, []byte("payload-2"))
	state := CrossZoneQueueState{Height: 20, ParamsHash: zonestypes.EmptyRootHash()}

	state, err := EnqueueCrossZoneOutbox(state, first, params)
	require.NoError(t, err)
	state, err = EnqueueCrossZoneOutbox(state, second, params)
	require.NoError(t, err)
	require.Len(t, state.Outbox, 2)
	require.Equal(t, uint64(1), state.Outbox[0].Message.SourceSequence)
	require.Len(t, state.Expiry, 2)
	require.Equal(t, uint64(2), state.Nonces[0].Nonce)
	require.NotEmpty(t, state.StateRoot)

	duplicateNonce := testCrossZoneMessage(t, params, 2, 3, []byte("payload-3"))
	_, err = EnqueueCrossZoneOutbox(state, duplicateNonce, params)
	require.ErrorContains(t, err, "nonce")

	state, routed, err := RouteCrossZoneOutboxToInbox(state, first.MessageID, 21, params)
	require.NoError(t, err)
	require.Equal(t, CrossZoneQueueInbox, routed.Kind)
	require.Len(t, state.Outbox, 1)
	require.Len(t, state.Inbox, 1)

	receipt := receiptFromCrossZoneMessage(first, 22)
	state, err = RecordCrossZoneReceipt(state, receipt, params)
	require.NoError(t, err)
	require.Len(t, state.Inbox, 0)
	require.Len(t, state.Receipts, 1)
	require.Len(t, state.Replay, 1)
	require.Len(t, state.Expiry, 1)
	require.Equal(t, first.MessageID, state.Replay[0].MessageID)

	_, err = EnqueueCrossZoneOutbox(state, first, params)
	require.ErrorContains(t, err, "replay tombstone")
}

func TestCrossZoneQueueRootsAreDeterministic(t *testing.T) {
	params := testCrossZoneParams()
	first := testCrossZoneMessage(t, params, 1, 1, []byte("payload-1"))
	second := testCrossZoneMessage(t, params, 2, 2, []byte("payload-2"))
	second.Sender = addr(3)
	second.MessageID = nil
	second.PayloadHash = nil
	second, err := NewCrossZoneMessageEnvelope(second, params)
	require.NoError(t, err)

	stateA := CrossZoneQueueState{}
	stateA, err = EnqueueCrossZoneOutbox(stateA, second, params)
	require.NoError(t, err)
	stateA, err = EnqueueCrossZoneOutbox(stateA, first, params)
	require.NoError(t, err)

	stateB := CrossZoneQueueState{}
	stateB, err = EnqueueCrossZoneOutbox(stateB, first, params)
	require.NoError(t, err)
	stateB, err = EnqueueCrossZoneOutbox(stateB, second, params)
	require.NoError(t, err)

	rootsA, err := ComputeCrossZoneQueueRoots(stateA, params)
	require.NoError(t, err)
	rootsB, err := ComputeCrossZoneQueueRoots(stateB, params)
	require.NoError(t, err)
	require.Equal(t, rootsA.OutboxRoot, rootsB.OutboxRoot)
	require.Equal(t, rootsA.NonceRoot, rootsB.NonceRoot)
	require.Equal(t, rootsA.ExpiryRoot, rootsB.ExpiryRoot)
	require.Equal(t, rootsA.StateRoot, rootsB.StateRoot)
	require.NoError(t, zonestypes.ValidateHash("cross-zone queue state root", rootsA.StateRoot))
}

func TestCrossZoneReceiptAndTombstoneHashesValidate(t *testing.T) {
	params := testCrossZoneParams()
	msg := testCrossZoneMessage(t, params, 9, 9, []byte("payload"))

	receipt, err := NewCrossZoneMessageReceipt(receiptFromCrossZoneMessage(msg, 30))
	require.NoError(t, err)
	require.Equal(t, ComputeCrossZoneReceiptHash(receipt), receipt.ReceiptHash)

	tombstone, err := NewCrossZoneReplayTombstone(CrossZoneReplayTombstone{
		MessageID:       msg.MessageID,
		SourceZone:      msg.SourceZone,
		Sender:          msg.Sender,
		Nonce:           msg.Nonce,
		SourceSequence:  msg.SourceSequence,
		CreatedHeight:   msg.CreatedHeight,
		TombstoneHeight: 30,
		ExpiryHeight:    msg.Deadline,
	})
	require.NoError(t, err)
	require.Equal(t, ComputeCrossZoneReplayTombstoneHash(tombstone), tombstone.TombstoneHash)
}

func receiptFromCrossZoneMessage(msg CrossZoneMessageEnvelope, height uint64) CrossZoneMessageReceipt {
	return CrossZoneMessageReceipt{
		MessageID:       msg.MessageID,
		SourceZone:      msg.SourceZone,
		DestinationZone: msg.DestinationZone,
		Sender:          msg.Sender,
		Recipient:       msg.Recipient,
		Status:          CrossZoneReceiptSuccess,
		GasUsed:         msg.GasLimit,
		ResultHash:      hashCrossZoneQueueParts("result", hex.EncodeToString(msg.MessageID)),
		Height:          height,
		SourceSequence:  msg.SourceSequence,
		Nonce:           msg.Nonce,
	}
}

func TestCrossZoneQueueRejectsLowFeeMessage(t *testing.T) {
	params := testCrossZoneParams()
	msg := testCrossZoneMessage(t, params, 1, 1, []byte("payload"))
	msg.FeeLimit = sdkmath.ZeroInt()
	msg.MessageID = nil
	msg.PayloadHash = nil
	msg, err := NewCrossZoneMessageEnvelope(msg, params)
	require.ErrorContains(t, err, "fee limit")
	require.Empty(t, msg.MessageID)
}
