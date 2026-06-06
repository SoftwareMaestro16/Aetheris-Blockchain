package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestRegisterValidatorWithValidKeySeparation(t *testing.T) {
	k := NewKeeper()
	validator, err := k.RegisterValidator(types.MsgRegisterValidator{
		Authority: prototype.DefaultAuthority,
		Validator: testValidator(0x11, "ed25519:validator-a"),
		Height:    1,
	})
	require.NoError(t, err)
	require.Equal(t, types.StatusCandidate, validator.Status)
	require.Equal(t, "ed25519:validator-a", validator.ConsensusPublicKey)
	require.NotEqual(t, validator.OperatorAddress, validator.ConsensusPublicKey)

	keys, found, err := k.ValidatorKeys(validator.OperatorAddress)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, validator.ConsensusPublicKey, keys.ConsensusPublicKey)
}

func TestRejectDuplicateConsensusKeys(t *testing.T) {
	k := NewKeeper()
	_, err := k.RegisterValidator(types.MsgRegisterValidator{
		Authority: prototype.DefaultAuthority,
		Validator: testValidator(0x11, "ed25519:duplicate"),
		Height:    1,
	})
	require.NoError(t, err)

	_, err = k.RegisterValidator(types.MsgRegisterValidator{
		Authority: prototype.DefaultAuthority,
		Validator: testValidator(0x22, "ed25519:duplicate"),
		Height:    2,
	})
	require.ErrorContains(t, err, "duplicate consensus key")
}

func TestRejectZeroWithdrawalAddress(t *testing.T) {
	k := NewKeeper()
	validator := testValidator(0x11, "ed25519:zero-withdrawal")
	validator.WithdrawalAddress = addressing.ZeroRawAddress

	_, err := k.RegisterValidator(types.MsgRegisterValidator{
		Authority: prototype.DefaultAuthority,
		Validator: validator,
		Height:    1,
	})
	require.ErrorContains(t, err, "must not be zero address")
}

func TestConsensusKeyRotationDelayIsEnforced(t *testing.T) {
	k := NewKeeper()
	validator := registerValidator(t, &k, 0x11, "ed25519:rotate-a")

	_, err := k.RotateConsensusKey(types.MsgRotateConsensusKey{
		Authority:             prototype.DefaultAuthority,
		OperatorAddress:       validator.OperatorAddress,
		NewConsensusPublicKey: "ed25519:rotate-b",
		ActivationHeight:      99,
		Height:                1,
	})
	require.ErrorContains(t, err, "delay")

	rotating, err := k.RotateConsensusKey(types.MsgRotateConsensusKey{
		Authority:             prototype.DefaultAuthority,
		OperatorAddress:       validator.OperatorAddress,
		NewConsensusPublicKey: "ed25519:rotate-b",
		ActivationHeight:      101,
		Height:                1,
	})
	require.NoError(t, err)
	require.Equal(t, "ed25519:rotate-a", rotating.ConsensusPublicKey)
	require.Equal(t, "ed25519:rotate-b", rotating.PendingConsensusPublicKey)

	_, applied, err := k.ApplyConsensusKeyRotation(validator.OperatorAddress, 100)
	require.ErrorContains(t, err, "delay")
	require.False(t, applied)

	rotated, applied, err := k.ApplyConsensusKeyRotation(validator.OperatorAddress, 101)
	require.NoError(t, err)
	require.True(t, applied)
	require.Equal(t, "ed25519:rotate-b", rotated.ConsensusPublicKey)
	require.Empty(t, rotated.PendingConsensusPublicKey)
}

func TestJailedValidatorCannotBecomeActive(t *testing.T) {
	k := NewKeeper()
	validator := registerValidator(t, &k, 0x11, "ed25519:jailed")
	_, err := k.SetValidatorStatus(prototype.DefaultAuthority, validator.OperatorAddress, types.StatusActive, 2)
	require.NoError(t, err)
	_, err = k.SetValidatorStatus(prototype.DefaultAuthority, validator.OperatorAddress, types.StatusJailed, 3)
	require.NoError(t, err)

	_, err = k.SetValidatorStatus(prototype.DefaultAuthority, validator.OperatorAddress, types.StatusActive, 4)
	require.ErrorContains(t, err, "invalid status transition")
}

func TestTombstonedValidatorCannotReregisterWithSameConsensusKey(t *testing.T) {
	k := NewKeeper()
	validator := registerValidator(t, &k, 0x11, "ed25519:tombstone")
	_, err := k.SetValidatorStatus(prototype.DefaultAuthority, validator.OperatorAddress, types.StatusTombstoned, 2)
	require.NoError(t, err)

	_, err = k.RegisterValidator(types.MsgRegisterValidator{
		Authority: prototype.DefaultAuthority,
		Validator: testValidator(0x22, "ed25519:tombstone"),
		Height:    3,
	})
	require.ErrorContains(t, err, "duplicate consensus key")
}

func TestExportImportPreservesHistory(t *testing.T) {
	source := NewKeeper()
	validator := registerValidator(t, &source, 0x11, "ed25519:history")
	_, err := source.UpdateValidatorMetadata(types.MsgUpdateValidatorMetadata{
		Authority:       prototype.DefaultAuthority,
		OperatorAddress: validator.OperatorAddress,
		Metadata:        `{"moniker":"a"}`,
		Height:          2,
	})
	require.NoError(t, err)
	_, err = source.SetValidatorCapabilities(types.MsgSetValidatorCapabilities{
		Authority:       prototype.DefaultAuthority,
		OperatorAddress: validator.OperatorAddress,
		Capabilities:    []string{"fast-sync", "mev-resistant"},
		Height:          3,
	})
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	history, found, err := source.ValidatorHistory(validator.OperatorAddress)
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, history, 3)

	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
	importedHistory, found, err := target.ValidatorHistory(validator.OperatorAddress)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, history, importedHistory)
}

func TestValidatorSecurityQueries(t *testing.T) {
	k := NewKeeper()
	validator := registerValidator(t, &k, 0x11, "ed25519:security")
	security, found, err := k.ValidatorSecurityStatus(validator.OperatorAddress)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, types.StatusCandidate, security.Status)

	performance, found, err := k.ValidatorPerformance(validator.OperatorAddress)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint32(9_000), performance.ReputationScore)
}

func TestMaliciousAuthorityCannotRegisterValidator(t *testing.T) {
	k := NewKeeper()
	_, err := k.RegisterValidator(types.MsgRegisterValidator{
		Authority: "4:0000000000000000000000000000000000000000000000000000000000000002",
		Validator: testValidator(0x11, "ed25519:bad-authority"),
		Height:    1,
	})
	require.ErrorContains(t, err, "governance authority")
}

func registerValidator(t *testing.T, k *Keeper, fill byte, consensusKey string) types.ValidatorRecord {
	t.Helper()
	validator, err := k.RegisterValidator(types.MsgRegisterValidator{
		Authority: prototype.DefaultAuthority,
		Validator: testValidator(fill, consensusKey),
		Height:    1,
	})
	require.NoError(t, err)
	return validator
}

func testValidator(fill byte, consensusKey string) types.ValidatorRecord {
	return types.ValidatorRecord{
		OperatorAddress:    testAddress(fill),
		ConsensusPublicKey: consensusKey,
		TreasuryAddress:    testAddress(fill + 1),
		WithdrawalAddress:  testAddress(fill + 2),
		EmergencyAddress:   testAddress(fill + 3),
		Metadata:           `{"moniker":"validator"}`,
		CommissionPolicy:   types.DefaultCommissionPolicy(),
		ReputationScore:    9_000,
		PerformanceScore:   8_500,
		Status:             types.StatusCandidate,
		Capabilities:       []string{"archive", "fast-sync"},
		SelfBond:           1_000_000,
		ExternalAuditFlags: []string{"soc2"},
		UptimeHistory: []types.UptimeSample{
			{Height: 1, UptimeBps: 9_900},
		},
		LatencyHistory: []types.LatencySample{
			{Height: 1, LatencyMs: 42},
		},
	}
}

func testAddress(fill byte) string {
	return addressing.FormatAccAddress(sdk.AccAddress(bytesOf(fill)))
}

func bytesOf(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
