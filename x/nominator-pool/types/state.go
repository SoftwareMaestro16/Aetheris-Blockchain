package types

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

const (
	PoolStatusActive        = "active"
	PoolStatusPaused        = "paused"
	PoolStatusFrozenLimited = "frozen_limited"
	PoolStatusClosed        = "closed"

	WithdrawalStatusPending   = "pending"
	WithdrawalStatusCancelled = "cancelled"
	WithdrawalStatusCompleted = "completed"

	MaxPoolsV1                            = uint32(10_000)
	MaxDelegatorsV1                       = uint32(1_000_000)
	MaxPendingDepositsV1                  = uint32(1_000_000)
	MaxPendingWithdrawalsV1               = uint32(1_000_000)
	MaxUnbondingEntriesV1                 = uint32(1_000_000)
	MaxPoolIDBytesV1                      = uint32(96)
	MaxBasisPoints                        = uint32(10_000)
	IndexScale                            = uint64(1_000_000_000)
	DefaultMaxCommissionBps               = uint32(2_000)
	DefaultMaxValidatorCommissionBps      = uint32(2_000)
	DefaultMaxOperatorPerformanceBonusBps = uint32(1_000)
	DefaultUnbondingBlocks                = appparams.StakingUnbondingDefaultBlocks
	DefaultValidatorChangeDelay           = uint64(100)
	DefaultMinPoolDeposit                 = uint64(10)
)

type Params struct {
	Authority                      string
	MaxPools                       uint32
	MaxDelegators                  uint32
	MaxPendingDeposits             uint32
	MaxPendingWithdrawals          uint32
	MaxUnbondingEntries            uint32
	MaxPoolIDBytes                 uint32
	MaxCommissionBps               uint32
	MaxValidatorCommissionBps      uint32
	MaxOperatorPerformanceBonusBps uint32
	UnbondingBlocks                uint64
	ValidatorChangeDelay           uint64
	MinPoolDeposit                 uint64
	DirectUserDelegationEnabled    bool
}

type State struct {
	Pools []NominatorPool
}

type NominatorPool struct {
	PoolID                     string
	ContractAddressUser        string
	ContractAddressRaw         string
	OfficialLiquidStaking      bool
	PoolOperator               string
	ValidatorTarget            string
	PendingValidatorTarget     string
	ValidatorChangeHeight      uint64
	TotalShares                uint64
	TotalBondedStake           uint64
	Allocations                []PoolAllocation
	PendingDeposits            []PendingDeposit
	PendingWithdrawals         []PendingWithdrawal
	DelegatorShares            []DelegatorShare
	RewardIndex                uint64
	RewardRemainder            uint64
	SlashIndex                 uint64
	PoolCommissionBps          uint32
	RewardEpoch                uint64
	ProtocolFeeAccrued         uint64
	ValidatorCommissionAccrued uint64
	ValidatorOperatorIncome    []ValidatorIncome
	ValidatorAllocations       []ValidatorRewardAllocation
	Status                     string
	UnbondingQueue             []UnbondingEntry
}

type PendingDeposit struct {
	Delegator string
	Amount    uint64
	Height    uint64
}

type PendingWithdrawal struct {
	WithdrawalID   string
	Delegator      string
	Shares         uint64
	Amount         uint64
	RequestHeight  uint64
	CompleteHeight uint64
	Status         string
}

type DelegatorShare struct {
	Delegator             string
	Shares                uint64
	RewardIndexCheckpoint uint64
	PendingRewards        uint64
	SlashIndexCheckpoint  uint64
}

type UnbondingEntry struct {
	WithdrawalID   string
	Delegator      string
	Amount         uint64
	CompleteHeight uint64
	Status         string
}

type PoolAllocation struct {
	ValidatorAddress string
	Amount           uint64
	Height           uint64
}

type ValidatorRewardAllocation struct {
	Validator                   string
	PoolAllocatedStake          uint64
	ValidatorSelfStake          uint64
	PerformanceBps              uint32
	CommissionBps               uint32
	SlashingLoss                uint64
	Jailed                      bool
	InfrastructureCost          uint64
	OperatorPerformanceBonusBps uint32
	GrossPoolRewards            uint64
	ValidatorCommission         uint64
	PoolProtocolFee             uint64
	NetPoolRewards              uint64
	ValidatorSelfStakeRewards   uint64
	OperatorPerformanceBonus    uint64
	ValidatorGrossIncome        uint64
	ValidatorNetIncome          int64
	RewardIndexDelta            uint64
	RewardIndexAfter            uint64
}

type ValidatorIncome struct {
	Validator                string
	SelfStakeRewards         uint64
	CommissionIncome         uint64
	OperatorPerformanceBonus uint64
	InfrastructureCost       uint64
	GrossIncome              uint64
	NetIncome                int64
}

type PoolRewardSummary struct {
	PoolID                    string
	Epoch                     uint64
	RewardRateBps             uint32
	EmissionsAllocated        uint64
	FeesAllocated             uint64
	RewardCap                 uint64
	GrossPoolRewards          uint64
	ValidatorCommission       uint64
	PoolProtocolFee           uint64
	PoolUserRewards           uint64
	SlashingLosses            uint64
	ValidatorSelfStakeRewards uint64
	OperatorPerformanceBonus  uint64
	ValidatorGrossIncome      uint64
	ValidatorNetIncome        int64
	RewardIndexBefore         uint64
	RewardIndexAfter          uint64
	RewardRemainder           uint64
	AllocationsTouched        uint64
}

type MsgCreateNominatorPool struct {
	Authority         string
	PoolID            string
	PoolOperator      string
	ValidatorTarget   string
	PoolCommissionBps uint32
	Height            uint64
	ValidatorStatus   string
}

type MsgDepositToPool struct {
	Authority string
	PoolID    string
	Delegator string
	Amount    uint64
	Height    uint64
}

type MsgCreateOfficialLiquidStakingPool struct {
	Authority           string
	PoolID              string
	ContractAddressUser string
	ContractAddressRaw  string
	PoolOperator        string
	PoolCommissionBps   uint32
	Height              uint64
}

type MsgDepositToOfficialLiquidStaking struct {
	Authority        string
	PoolID           string
	UserAddress      string
	Amount           uint64
	Height           uint64
	ValidatorAddress string
}

type MsgDelegateToValidator struct {
	Authority        string
	UserAddress      string
	ValidatorAddress string
	Amount           uint64
	Height           uint64
}

type MsgInjectPooledStake struct {
	CallerContractUser string
	PoolID             string
	ValidatorAddress   string
	Amount             uint64
	Height             uint64
}

type MsgRequestPoolWithdrawal struct {
	Authority    string
	PoolID       string
	WithdrawalID string
	Delegator    string
	Shares       uint64
	Height       uint64
}

type MsgCancelPoolWithdrawal struct {
	Authority    string
	PoolID       string
	WithdrawalID string
	Delegator    string
	Height       uint64
}

type MsgClaimPoolRewards struct {
	Authority string
	PoolID    string
	Delegator string
	Height    uint64
}

type MsgSyncPoolRewards struct {
	Authority          string
	PoolID             string
	Epoch              uint64
	RewardRateBps      uint32
	EmissionsAllocated uint64
	FeesAllocated      uint64
	Height             uint64
	Allocations        []ValidatorRewardAllocation
}

type MsgClaimStakingRewards struct {
	Authority         string
	Delegator         string
	Validator         string
	Height            uint64
	InternalMigration bool
}

type MsgUpdatePoolCommission struct {
	Authority         string
	PoolID            string
	PoolOperator      string
	PoolCommissionBps uint32
	Height            uint64
}

type MsgChangePoolValidator struct {
	Authority       string
	PoolID          string
	PoolOperator    string
	ValidatorTarget string
	ValidatorStatus string
	Height          uint64
}

type QueryPoolShareRequest struct {
	PoolID    string
	Delegator string
}

type QueryPoolShareResponse struct {
	Share          DelegatorShare
	PendingRewards uint64
}

type QueryPoolAllocationsRequest struct {
	PoolID string
}

type QueryPoolAllocationsResponse struct {
	Allocations []ValidatorRewardAllocation
}

type QueryStakingRewardsRequest struct {
	Delegator         string
	Validator         string
	InternalMigration bool
}

type QueryStakingRewardsResponse struct {
	RewardAmount uint64
}

func DefaultParams() Params {
	return Params{
		Authority:                      prototype.DefaultAuthority,
		MaxPools:                       MaxPoolsV1,
		MaxDelegators:                  MaxDelegatorsV1,
		MaxPendingDeposits:             MaxPendingDepositsV1,
		MaxPendingWithdrawals:          MaxPendingWithdrawalsV1,
		MaxUnbondingEntries:            MaxUnbondingEntriesV1,
		MaxPoolIDBytes:                 MaxPoolIDBytesV1,
		MaxCommissionBps:               DefaultMaxCommissionBps,
		MaxValidatorCommissionBps:      DefaultMaxValidatorCommissionBps,
		MaxOperatorPerformanceBonusBps: DefaultMaxOperatorPerformanceBonusBps,
		UnbondingBlocks:                DefaultUnbondingBlocks,
		ValidatorChangeDelay:           DefaultValidatorChangeDelay,
		MinPoolDeposit:                 DefaultMinPoolDeposit,
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool authority", p.Authority); err != nil {
		return err
	}
	if p.MaxPools == 0 || p.MaxPools > MaxPoolsV1 {
		return fmt.Errorf("nominator pool max pools must be between 1 and %d", MaxPoolsV1)
	}
	if p.MaxDelegators == 0 || p.MaxDelegators > MaxDelegatorsV1 {
		return fmt.Errorf("nominator pool max delegators must be between 1 and %d", MaxDelegatorsV1)
	}
	if p.MaxPendingDeposits == 0 || p.MaxPendingDeposits > MaxPendingDepositsV1 {
		return fmt.Errorf("nominator pool max pending deposits must be between 1 and %d", MaxPendingDepositsV1)
	}
	if p.MaxPendingWithdrawals == 0 || p.MaxPendingWithdrawals > MaxPendingWithdrawalsV1 {
		return fmt.Errorf("nominator pool max pending withdrawals must be between 1 and %d", MaxPendingWithdrawalsV1)
	}
	if p.MaxUnbondingEntries == 0 || p.MaxUnbondingEntries > MaxUnbondingEntriesV1 {
		return fmt.Errorf("nominator pool max unbonding entries must be between 1 and %d", MaxUnbondingEntriesV1)
	}
	if p.MaxPoolIDBytes == 0 || p.MaxPoolIDBytes > MaxPoolIDBytesV1 {
		return fmt.Errorf("nominator pool max pool id bytes must be between 1 and %d", MaxPoolIDBytesV1)
	}
	if p.MaxCommissionBps > MaxBasisPoints {
		return fmt.Errorf("nominator pool max commission must be <= %d", MaxBasisPoints)
	}
	if p.MaxValidatorCommissionBps > MaxBasisPoints {
		return fmt.Errorf("nominator pool max validator commission must be <= %d", MaxBasisPoints)
	}
	if p.MaxOperatorPerformanceBonusBps > MaxBasisPoints {
		return fmt.Errorf("nominator pool max operator performance bonus must be <= %d", MaxBasisPoints)
	}
	if err := appparams.ValidateStakingUnbondingBlocks(p.UnbondingBlocks); err != nil {
		return fmt.Errorf("nominator pool %w", err)
	}
	if p.ValidatorChangeDelay == 0 {
		return errors.New("nominator pool validator change delay must be positive")
	}
	if p.MinPoolDeposit == 0 {
		return errors.New("nominator pool minimum pool deposit must be positive")
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("nominator pool update requires governance authority")
	}
	return nil
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Pools)) > params.MaxPools {
		return errors.New("nominator pool count limit exceeded")
	}
	ids := map[string]struct{}{}
	for _, pool := range s.Pools {
		if err := pool.Validate(params); err != nil {
			return err
		}
		if _, found := ids[pool.PoolID]; found {
			return fmt.Errorf("duplicate nominator pool id %s", pool.PoolID)
		}
		ids[pool.PoolID] = struct{}{}
	}
	return nil
}

func (p NominatorPool) Validate(params Params) error {
	if err := validateID("nominator pool id", p.PoolID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if p.OfficialLiquidStaking || strings.TrimSpace(p.ContractAddressUser) != "" || strings.TrimSpace(p.ContractAddressRaw) != "" {
		if err := ValidateUserFacingAEAddress("official liquid staking contract address", p.ContractAddressUser); err != nil {
			return err
		}
		if err := ValidateRawAddress("official liquid staking contract raw address", p.ContractAddressRaw); err != nil {
			return err
		}
		if err := ValidateAddressPair("official liquid staking contract address pair", p.ContractAddressUser, p.ContractAddressRaw); err != nil {
			return err
		}
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool operator", p.PoolOperator); err != nil {
		return err
	}
	if strings.TrimSpace(p.ValidatorTarget) != "" {
		if err := addressing.ValidateAuthorityAddress("nominator pool validator target", p.ValidatorTarget); err != nil {
			return err
		}
	}
	if strings.TrimSpace(p.PendingValidatorTarget) != "" {
		if err := addressing.ValidateAuthorityAddress("nominator pool pending validator target", p.PendingValidatorTarget); err != nil {
			return err
		}
		if p.ValidatorChangeHeight == 0 {
			return errors.New("nominator pool pending validator change requires activation height")
		}
	}
	if p.PoolCommissionBps > params.MaxCommissionBps {
		return errors.New("nominator pool commission exceeds configured bound")
	}
	if !isPoolStatus(p.Status) {
		return fmt.Errorf("unsupported nominator pool status %q", p.Status)
	}
	if uint32(len(p.DelegatorShares)) > params.MaxDelegators {
		return errors.New("nominator pool delegator limit exceeded")
	}
	if uint32(len(p.PendingDeposits)) > params.MaxPendingDeposits {
		return errors.New("nominator pool pending deposit limit exceeded")
	}
	if uint32(len(p.PendingWithdrawals)) > params.MaxPendingWithdrawals {
		return errors.New("nominator pool pending withdrawal limit exceeded")
	}
	if uint32(len(p.UnbondingQueue)) > params.MaxUnbondingEntries {
		return errors.New("nominator pool unbonding queue limit exceeded")
	}
	if p.TotalShares != sumShares(p.DelegatorShares) {
		return errors.New("nominator pool total shares do not match delegator shares")
	}
	if err := ValidateAllocations(p.Allocations, p.TotalBondedStake); err != nil {
		return err
	}
	delegators := map[string]struct{}{}
	for _, delegator := range p.DelegatorShares {
		if err := delegator.Validate(); err != nil {
			return err
		}
		if _, found := delegators[delegator.Delegator]; found {
			return fmt.Errorf("duplicate pool delegator %s", delegator.Delegator)
		}
		delegators[delegator.Delegator] = struct{}{}
	}
	withdrawals := map[string]struct{}{}
	for _, withdrawal := range p.PendingWithdrawals {
		if err := withdrawal.Validate(); err != nil {
			return err
		}
		if _, found := withdrawals[withdrawal.WithdrawalID]; found {
			return fmt.Errorf("duplicate pool withdrawal %s", withdrawal.WithdrawalID)
		}
		withdrawals[withdrawal.WithdrawalID] = struct{}{}
		if withdrawal.Status == WithdrawalStatusPending {
			if _, found := delegators[withdrawal.Delegator]; !found {
				return fmt.Errorf("withdrawal %s references unknown delegator", withdrawal.WithdrawalID)
			}
		}
	}
	for _, deposit := range p.PendingDeposits {
		if err := deposit.Validate(); err != nil {
			return err
		}
	}
	for _, entry := range p.UnbondingQueue {
		if err := entry.Validate(); err != nil {
			return err
		}
	}
	incomeByValidator := map[string]struct{}{}
	for _, income := range p.ValidatorOperatorIncome {
		if err := income.Validate(); err != nil {
			return err
		}
		if _, found := incomeByValidator[income.Validator]; found {
			return fmt.Errorf("duplicate validator income %s", income.Validator)
		}
		incomeByValidator[income.Validator] = struct{}{}
	}
	for _, allocation := range p.ValidatorAllocations {
		if err := allocation.Validate(params); err != nil {
			return err
		}
	}
	return nil
}

func (a PoolAllocation) Validate() error {
	if err := ValidateUserFacingAEAddress("pool allocation validator address", a.ValidatorAddress); err != nil {
		return err
	}
	if a.Amount == 0 || a.Height == 0 {
		return errors.New("pool allocation amount and height must be positive")
	}
	return nil
}

func ValidateAllocations(allocations []PoolAllocation, totalBondedStake uint64) error {
	previous := ""
	total := uint64(0)
	for _, allocation := range allocations {
		if err := allocation.Validate(); err != nil {
			return err
		}
		if allocation.ValidatorAddress <= previous {
			return errors.New("pool allocations must be sorted by unique validator address")
		}
		previous = allocation.ValidatorAddress
		if allocation.Amount > totalBondedStake-total {
			return errors.New("pool allocations exceed bonded stake")
		}
		total += allocation.Amount
	}
	return nil
}

func (a ValidatorRewardAllocation) Validate(params Params) error {
	if err := addressing.ValidateAuthorityAddress("nominator pool reward validator", a.Validator); err != nil {
		return err
	}
	if a.PerformanceBps > MaxBasisPoints {
		return errors.New("nominator pool validator performance exceeds basis points")
	}
	if a.CommissionBps > params.MaxValidatorCommissionBps {
		return errors.New("nominator pool validator commission exceeds configured bound")
	}
	if a.OperatorPerformanceBonusBps > params.MaxOperatorPerformanceBonusBps {
		return errors.New("nominator pool operator performance bonus exceeds configured bound")
	}
	if a.Jailed && a.OperatorPerformanceBonus > 0 {
		return errors.New("jailed validator cannot receive positive operator bonus")
	}
	return nil
}

func (i ValidatorIncome) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool validator income", i.Validator); err != nil {
		return err
	}
	expectedGross, err := CheckedAddUint64(i.SelfStakeRewards, i.CommissionIncome)
	if err != nil {
		return err
	}
	expectedGross, err = CheckedAddUint64(expectedGross, i.OperatorPerformanceBonus)
	if err != nil {
		return err
	}
	if i.GrossIncome != expectedGross {
		return errors.New("nominator pool validator gross income does not reconcile")
	}
	if i.NetIncome != SaturatingNetIncome(i.GrossIncome, i.InfrastructureCost) {
		return errors.New("nominator pool validator net income does not reconcile")
	}
	return nil
}

func (d PendingDeposit) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool pending deposit delegator", d.Delegator); err != nil {
		return err
	}
	if d.Amount == 0 || d.Height == 0 {
		return errors.New("nominator pool pending deposit amount and height must be positive")
	}
	return nil
}

func (w PendingWithdrawal) Validate() error {
	if err := validateID("nominator pool withdrawal id", w.WithdrawalID, MaxPoolIDBytesV1); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool withdrawal delegator", w.Delegator); err != nil {
		return err
	}
	if w.Shares == 0 || w.Amount == 0 || w.RequestHeight == 0 || w.CompleteHeight <= w.RequestHeight {
		return errors.New("nominator pool withdrawal amounts and heights are invalid")
	}
	if !isWithdrawalStatus(w.Status) {
		return fmt.Errorf("unsupported nominator pool withdrawal status %q", w.Status)
	}
	return nil
}

func (d DelegatorShare) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool delegator", d.Delegator); err != nil {
		return err
	}
	if d.Shares == 0 {
		return errors.New("nominator pool delegator shares must be positive")
	}
	return nil
}

func (e UnbondingEntry) Validate() error {
	if err := validateID("nominator pool unbonding withdrawal id", e.WithdrawalID, MaxPoolIDBytesV1); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool unbonding delegator", e.Delegator); err != nil {
		return err
	}
	if e.Amount == 0 || e.CompleteHeight == 0 {
		return errors.New("nominator pool unbonding amount and completion height must be positive")
	}
	if !isWithdrawalStatus(e.Status) {
		return fmt.Errorf("unsupported nominator pool unbonding status %q", e.Status)
	}
	return nil
}

func (s State) Normalize(params Params) State {
	s.Pools = SortPools(s.Pools)
	for idx := range s.Pools {
		s.Pools[idx].Allocations = SortAllocations(s.Pools[idx].Allocations)
		s.Pools[idx].PendingDeposits = SortDeposits(s.Pools[idx].PendingDeposits)
		s.Pools[idx].PendingWithdrawals = SortWithdrawals(s.Pools[idx].PendingWithdrawals)
		s.Pools[idx].DelegatorShares = SortDelegators(s.Pools[idx].DelegatorShares)
		s.Pools[idx].UnbondingQueue = SortUnbonding(s.Pools[idx].UnbondingQueue)
		s.Pools[idx].ValidatorOperatorIncome = SortValidatorIncome(s.Pools[idx].ValidatorOperatorIncome)
		s.Pools[idx].ValidatorAllocations = SortValidatorRewardAllocations(s.Pools[idx].ValidatorAllocations)
	}
	return s
}

func SortAllocations(values []PoolAllocation) []PoolAllocation {
	out := append([]PoolAllocation(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ValidatorAddress < out[j].ValidatorAddress })
	return out
}

func SortValidatorIncome(values []ValidatorIncome) []ValidatorIncome {
	out := append([]ValidatorIncome(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Validator < out[j].Validator })
	return out
}

func SortValidatorRewardAllocations(values []ValidatorRewardAllocation) []ValidatorRewardAllocation {
	out := append([]ValidatorRewardAllocation(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Validator < out[j].Validator })
	return out
}

func SortPools(values []NominatorPool) []NominatorPool {
	out := append([]NominatorPool(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].PoolID < out[j].PoolID })
	return out
}

func SortDelegators(values []DelegatorShare) []DelegatorShare {
	out := append([]DelegatorShare(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Delegator < out[j].Delegator })
	return out
}

func SortWithdrawals(values []PendingWithdrawal) []PendingWithdrawal {
	out := append([]PendingWithdrawal(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].WithdrawalID < out[j].WithdrawalID })
	return out
}

func SortDeposits(values []PendingDeposit) []PendingDeposit {
	out := append([]PendingDeposit(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].Delegator < out[j].Delegator
	})
	return out
}

func SortUnbonding(values []UnbondingEntry) []UnbondingEntry {
	out := append([]UnbondingEntry(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CompleteHeight != out[j].CompleteHeight {
			return out[i].CompleteHeight < out[j].CompleteHeight
		}
		return out[i].WithdrawalID < out[j].WithdrawalID
	})
	return out
}

func ShareValue(pool NominatorPool, shares uint64) uint64 {
	if pool.TotalShares == 0 {
		return 0
	}
	return shares * pool.TotalBondedStake / pool.TotalShares
}

func SharesForDeposit(pool NominatorPool, amount uint64) uint64 {
	if pool.TotalShares == 0 || pool.TotalBondedStake == 0 {
		return amount
	}
	shares := amount * pool.TotalShares / pool.TotalBondedStake
	if shares == 0 && amount > 0 {
		return 1
	}
	return shares
}

func ValidateOfficialLiquidStakingDeposit(msg MsgDepositToOfficialLiquidStaking, params Params) error {
	if err := params.Authorize(msg.Authority); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("official liquid staking depositor", msg.UserAddress); err != nil {
		return err
	}
	if strings.TrimSpace(msg.ValidatorAddress) != "" {
		return errors.New("official liquid staking deposit must not include a validator address")
	}
	if msg.Amount < params.MinPoolDeposit {
		return fmt.Errorf("official liquid staking deposit below configured minimum %d", params.MinPoolDeposit)
	}
	if msg.Height == 0 {
		return errors.New("official liquid staking deposit height must be positive")
	}
	return validateID("official liquid staking pool id", msg.PoolID, params.MaxPoolIDBytes)
}

func ValidateDirectUserDelegation(msg MsgDelegateToValidator, params Params) error {
	if err := params.Authorize(msg.Authority); err != nil {
		return err
	}
	if !params.DirectUserDelegationEnabled {
		return errors.New("direct user delegation to validators is disabled; use official liquid staking pool deposit")
	}
	if err := ValidateUserFacingAEAddress("direct delegation user address", msg.UserAddress); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("direct delegation validator address", msg.ValidatorAddress); err != nil {
		return err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return errors.New("direct delegation amount and height must be positive")
	}
	return nil
}

func ValidateUserFacingAEAddress(field, text string) error {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, addressing.UserFriendlyPrefix) {
		return fmt.Errorf("%s must use AE user-facing address format", field)
	}
	return addressing.ValidateUserAddress(field, text)
}

func ValidateRawAddress(field, text string) error {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, addressing.RawPrefix) {
		return fmt.Errorf("%s must use 4: raw address format", field)
	}
	_, err := addressing.Parse(text)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", field, err)
	}
	return nil
}

func ValidateAddressPair(field, userAddress, rawAddress string) error {
	userBytes, err := addressing.Parse(userAddress)
	if err != nil {
		return fmt.Errorf("invalid %s user address: %w", field, err)
	}
	rawBytes, err := addressing.Parse(rawAddress)
	if err != nil {
		return fmt.Errorf("invalid %s raw address: %w", field, err)
	}
	userKey, err := addressing.AddressTextBytesKey(userAddress)
	if err != nil {
		return err
	}
	rawKey, err := addressing.AddressTextBytesKey(rawAddress)
	if err != nil {
		return err
	}
	if userKey != rawKey || string(userBytes) != string(rawBytes) {
		return fmt.Errorf("%s AE and raw addresses must represent the same account", field)
	}
	return nil
}

func RawAddressForUserAddress(userAddress string) (string, error) {
	if err := ValidateUserFacingAEAddress("user address", userAddress); err != nil {
		return "", err
	}
	bz, err := addressing.Parse(userAddress)
	if err != nil {
		return "", err
	}
	return addressing.Format(bz), nil
}

func RewardDelta(amount uint64, totalShares uint64) uint64 {
	if amount == 0 || totalShares == 0 {
		return 0
	}
	return amount * IndexScale / totalShares
}

func IndexedRewardAmount(delta uint64, totalShares uint64) uint64 {
	if delta == 0 || totalShares == 0 {
		return 0
	}
	return delta * totalShares / IndexScale
}

func AccruedReward(delegator DelegatorShare, rewardIndex uint64) uint64 {
	if rewardIndex <= delegator.RewardIndexCheckpoint {
		return delegator.PendingRewards
	}
	return delegator.PendingRewards + delegator.Shares*(rewardIndex-delegator.RewardIndexCheckpoint)/IndexScale
}

func SyncPoolRewards(params Params, pool NominatorPool, msg MsgSyncPoolRewards) (NominatorPool, PoolRewardSummary, error) {
	if err := params.Authorize(msg.Authority); err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	if msg.PoolID != pool.PoolID {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward sync pool mismatch")
	}
	if msg.Epoch == 0 || msg.Height == 0 {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward sync epoch and height must be positive")
	}
	if msg.Epoch <= pool.RewardEpoch {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward sync epoch must increase")
	}
	if msg.RewardRateBps > MaxBasisPoints {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward rate exceeds basis points")
	}
	if len(msg.Allocations) == 0 {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward sync requires validator allocations")
	}
	rewardCap, err := CheckedAddUint64(msg.EmissionsAllocated, msg.FeesAllocated)
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}

	next := pool
	next.RewardEpoch = msg.Epoch
	next.ValidatorAllocations = nil
	summary := PoolRewardSummary{
		PoolID:             pool.PoolID,
		Epoch:              msg.Epoch,
		RewardRateBps:      msg.RewardRateBps,
		EmissionsAllocated: msg.EmissionsAllocated,
		FeesAllocated:      msg.FeesAllocated,
		RewardCap:          rewardCap,
		RewardIndexBefore:  pool.RewardIndex,
	}
	income := map[string]ValidatorIncome{}
	totalRewardOut := uint64(0)
	index := pool.RewardIndex
	remainder := pool.RewardRemainder
	for _, allocation := range SortValidatorRewardAllocations(msg.Allocations) {
		if err := allocation.Validate(params); err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		effectivePerformanceBps := allocation.PerformanceBps
		if allocation.Jailed {
			effectivePerformanceBps = 0
			allocation.OperatorPerformanceBonusBps = 0
		}
		grossPoolRewards, err := RewardForStake(allocation.PoolAllocatedStake, msg.RewardRateBps, effectivePerformanceBps)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		commission, err := MulDivUint64(grossPoolRewards, uint64(allocation.CommissionBps), uint64(MaxBasisPoints))
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		afterCommission := grossPoolRewards - commission
		poolFee, err := MulDivUint64(afterCommission, uint64(pool.PoolCommissionBps), uint64(MaxBasisPoints))
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		netPoolRewards := afterCommission - poolFee
		selfStakeRewards, err := RewardForStake(allocation.ValidatorSelfStake, msg.RewardRateBps, effectivePerformanceBps)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		operatorBonus, err := MulDivUint64(grossPoolRewards, uint64(allocation.OperatorPerformanceBonusBps), uint64(MaxBasisPoints))
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		validatorGross, err := CheckedAddUint64(selfStakeRewards, commission)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		validatorGross, err = CheckedAddUint64(validatorGross, operatorBonus)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		totalRewardOut, err = CheckedAddUint64(totalRewardOut, grossPoolRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		totalRewardOut, err = CheckedAddUint64(totalRewardOut, selfStakeRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		totalRewardOut, err = CheckedAddUint64(totalRewardOut, operatorBonus)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		if totalRewardOut > rewardCap {
			return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool rewards exceed emissions and fee allocation cap")
		}

		delta := RewardDelta(netPoolRewards, pool.TotalShares)
		index, err = CheckedAddUint64(index, delta)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		distributed := IndexedRewardAmount(delta, pool.TotalShares)
		if netPoolRewards >= distributed {
			remainder, err = CheckedAddUint64(remainder, netPoolRewards-distributed)
			if err != nil {
				return NominatorPool{}, PoolRewardSummary{}, err
			}
		}
		allocation.PerformanceBps = effectivePerformanceBps
		allocation.GrossPoolRewards = grossPoolRewards
		allocation.ValidatorCommission = commission
		allocation.PoolProtocolFee = poolFee
		allocation.NetPoolRewards = netPoolRewards
		allocation.ValidatorSelfStakeRewards = selfStakeRewards
		allocation.OperatorPerformanceBonus = operatorBonus
		allocation.ValidatorGrossIncome = validatorGross
		allocation.ValidatorNetIncome = SaturatingNetIncome(validatorGross, allocation.InfrastructureCost)
		allocation.RewardIndexDelta = delta
		allocation.RewardIndexAfter = index
		next.ValidatorAllocations = append(next.ValidatorAllocations, allocation)

		entry := income[allocation.Validator]
		entry.Validator = allocation.Validator
		entry.SelfStakeRewards, err = CheckedAddUint64(entry.SelfStakeRewards, selfStakeRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.CommissionIncome, err = CheckedAddUint64(entry.CommissionIncome, commission)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.OperatorPerformanceBonus, err = CheckedAddUint64(entry.OperatorPerformanceBonus, operatorBonus)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.InfrastructureCost, err = CheckedAddUint64(entry.InfrastructureCost, allocation.InfrastructureCost)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.GrossIncome, err = CheckedAddUint64(entry.GrossIncome, validatorGross)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.NetIncome = SaturatingNetIncome(entry.GrossIncome, entry.InfrastructureCost)
		income[allocation.Validator] = entry

		summary.GrossPoolRewards, err = CheckedAddUint64(summary.GrossPoolRewards, grossPoolRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.ValidatorCommission, err = CheckedAddUint64(summary.ValidatorCommission, commission)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.PoolProtocolFee, err = CheckedAddUint64(summary.PoolProtocolFee, poolFee)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.PoolUserRewards, err = CheckedAddUint64(summary.PoolUserRewards, netPoolRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.SlashingLosses, err = CheckedAddUint64(summary.SlashingLosses, allocation.SlashingLoss)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.ValidatorSelfStakeRewards, err = CheckedAddUint64(summary.ValidatorSelfStakeRewards, selfStakeRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.OperatorPerformanceBonus, err = CheckedAddUint64(summary.OperatorPerformanceBonus, operatorBonus)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.ValidatorGrossIncome, err = CheckedAddUint64(summary.ValidatorGrossIncome, validatorGross)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.AllocationsTouched++
	}
	next.ValidatorOperatorIncome = next.ValidatorOperatorIncome[:0]
	for _, value := range income {
		if err := value.Validate(); err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		next.ValidatorOperatorIncome = append(next.ValidatorOperatorIncome, value)
	}
	next.ValidatorOperatorIncome = SortValidatorIncome(next.ValidatorOperatorIncome)
	next.RewardIndex = index
	next.RewardRemainder = remainder
	next.ProtocolFeeAccrued, err = CheckedAddUint64(next.ProtocolFeeAccrued, summary.PoolProtocolFee)
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	next.ValidatorCommissionAccrued, err = CheckedAddUint64(next.ValidatorCommissionAccrued, summary.ValidatorCommission)
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	next.TotalBondedStake, err = CheckedAddUint64(next.TotalBondedStake, summary.PoolUserRewards)
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	if summary.SlashingLosses > next.TotalBondedStake {
		summary.SlashingLosses = next.TotalBondedStake
	}
	next.TotalBondedStake -= summary.SlashingLosses
	next.SlashIndex, err = CheckedAddUint64(next.SlashIndex, RewardDelta(summary.SlashingLosses, pool.TotalShares))
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	summary.RewardIndexAfter = next.RewardIndex
	summary.RewardRemainder = next.RewardRemainder
	summary.ValidatorNetIncome = SaturatingNetIncome(summary.ValidatorGrossIncome, totalInfrastructureCost(next.ValidatorOperatorIncome))
	return next, summary, nil
}

func RewardForStake(stake uint64, rewardRateBps uint32, performanceBps uint32) (uint64, error) {
	if rewardRateBps > MaxBasisPoints || performanceBps > MaxBasisPoints {
		return 0, errors.New("nominator pool reward inputs exceed basis points")
	}
	first, err := MulDivUint64(stake, uint64(rewardRateBps), uint64(MaxBasisPoints))
	if err != nil {
		return 0, err
	}
	return MulDivUint64(first, uint64(performanceBps), uint64(MaxBasisPoints))
}

func MulDivUint64(value, multiplier, denominator uint64) (uint64, error) {
	if denominator == 0 {
		return 0, errors.New("nominator pool division by zero")
	}
	product := new(big.Int).Mul(new(big.Int).SetUint64(value), new(big.Int).SetUint64(multiplier))
	product.Quo(product, new(big.Int).SetUint64(denominator))
	if !product.IsUint64() {
		return 0, errors.New("nominator pool uint64 accounting overflow")
	}
	return product.Uint64(), nil
}

func CheckedAddUint64(left, right uint64) (uint64, error) {
	if math.MaxUint64-left < right {
		return 0, errors.New("nominator pool uint64 accounting overflow")
	}
	return left + right, nil
}

func SaturatingNetIncome(gross uint64, cost uint64) int64 {
	if cost >= gross {
		delta := cost - gross
		if delta > uint64(math.MaxInt64) {
			return math.MinInt64
		}
		return -int64(delta)
	}
	delta := gross - cost
	if delta > uint64(math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(delta)
}

func totalInfrastructureCost(values []ValidatorIncome) uint64 {
	total := uint64(0)
	for _, value := range values {
		next, err := CheckedAddUint64(total, value.InfrastructureCost)
		if err != nil {
			return math.MaxUint64
		}
		total = next
	}
	return total
}

func IsJailedValidatorStatus(status string) bool {
	return status == validatorregistrytypes.StatusJailed || status == validatorregistrytypes.StatusTombstoned
}

func validateID(field, value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if uint32(len(value)) > maxBytes || strings.ContainsAny(value, " \t\r\n") {
		return fmt.Errorf("%s must be non-blank, whitespace-free, and within configured length", field)
	}
	return nil
}

func isPoolStatus(status string) bool {
	return status == PoolStatusActive || status == PoolStatusPaused || status == PoolStatusFrozenLimited || status == PoolStatusClosed
}

func isWithdrawalStatus(status string) bool {
	return status == WithdrawalStatusPending || status == WithdrawalStatusCancelled || status == WithdrawalStatusCompleted
}

func sumShares(values []DelegatorShare) uint64 {
	total := uint64(0)
	for _, value := range values {
		total += value.Shares
	}
	return total
}
