package types

import (
	"bytes"
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeRecordSignatureIdentityAndExpiry(t *testing.T) {
	salt := []byte("aetheris-test-network")
	record := signedNodeRecord(t, 0x11, salt, 100, NodeRoleFull, NodeRoleRouting)

	require.NoError(t, record.Validate(salt, 99))

	expired := record
	require.ErrorContains(t, expired.Validate(salt, 101), "expired")

	tampered := record
	tampered.Roles = append(tampered.Roles, NodeRoleStorageProvider)
	require.ErrorContains(t, tampered.Validate(salt, 99), "signature")

	wrongSalt := []byte("wrong-network")
	require.ErrorContains(t, record.Validate(wrongSalt, 99), "node id")
}

func TestNetworkAddressHashCanonicalizesOffchainAddresses(t *testing.T) {
	left, err := HashNetworkAddresses([]string{
		"tcp://10.0.0.2:26656",
		"tcp://10.0.0.1:26656",
		"tcp://10.0.0.1:26656",
	})
	require.NoError(t, err)
	right, err := HashNetworkAddresses([]string{
		"tcp://10.0.0.1:26656",
		"tcp://10.0.0.2:26656",
	})
	require.NoError(t, err)

	require.Equal(t, left, right)
}

func TestSessionNegotiationCreatesDeterministicStreams(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x21, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x22, salt, 100, NodeRoleService)

	req := SessionRequest{
		LocalNodeID:      local.NodeID,
		RemoteNodeID:     remote.NodeID,
		ProtocolVersions: []string{DefaultProtocolVersion},
		ChannelClasses:   []ChannelClass{ChannelService, ChannelConsensus, ChannelData},
		OpenedHeight:     10,
		ExpiresHeight:    50,
		Nonce:            []byte("session-nonce"),
		QOSPolicy:        QoSPolicyConsensusFirst,
	}
	session, err := NegotiateSession(local, remote, req)
	require.NoError(t, err)
	require.NoError(t, session.Validate())
	require.Equal(t, ChannelConsensus, session.Streams[0].Channel)
	require.Equal(t, ChannelData, session.Streams[len(session.Streams)-1].Channel)

	again, err := NegotiateSession(local, remote, req)
	require.NoError(t, err)
	require.Equal(t, session, again)

	req.ProtocolVersions = []string{"unsupported"}
	_, err = NegotiateSession(local, remote, req)
	require.ErrorContains(t, err, "protocol")
}

func TestConsensusEnvelopeOutranksServiceAndBulkData(t *testing.T) {
	policies := DefaultChannelPolicies()
	service := TransportEnvelope{
		Channel:        ChannelService,
		SizeBytes:      512,
		EnqueuedHeight: 1,
		Sequence:       1,
		PayloadHash:    HashParts("service"),
	}
	data := TransportEnvelope{
		Channel:        ChannelData,
		SizeBytes:      2 << 20,
		EnqueuedHeight: 1,
		Sequence:       2,
		PayloadHash:    HashParts("data"),
	}
	consensus := TransportEnvelope{
		Channel:        ChannelConsensus,
		SizeBytes:      128,
		EnqueuedHeight: 100,
		Sequence:       99,
		PayloadHash:    HashParts("consensus"),
	}

	next, found, err := SelectNextEnvelope([]TransportEnvelope{service, data, consensus}, policies)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, ChannelConsensus, next.Channel)
}

func TestChunkPayloadRoundTripAndCorruptionDetection(t *testing.T) {
	payload := bytes.Repeat([]byte("aetheris-networking"), 512)
	chunks, err := ChunkPayload(payload, 257)
	require.NoError(t, err)
	require.Greater(t, len(chunks), 1)

	reordered := append([]PayloadChunk(nil), chunks...)
	for i, j := 0, len(reordered)-1; i < j; i, j = i+1, j-1 {
		reordered[i], reordered[j] = reordered[j], reordered[i]
	}
	decoded, err := ReassemblePayload(reordered)
	require.NoError(t, err)
	require.Equal(t, payload, decoded)

	reordered[0].Bytes[0] ^= 0xff
	_, err = ReassemblePayload(reordered)
	require.ErrorContains(t, err, "chunk hash")
}

func TestNetworkingStateRegistersNodesAndSessionsCanonically(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x31, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x32, salt, 100, NodeRoleService)

	state := EmptyState()
	var err error
	state, err = RegisterNodeRecord(state, remote, salt, 10)
	require.NoError(t, err)
	state, err = RegisterNodeRecord(state, local, salt, 10)
	require.NoError(t, err)
	require.Len(t, state.NodeRecords, 2)
	require.Less(t, state.NodeRecords[0].NodeID, state.NodeRecords[1].NodeID)

	session, err := NegotiateSession(local, remote, SessionRequest{
		LocalNodeID:      local.NodeID,
		RemoteNodeID:     remote.NodeID,
		ProtocolVersions: []string{DefaultProtocolVersion},
		OpenedHeight:     11,
		ExpiresHeight:    50,
		Nonce:            []byte("nonce"),
	})
	require.NoError(t, err)
	state, err = OpenSession(state, session, 12)
	require.NoError(t, err)
	require.Len(t, state.Sessions, 1)

	_, err = OpenSession(state, session, 12)
	require.ErrorContains(t, err, "already exists")

	pruned, err := PruneExpired(state, 101)
	require.NoError(t, err)
	require.Empty(t, pruned.NodeRecords)
	require.Empty(t, pruned.Sessions)
}

func signedNodeRecord(t *testing.T, seed byte, salt []byte, expiresHeight uint64, roles ...NodeRole) NodeRecord {
	t.Helper()

	privateKey := deterministicPrivateKey(seed)
	addressHash, err := HashNetworkAddresses([]string{"tcp://127.0.0.1:26656"})
	require.NoError(t, err)
	record, err := SignNodeRecord(NodeRecord{
		Roles:                roles,
		NetworkAddressesHash: addressHash,
		ZonesSupported:       []string{"APPLICATION_ZONE", "FINANCIAL_ZONE"},
		ServicesSupported:    []string{"state-sync", "execution-stream"},
		ProtocolVersions:     []string{DefaultProtocolVersion},
		ExpiresHeight:        expiresHeight,
	}, privateKey, salt)
	require.NoError(t, err)
	return record
}

func deterministicPrivateKey(fill byte) ed25519.PrivateKey {
	seed := bytes.Repeat([]byte{fill}, ed25519.SeedSize)
	return ed25519.NewKeyFromSeed(seed)
}
