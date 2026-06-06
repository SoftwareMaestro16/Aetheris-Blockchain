package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeNameDeterministic(t *testing.T) {
	left, err := NormalizeName(" Alice.AET. ", "aet")
	require.NoError(t, err)
	right, err := NormalizeName("alice", ".AET")
	require.NoError(t, err)
	require.Equal(t, "alice.aet", left)
	require.Equal(t, left, right)
}

func TestExpiredNameCannotResolveAsActive(t *testing.T) {
	record := NameRecord{Name: "alice.aet", ExpiryHeight: 10}
	require.True(t, IsActive(record, 9))
	require.False(t, IsActive(record, 10))
	require.False(t, IsActive(record, 11))
}

func TestOwnershipBindingInvariant(t *testing.T) {
	params := DefaultIdentityRootParams()
	params.NFTBindingEnabled = true
	state := EmptyIdentityRootState()
	state.Records = append(state.Records, NameRecord{
		Name:          "alice.aet",
		Owner:         "owner-a",
		ResolverRoot:  DefaultResolverRoot,
		ExpiryHeight:  100,
		RenewalHeight: 1,
		CreatedHeight: 1,
		UpdatedHeight: 1,
		NFTBinding: IdentityNFTBindingReference{
			Name:    "alice.aet",
			Enabled: true,
			ClassID: "identity",
			NFTID:   "alice",
			Owner:   "owner-b",
		},
	})

	require.ErrorContains(t, state.Validate(params), "NFT binding owner")
}
