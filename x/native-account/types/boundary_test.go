package types

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultBoundariesValidateAndDeclareNativeAccountOwner(t *testing.T) {
	boundaries := DefaultBoundaries()

	require.NoError(t, ValidateBoundaries(boundaries))

	var native Boundary
	for _, boundary := range boundaries {
		if boundary.Path == ModulePath {
			native = boundary
			break
		}
	}
	require.Equal(t, ModulePath, native.Path)
	require.Contains(t, native.Owner, "native account state")
	require.Contains(t, native.Owner, "auth policy")
	require.Contains(t, native.OwnedState, "account/by_user")
	require.Contains(t, native.OwnedState, "account/storage")
	require.Contains(t, native.RejectedWrites, "private keys")
	require.Contains(t, native.RejectedWrites, "seed phrases")
	require.Contains(t, native.RejectedWrites, "private keys")
}

func TestRejectedCrossModuleWritesCoverSecurityBoundaries(t *testing.T) {
	require.True(t, IsRejectedCrossModuleWrite("app/addressing", ModulePath, "account state"))
	require.True(t, IsRejectedCrossModuleWrite("x/identity", ModulePath, "auth policy"))
	require.True(t, IsRejectedCrossModuleWrite("x/storage-rent", ModulePath, "automatic wallet deletion"))
	require.True(t, IsRejectedCrossModuleWrite("x/storage-rent", "protocol-critical/system state", "rent freeze"))
	require.True(t, IsRejectedCrossModuleWrite("x/nominator-pool", "x/validator-*", "user-selected validator delegation"))
	require.True(t, IsRejectedCrossModuleWrite("x/fees", ModulePath, "duplicated wallet balance"))
	require.True(t, IsRejectedCrossModuleWrite("x/contracts, x/vm, x/aetravm/*", ModulePath, "sequence bypass"))

	require.False(t, IsRejectedCrossModuleWrite(ModulePath, "x/reputation", "reputation id reference"))
}

func TestDefaultAssetRoutesPassValidation(t *testing.T) {
	routes := DefaultAssetRoutes()

	require.NoError(t, ValidateAssetRoutes(routes))

	for _, route := range routes {
		require.NotEmpty(t, route.Behavior)
		require.NotEmpty(t, route.Route)
	}
	require.NoError(t, ValidateNoNativeAssetModules([]string{"auth", "bank", "staking", ModuleName, "fees"}))
}

func TestBoundaryManifestIsDeterministic(t *testing.T) {
	lines := BoundaryManifestLines()
	require.NotEmpty(t, lines)

	hash := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	require.Equal(t, "a5edcd1809b0873772833b4968666eb77dee250f850bd0c240e4159f46478d74", hex.EncodeToString(hash[:]))
}
