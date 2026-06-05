package types

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestNewEventCanonicalizesAndValidates(t *testing.T) {
	event, err := NewEvent(EventTransfer, CategoryProtocol, []byte("tx"), 2, 1, addr(1), []Attribute{
		{Key: "to", Value: "b"},
		{Key: "from", Value: "a"},
	})
	require.NoError(t, err)
	require.Equal(t, []Attribute{{Key: "from", Value: "a"}, {Key: "to", Value: "b"}}, event.Attributes)
	require.NoError(t, event.Validate())

	_, err = NewEvent("bad", CategoryProtocol, nil, 0, 0, nil, nil)
	require.ErrorContains(t, err, "event type")
	_, err = NewEvent(EventTransfer, "bad", nil, 0, 0, nil, nil)
	require.ErrorContains(t, err, "event category")
	_, err = NewEvent(EventTransfer, CategoryProtocol, nil, 0, 0, sdk.AccAddress(make([]byte, 20)), nil)
	require.ErrorContains(t, err, "event actor")
}

func TestSortEventsDeterministic(t *testing.T) {
	events := []ProtocolEvent{
		mustEvent(t, EventMemoAttached, 2, 2, "b"),
		mustEvent(t, EventTransfer, 1, 1, "c"),
		mustEvent(t, EventFeeDistributed, 1, 1, "a"),
	}
	sorted := SortEvents(events)
	require.Equal(t, EventFeeDistributed, sorted[0].Type)
	require.Equal(t, EventTransfer, sorted[1].Type)
	require.Equal(t, EventMemoAttached, sorted[2].Type)
}

func TestAllRequiredEventTypesExist(t *testing.T) {
	for _, eventType := range []string{
		EventTransfer,
		EventMemoAttached,
		EventDomainAuctionStarted,
		EventDomainResolved,
		EventContractMessageQueued,
		EventContractMessageProcessed,
		EventReputationUpdated,
		EventFeeDistributed,
	} {
		require.True(t, IsEventType(eventType), eventType)
	}
}

func mustEvent(t *testing.T, eventType string, height uint64, sequence uint64, tx string) ProtocolEvent {
	t.Helper()
	event, err := NewEvent(eventType, CategoryProtocol, []byte(tx), height, sequence, nil, nil)
	require.NoError(t, err)
	return event
}

func addr(seed byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{seed}, 20))
}
