package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPersistentStateRentAccruesForActiveContractAndLongLivedRecords(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteBlock = 2

	fixtures := []PersistentStateRecord{
		{SubjectID: "contract-1", Class: StateClassContract, CodeBytes: 10, DataBytes: 20, IndexBytes: 5, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 7},
		{SubjectID: "domain-1", Class: StateClassDomainRecord, DataBytes: 35, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 7},
		{SubjectID: "rep-1", Class: StateClassStakingReputation, DataBytes: 30, IndexBytes: 5, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 7},
	}

	for _, fixture := range fixtures {
		result, err := AccruePersistentStateRent(fixture, params, 9)
		require.NoError(t, err, fixture.SubjectID)
		require.Equal(t, uint64(35), result.StorageBytes, fixture.SubjectID)
		require.Equal(t, uint64(140), result.RentDelta, fixture.SubjectID)
		require.Equal(t, uint64(140), result.Subject.RentDebt, fixture.SubjectID)
		require.Equal(t, uint64(9), result.Subject.LastChargedHeight, fixture.SubjectID)
	}
}

func TestPersistentStateRentSkipsUnactivatedEmptyAndDeletedState(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteBlock = 10

	unactivated := PersistentStateRecord{SubjectID: "virtual-ae", Class: StateClassWallet, Persistent: false, LastChargedHeight: 1}
	result, err := AccruePersistentStateRent(unactivated, params, 50)
	require.NoError(t, err)
	require.Zero(t, result.RentDelta)
	require.Zero(t, result.Subject.RentDebt)

	empty := PersistentStateRecord{SubjectID: "empty", Class: StateClassContract, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 1}
	result, err = AccruePersistentStateRent(empty, params, 50)
	require.NoError(t, err)
	require.Zero(t, result.RentDelta)
	require.Zero(t, result.Subject.RentDebt)

	deleted := PersistentStateRecord{SubjectID: "deleted", Class: StateClassContract, DataBytes: 100, Persistent: true, Status: ContractStatusDeleted, LastChargedHeight: 1}
	result, err = AccruePersistentStateRent(deleted, params, 50)
	require.NoError(t, err)
	require.Zero(t, result.RentDelta)
	require.Zero(t, result.Subject.RentDebt)
}

func TestWalletEffectiveFeeIncludesGasRentAndUnpaidDebt(t *testing.T) {
	fee, err := EffectiveWalletFee(3, 5, 7)
	require.NoError(t, err)
	require.Equal(t, uint64(15), fee)

	_, err = EffectiveWalletFee(^uint64(0), 1, 0)
	require.ErrorContains(t, err, "overflow")
}

func TestPoolRentUsesProtocolReservesAndFrozenLimitedAllowsOnlyRecoveryActions(t *testing.T) {
	payer, remaining := PayPoolRent(PoolRentPayer{
		ProtocolFeeReserve: 30,
		GovernanceReserve:  15,
		UserFacingCharge:   99,
	}, 40)
	require.Zero(t, remaining)
	require.Equal(t, uint64(0), payer.ProtocolFeeReserve)
	require.Equal(t, uint64(5), payer.GovernanceReserve)
	require.Equal(t, uint64(99), payer.UserFacingCharge)

	pool := PersistentStateRecord{Class: StateClassPoolContract, OfficialPool: true, Status: ContractStatusActive}
	require.Equal(t, ContractStatusFrozenLimited, OfficialPoolStatusForDebt(pool, 1))
	require.False(t, FrozenLimitedPoolAllows(PoolActionDeposit))
	require.True(t, FrozenLimitedPoolAllows(PoolActionClaim))
	require.True(t, FrozenLimitedPoolAllows(PoolActionUnbond))
	require.True(t, FrozenLimitedPoolAllows(PoolActionMaturedWithdrawal))
	require.True(t, FrozenLimitedPoolAllows(PoolActionGovernanceRecovery))
}

func TestProtocolCriticalSystemStateIsProtocolPaidAndCannotFreezeFromRent(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteBlock = 1
	system := PersistentStateRecord{
		SubjectID:         "system/module",
		Class:             StateClassSystemModule,
		DataBytes:         10,
		Persistent:        true,
		Status:            ContractStatusActive,
		ProtocolCritical:  true,
		LastChargedHeight: 1,
	}

	result, err := AccruePersistentStateRent(system, params, 6)
	require.NoError(t, err)
	require.Equal(t, uint64(50), result.RentDelta)
	require.True(t, result.ProtocolPaid)
	require.False(t, result.UserFacingFee)
	require.Equal(t, ContractStatusActive, OfficialPoolStatusForDebt(result.Subject, result.Subject.RentDebt))
}

func TestSystemRentAccountingWarnsTopsUpDeterministicallyAndKeepsCriticalExecutable(t *testing.T) {
	result := ComputeSystemRentAccounting(SystemRentAccounting{
		AvailableFunds:                   10,
		ProjectedRentPerBlock:            5,
		WarningRunwayBlocks:              10,
		CriticalRunwayBlocks:             3,
		FeeCollectorBalance:              4,
		TreasuryBalance:                  3,
		GovernanceConfiguredPayerBalance: 2,
		RequiredTopUp:                    10,
		ProtocolCriticalExecutable:       true,
	})

	require.Equal(t, uint64(2), result.RunwayBlocks)
	require.Equal(t, SystemRentAlertInvariant, result.Alert)
	require.Equal(t, uint64(9), result.TopUpAmount)
	require.Equal(t, uint64(1), result.RemainingDebt)
	require.True(t, result.FreezeForbidden)
	require.True(t, result.Executable)
}

func TestFrozenContractPreservesReadAndProofButBlocksNormalExecute(t *testing.T) {
	frozen := PersistentStateRecord{SubjectID: "contract", Class: StateClassContract, Status: ContractStatusFrozen, RentDebt: 20}
	require.True(t, CanReadPersistentState(frozen))
	require.True(t, CanQueryPersistentStateProof(frozen))
	require.False(t, CanExecutePersistentStateAction(frozen, "execute"))

	limited := PersistentStateRecord{SubjectID: "official-pool", Class: StateClassPoolContract, Status: ContractStatusFrozenLimited, OfficialPool: true}
	require.False(t, CanExecutePersistentStateAction(limited, PoolActionDeposit))
	require.True(t, CanExecutePersistentStateAction(limited, PoolActionMaturedWithdrawal))
}

func TestPersistentStateClassValidationRejectsUnknownClasses(t *testing.T) {
	require.NoError(t, ValidatePersistentStateClass(StateClassPoolRewardIndex))
	require.ErrorContains(t, ValidatePersistentStateClass("native_token_module"), "unsupported")
}
