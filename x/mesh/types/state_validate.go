package types

import (
	"errors"
	"fmt"
	"sort"
)

func validateDestinations(destinations []MeshDestination) error {
	seen := make(map[string]struct{}, len(destinations))
	for i, destination := range destinations {
		if err := destination.Validate(); err != nil {
			return err
		}
		key := destinationKey(destination.ZoneID, destination.ShardID)
		if _, found := seen[key]; found {
			return errors.New("duplicate mesh destination")
		}
		seen[key] = struct{}{}
		if i > 0 && compareDestinations(destinations[i-1], destination) >= 0 {
			return errors.New("mesh destinations must be sorted canonically")
		}
	}
	return nil
}

func validateCommitments(commitments []FinalizedCommitment) error {
	seen := make(map[string]struct{}, len(commitments))
	for i, commitment := range commitments {
		if err := commitment.Validate(); err != nil {
			return err
		}
		key := commitmentKey(commitment.ZoneID, commitment.ShardID, commitment.Height)
		if _, found := seen[key]; found {
			return errors.New("duplicate mesh finalized commitment")
		}
		seen[key] = struct{}{}
		if i > 0 && compareCommitments(commitments[i-1], commitment) >= 0 {
			return errors.New("mesh finalized commitments must be sorted canonically")
		}
	}
	return nil
}

func validateReplayMarkers(markers []ReplayMarker) error {
	seen := make(map[string]struct{}, len(markers))
	for i, marker := range markers {
		if err := marker.Validate(); err != nil {
			return err
		}
		if _, found := seen[marker.MessageID]; found {
			return errors.New("duplicate mesh replay marker")
		}
		seen[marker.MessageID] = struct{}{}
		if i > 0 && markers[i-1].MessageID >= marker.MessageID {
			return errors.New("mesh replay markers must be sorted canonically")
		}
	}
	return nil
}

func validateReceipts(receipts []MeshReceipt) error {
	seen := make(map[string]struct{}, len(receipts))
	for i, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if _, found := seen[receipt.MessageID]; found {
			return errors.New("duplicate mesh receipt")
		}
		seen[receipt.MessageID] = struct{}{}
		if i > 0 && receipts[i-1].MessageID >= receipt.MessageID {
			return errors.New("mesh receipts must be sorted canonically")
		}
	}
	return nil
}

func validateBounceReceipts(receipts []BounceReceipt) error {
	seen := make(map[string]struct{}, len(receipts))
	for i, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if _, found := seen[receipt.MessageID]; found {
			return errors.New("duplicate mesh bounce receipt")
		}
		seen[receipt.MessageID] = struct{}{}
		if i > 0 && receipts[i-1].MessageID >= receipt.MessageID {
			return errors.New("mesh bounce receipts must be sorted canonically")
		}
	}
	return nil
}

func validateRefundReceipts(receipts []RefundReceipt) error {
	seen := make(map[string]struct{}, len(receipts))
	for i, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if _, found := seen[receipt.MessageID]; found {
			return errors.New("duplicate mesh refund receipt")
		}
		seen[receipt.MessageID] = struct{}{}
		if i > 0 && receipts[i-1].MessageID >= receipt.MessageID {
			return errors.New("mesh refund receipts must be sorted canonically")
		}
	}
	return nil
}

func findCommitment(state MeshState, zone ZoneID, shard ShardID, height uint64) (FinalizedCommitment, bool) {
	for _, commitment := range state.FinalizedCommitments {
		if commitment.ZoneID == zone && commitment.ShardID == shard && commitment.Height == height {
			return commitment, true
		}
	}
	return FinalizedCommitment{}, false
}

func hasDestination(state MeshState, zone ZoneID, shard ShardID) bool {
	for _, destination := range state.Destinations {
		if destination.ZoneID == zone && destination.ShardID == shard {
			return true
		}
	}
	return false
}

func hasActiveDestination(state MeshState, zone ZoneID, shard ShardID) bool {
	for _, destination := range state.Destinations {
		if destination.ZoneID == zone && destination.ShardID == shard && destination.Active {
			return true
		}
	}
	return false
}

func hasCommitment(state MeshState, zone ZoneID, shard ShardID, height uint64) bool {
	_, found := findCommitment(state, zone, shard, height)
	return found
}

func hasReplayMarker(state MeshState, messageID string) bool {
	for _, marker := range state.ReplayMarkers {
		if marker.MessageID == messageID {
			return true
		}
	}
	return false
}

func hasReceipt(state MeshState, messageID string) bool {
	for _, receipt := range state.Receipts {
		if receipt.MessageID == messageID {
			return true
		}
	}
	return false
}

func sortDestinations(destinations []MeshDestination) {
	sort.SliceStable(destinations, func(i, j int) bool {
		return compareDestinations(destinations[i], destinations[j]) < 0
	})
}

func sortCommitments(commitments []FinalizedCommitment) {
	sort.SliceStable(commitments, func(i, j int) bool {
		return compareCommitments(commitments[i], commitments[j]) < 0
	})
}

func sortReplayMarkers(markers []ReplayMarker) {
	sort.SliceStable(markers, func(i, j int) bool {
		return markers[i].MessageID < markers[j].MessageID
	})
}

func sortReceipts(receipts []MeshReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		return receipts[i].MessageID < receipts[j].MessageID
	})
}

func sortBounceReceipts(receipts []BounceReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		return receipts[i].MessageID < receipts[j].MessageID
	})
}

func sortRefundReceipts(receipts []RefundReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		return receipts[i].MessageID < receipts[j].MessageID
	})
}

func compareDestinations(left, right MeshDestination) int {
	if left.ZoneID != right.ZoneID {
		return compareString(string(left.ZoneID), string(right.ZoneID))
	}
	return compareString(string(left.ShardID), string(right.ShardID))
}

func compareCommitments(left, right FinalizedCommitment) int {
	if left.ZoneID != right.ZoneID {
		return compareString(string(left.ZoneID), string(right.ZoneID))
	}
	if left.ShardID != right.ShardID {
		return compareString(string(left.ShardID), string(right.ShardID))
	}
	return compareUint64(left.Height, right.Height)
}

func destinationKey(zone ZoneID, shard ShardID) string {
	return fmt.Sprintf("%s/%s", zone, shard)
}

func commitmentKey(zone ZoneID, shard ShardID, height uint64) string {
	return fmt.Sprintf("%s/%s/%020d", zone, shard, height)
}

func maxUint64(left, right uint64) uint64 {
	if left > right {
		return left
	}
	return right
}
