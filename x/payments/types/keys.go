package types

import (
	"fmt"
	"strings"
)

const (
	ModuleName = "payments"
	StoreKey   = ModuleName

	PaymentsKeyChannelPrefix             = "channel"
	PaymentsKeyParticipantIndexPrefix    = "participant_index"
	PaymentsKeyPendingCloseIndexPrefix   = "pending_close"
	PaymentsKeyConditionIndexPrefix      = "condition"
	PaymentsKeyRoutingAdIndexPrefix      = "routing_ad"
	PaymentsKeySettlementTombstonePrefix = "settlement_tombstone"
	PaymentsKeySettlementPrefix          = "settlement"
	PaymentsKeyCustodyPrefix             = "custody"
	PaymentsKeyBlockAccumulatorPrefix    = "block_accumulator"
)

func PaymentChannelKey(channelID string) string {
	return paymentKey(PaymentsKeyChannelPrefix, normalizeHash(channelID))
}

func PaymentParticipantIndexKey(participant, channelID string) string {
	return paymentKey(PaymentsKeyParticipantIndexPrefix, strings.TrimSpace(participant), normalizeHash(channelID))
}

func PaymentPendingCloseIndexKey(channelID string) string {
	return paymentKey(PaymentsKeyPendingCloseIndexPrefix, normalizeHash(channelID))
}

func PaymentConditionIndexKey(channelID, conditionID string) string {
	return paymentKey(PaymentsKeyConditionIndexPrefix, normalizeHash(channelID), normalizeHash(conditionID))
}

func PaymentRoutingAdvertisementIndexKey(channelID string) string {
	return paymentKey(PaymentsKeyRoutingAdIndexPrefix, normalizeHash(channelID))
}

func PaymentSettlementTombstoneKey(channelID string) string {
	return paymentKey(PaymentsKeySettlementTombstonePrefix, normalizeHash(channelID))
}

func PaymentSettlementKey(channelID string) string {
	return paymentKey(PaymentsKeySettlementPrefix, normalizeHash(channelID))
}

func PaymentCustodyKey(channelID string) string {
	return paymentKey(PaymentsKeyCustodyPrefix, normalizeHash(channelID))
}

func PaymentBlockAccumulatorKey(blockHeight uint64) string {
	return paymentKey(PaymentsKeyBlockAccumulatorPrefix, fmt.Sprintf("%020d", blockHeight))
}

func paymentKey(parts ...string) string {
	out := []string{ModuleName}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, "/")
}
