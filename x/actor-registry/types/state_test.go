package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActorIDAndAddressDerivationDeterministic(t *testing.T) {
	code := DefaultRoot("code")
	left := DeriveActorID("owner", code, "salt")
	right := DeriveActorID("owner", code, "salt")
	require.Equal(t, left, right)
	require.Equal(t, DeriveContractAddress(left), DeriveContractAddress(right))
	require.NoError(t, ValidateHash("actor id", left))
}

func TestLogicalTimeMonotonic(t *testing.T) {
	next, err := NextLogicalTime(5, 0)
	require.NoError(t, err)
	require.Equal(t, uint64(6), next)

	_, err = NextLogicalTime(5, 5)
	require.ErrorContains(t, err, "monotonically")
	_, err = NextLogicalTime(5, 4)
	require.ErrorContains(t, err, "monotonically")
}

func TestFrozenAndDeletedPolicyHelpers(t *testing.T) {
	actor := ActorRecord{Status: ActorStatusFrozen}
	require.False(t, CanExecuteNormalMessage(actor))
	actor.Status = ActorStatusActive
	require.True(t, CanExecuteNormalMessage(actor))
	actor.Status = ActorStatusDeleted
	require.False(t, CanReceiveValue(actor, DefaultActorRegistryParams()))

	params := DefaultActorRegistryParams()
	params.DeletedValuePolicy = DeletedValuePolicyRefund
	require.True(t, CanReceiveValue(actor, params))
}
