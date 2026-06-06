package types

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityStoreV2SpecPrimaryKeyLayout(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	auctionID := identityHash("auction-id")
	pathHash := identityHash("resolution-path")
	commitmentHash := identityHash("commitment")

	domainKey, err := IdentityStoreV2SpecDomainKey("alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecDomainsPrefix+"/"+nameHash, domainKey)
	nameKey, err := IdentityStoreV2SpecDomainNameKey("ALICE.AET")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecDomainNamesPrefix+"/alice.aet", nameKey)
	commitKey, err := IdentityStoreV2SpecCommitmentKey(commitmentHash)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecCommitmentsPrefix+"/"+commitmentHash, commitKey)
	nftKey, err := IdentityStoreV2SpecNFTBindingKey("domain", "alice")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecNFTBindingsPrefix+"/domain/alice", nftKey)
	nftByNameKey, err := IdentityStoreV2SpecNFTBindingByNameKey("alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecNFTBindingsByNamePrefix+"/"+nameHash, nftByNameKey)
	resolverKey, err := IdentityStoreV2SpecResolverKey("alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecResolversPrefix+"/"+nameHash, resolverKey)
	reverseKey, err := IdentityStoreV2SpecReverseKey(addr(1))
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2ReversePrefix+"/"+hex.EncodeToString(addr(1)), reverseKey)
	delegationKey, err := IdentityStoreV2SpecDelegationKey("alice.aet", addr(2), DelegationScopeResolverUpdate)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecDelegationsPrefix+"/"+nameHash+"/"+hex.EncodeToString(addr(2))+"/"+string(DelegationScopeResolverUpdate), delegationKey)
	subdomainKey, err := IdentityStoreV2SpecSubdomainKey("alice.aet", "api")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecSubdomainsPrefix+"/"+nameHash+"/"+identityHash("identity-v2-child-label", "api"), subdomainKey)
	auctionKey, err := IdentityStoreV2SpecAuctionKey(auctionID)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecAuctionsPrefix+"/"+auctionID, auctionKey)
	auctionByNameKey, err := IdentityStoreV2SpecAuctionByNameKey("alice.aet", auctionID)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecAuctionsByNamePrefix+"/"+nameHash+"/"+auctionID, auctionByNameKey)
	cacheKey, err := IdentityStoreV2SpecResolutionCacheKey("alice.aet", pathHash)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecResolutionCachePrefix+"/"+nameHash+"/"+pathHash, cacheKey)
	expiryKey, err := IdentityStoreV2SpecExpiryIndexKey(123, "alice.aet")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s/%020d/%s", IdentityStoreV2SpecExpiryIndexPrefix, uint64(123), nameHash), expiryKey)
	ownerKey, err := IdentityStoreV2SpecOwnerIndexKey(addr(3), "alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecOwnerIndexPrefix+"/"+hex.EncodeToString(addr(3))+"/"+nameHash, ownerKey)
	resolverIndexKey, err := IdentityStoreV2SpecResolverIndexKey(addr(4), "alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecResolverIndexPrefix+"/"+hex.EncodeToString(addr(4))+"/"+nameHash, resolverIndexKey)
}

func TestIdentityStoreV2SpecPerformanceAccessPatterns(t *testing.T) {
	direct, err := IdentityStoreV2SpecDirectResolverReadAccessSet("alice.aet")
	require.NoError(t, err)
	require.Len(t, direct.Reads, 1)
	require.True(t, strings.HasPrefix(direct.Reads[0], IdentityStoreV2SpecResolversPrefix+"/"))

	pathKeys, err := IdentityStoreV2SpecRecursiveResolutionPathKeys("api.dex.alice.aet")
	require.NoError(t, err)
	require.Len(t, pathKeys, 6)
	require.True(t, strings.HasPrefix(pathKeys[0], IdentityStoreV2SpecDomainsPrefix+"/"))
	require.True(t, strings.HasPrefix(pathKeys[1], IdentityStoreV2SpecResolversPrefix+"/"))
	require.True(t, strings.HasPrefix(pathKeys[4], IdentityStoreV2SpecDomainsPrefix+"/"))
	require.True(t, strings.HasPrefix(pathKeys[5], IdentityStoreV2SpecResolversPrefix+"/"))

	expiryPrefix, err := IdentityStoreV2SpecBoundedExpiryScanPrefix(123)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s/%020d", IdentityStoreV2SpecExpiryIndexPrefix, uint64(123)), expiryPrefix)
}
