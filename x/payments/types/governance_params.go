package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultMinimumChannelCollateral           = "1"
	DefaultMaximumChannelCollateral           = "1000000000000000000"
	DefaultMaximumPromiseAmountRatioBps       = uint32(10_000)
	DefaultMaximumPromiseLifetime             = uint64(10_000)
	DefaultExpiredPromiseCleanupLimitPerBlock = uint64(256)
)

type PaymentGovernanceParams struct {
	Channel     PaymentChannelGovernanceParams
	Conditional PaymentConditionalGovernanceParams
	ParamsHash  string
}

type PaymentChannelGovernanceParams struct {
	MinimumChannelCollateral  string
	MaximumChannelCollateral  string
	MinimumChallengePeriod    uint64
	MaximumChallengePeriod    uint64
	DefaultChallengePeriod    uint64
	MinimumCloseDelay         uint64
	MaximumCloseDelay         uint64
	ChannelOpenBaseFee        string
	ChannelStorageFeePerByte  string
	ChannelTombstoneRetention uint64
}

type PaymentConditionalGovernanceParams struct {
	MaximumActivePromisesPerChannel    uint64
	MaximumPromiseAmountRatioBps       uint32
	MinimumTimeoutMargin               uint64
	MaximumPromiseLifetime             uint64
	BatchResolutionMaximumSize         uint64
	PromiseStorageFee                  string
	ExpiredPromiseCleanupLimitPerBlock uint64
}

func DefaultPaymentGovernanceParams() PaymentGovernanceParams {
	feeSchedule := DefaultPaymentFeeSchedule().Normalize()
	params := PaymentGovernanceParams{
		Channel: PaymentChannelGovernanceParams{
			MinimumChannelCollateral:  DefaultMinimumChannelCollateral,
			MaximumChannelCollateral:  DefaultMaximumChannelCollateral,
			MinimumChallengePeriod:    MinChallengePeriod,
			MaximumChallengePeriod:    MaxChallengePeriod,
			DefaultChallengePeriod:    DefaultDisputePeriod,
			MinimumCloseDelay:         MinCloseDelay,
			MaximumCloseDelay:         MaxCloseDelay,
			ChannelOpenBaseFee:        feeSchedule.ChannelOpenFee,
			ChannelStorageFeePerByte:  feeSchedule.StorageByteFee,
			ChannelTombstoneRetention: DefaultReplayHorizon,
		},
		Conditional: PaymentConditionalGovernanceParams{
			MaximumActivePromisesPerChannel:    MaxConditionsPerState,
			MaximumPromiseAmountRatioBps:       DefaultMaximumPromiseAmountRatioBps,
			MinimumTimeoutMargin:               DefaultTimeoutMargin,
			MaximumPromiseLifetime:             DefaultMaximumPromiseLifetime,
			BatchResolutionMaximumSize:         MaxSettlementBatchOps,
			PromiseStorageFee:                  feeSchedule.ConditionalPromiseSettlementFee,
			ExpiredPromiseCleanupLimitPerBlock: DefaultExpiredPromiseCleanupLimitPerBlock,
		},
	}
	params = params.Normalize()
	params.ParamsHash = ComputePaymentGovernanceParamsHash(params)
	return params
}

func ComputePaymentGovernanceParamsHash(params PaymentGovernanceParams) string {
	params = params.Normalize()
	return HashParts(
		"payments-governance-params-v1",
		params.Channel.MinimumChannelCollateral,
		params.Channel.MaximumChannelCollateral,
		fmt.Sprintf("%020d", params.Channel.MinimumChallengePeriod),
		fmt.Sprintf("%020d", params.Channel.MaximumChallengePeriod),
		fmt.Sprintf("%020d", params.Channel.DefaultChallengePeriod),
		fmt.Sprintf("%020d", params.Channel.MinimumCloseDelay),
		fmt.Sprintf("%020d", params.Channel.MaximumCloseDelay),
		params.Channel.ChannelOpenBaseFee,
		params.Channel.ChannelStorageFeePerByte,
		fmt.Sprintf("%020d", params.Channel.ChannelTombstoneRetention),
		fmt.Sprintf("%020d", params.Conditional.MaximumActivePromisesPerChannel),
		fmt.Sprintf("%010d", params.Conditional.MaximumPromiseAmountRatioBps),
		fmt.Sprintf("%020d", params.Conditional.MinimumTimeoutMargin),
		fmt.Sprintf("%020d", params.Conditional.MaximumPromiseLifetime),
		fmt.Sprintf("%020d", params.Conditional.BatchResolutionMaximumSize),
		params.Conditional.PromiseStorageFee,
		fmt.Sprintf("%020d", params.Conditional.ExpiredPromiseCleanupLimitPerBlock),
	)
}

func BuildGovernedPaymentFeeSchedule(params PaymentGovernanceParams) (PaymentFeeSchedule, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return PaymentFeeSchedule{}, err
	}
	schedule := DefaultPaymentFeeSchedule().Normalize()
	schedule.ChannelOpenFee = params.Channel.ChannelOpenBaseFee
	schedule.OpenFeeMin = params.Channel.ChannelOpenBaseFee
	schedule.StorageByteFee = params.Channel.ChannelStorageFeePerByte
	schedule.ConditionalPromiseSettlementFee = params.Conditional.PromiseStorageFee
	return schedule.Normalize(), schedule.Validate()
}

func ValidateChannelOpenRequestWithGovernance(req ChannelOpenRequest, params PaymentGovernanceParams) error {
	req = req.Normalize()
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}
	if err := params.Channel.ValidateOpenCollateral(req.Collateral); err != nil {
		return err
	}
	if req.ChallengePeriod < params.Channel.MinimumChallengePeriod || req.ChallengePeriod > params.Channel.MaximumChallengePeriod {
		return fmt.Errorf("payments open challenge period must be between governance bounds %d and %d", params.Channel.MinimumChallengePeriod, params.Channel.MaximumChallengePeriod)
	}
	if req.CloseDelay < params.Channel.MinimumCloseDelay || req.CloseDelay > params.Channel.MaximumCloseDelay {
		return fmt.Errorf("payments open close delay must be between governance bounds %d and %d", params.Channel.MinimumCloseDelay, params.Channel.MaximumCloseDelay)
	}
	paid, err := parseNonNegativeInt("payments opening fee paid", req.OpeningFeePaid)
	if err != nil {
		return err
	}
	required, err := parseNonNegativeInt("payments governance channel open base fee", params.Channel.ChannelOpenBaseFee)
	if err != nil {
		return err
	}
	if paid.LT(required) {
		return errors.New("payments opening fee paid is below governance base fee")
	}
	return nil
}

func ValidateConditionalPromisesForChannelWithGovernance(channel ChannelRecord, promises []ConditionalPromise, settledClaims []ConditionClaimRecord, params PaymentGovernanceParams) error {
	channel = channel.Normalize()
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if uint64(len(promises)) > params.Conditional.MaximumActivePromisesPerChannel {
		return fmt.Errorf("payments active promises exceed governance maximum %d", params.Conditional.MaximumActivePromisesPerChannel)
	}
	if err := ValidateConditionalPromisesForChannel(channel, promises, settledClaims); err != nil {
		return err
	}
	collateral, err := parsePositiveInt("payments governance channel collateral", channel.Collateral)
	if err != nil {
		return err
	}
	maxPromiseAmount := collateral.MulRaw(int64(params.Conditional.MaximumPromiseAmountRatioBps)).QuoRaw(10_000)
	for _, promise := range normalizeConditionalPromises(promises) {
		if err := params.Conditional.ValidatePromiseWindow(channel, promise); err != nil {
			return err
		}
		amount, err := parsePositiveInt("payments governance promise amount", promise.Amount)
		if err != nil {
			return err
		}
		fee, err := parseNonNegativeInt("payments governance promise fee", promise.Fee)
		if err != nil {
			return err
		}
		if amount.Add(fee).GT(maxPromiseAmount) {
			return errors.New("payments promise amount exceeds governance channel ratio")
		}
	}
	return nil
}

func ValidateBatchConditionSettlementWithGovernance(req BatchConditionSettlementRequest, state PaymentsState, settledClaims []ConditionClaimRecord, params PaymentGovernanceParams) error {
	req = req.Normalize()
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if uint64(len(req.LinkageProof.Promises)) > params.Conditional.BatchResolutionMaximumSize {
		return fmt.Errorf("payments batch promise resolution exceeds governance maximum %d", params.Conditional.BatchResolutionMaximumSize)
	}
	if err := req.ValidateForState(state, settledClaims); err != nil {
		return err
	}
	return nil
}

func SettlementTombstoneExpiryHeight(closedHeight uint64, params PaymentGovernanceParams) (uint64, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return 0, err
	}
	if closedHeight == 0 {
		return 0, errors.New("payments tombstone closed height must be positive")
	}
	if closedHeight+params.Channel.ChannelTombstoneRetention < closedHeight {
		return 0, errors.New("payments tombstone retention overflows height")
	}
	return closedHeight + params.Channel.ChannelTombstoneRetention, nil
}

func ExpiredPromiseCleanupLimit(params PaymentGovernanceParams) (uint64, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return 0, err
	}
	return params.Conditional.ExpiredPromiseCleanupLimitPerBlock, nil
}

func (p PaymentGovernanceParams) Normalize() PaymentGovernanceParams {
	defaults := DefaultPaymentGovernanceParamsNoHash()
	p.Channel = p.Channel.Normalize(defaults.Channel)
	p.Conditional = p.Conditional.Normalize(defaults.Conditional)
	p.ParamsHash = normalizeOptionalHash(p.ParamsHash)
	return p
}

func (p PaymentGovernanceParams) WithHash() PaymentGovernanceParams {
	p = p.Normalize()
	p.ParamsHash = ComputePaymentGovernanceParamsHash(p)
	return p.Normalize()
}

func (p PaymentGovernanceParams) Validate() error {
	params := p.Normalize()
	if err := params.Channel.Validate(); err != nil {
		return err
	}
	if err := params.Conditional.Validate(); err != nil {
		return err
	}
	if err := ValidateHash("payments governance params hash", params.ParamsHash); err != nil {
		return err
	}
	if expected := ComputePaymentGovernanceParamsHash(params); params.ParamsHash != expected {
		return errors.New("payments governance params hash mismatch")
	}
	return nil
}

func DefaultPaymentGovernanceParamsNoHash() PaymentGovernanceParams {
	feeSchedule := DefaultPaymentFeeSchedule().Normalize()
	return PaymentGovernanceParams{
		Channel: PaymentChannelGovernanceParams{
			MinimumChannelCollateral:  DefaultMinimumChannelCollateral,
			MaximumChannelCollateral:  DefaultMaximumChannelCollateral,
			MinimumChallengePeriod:    MinChallengePeriod,
			MaximumChallengePeriod:    MaxChallengePeriod,
			DefaultChallengePeriod:    DefaultDisputePeriod,
			MinimumCloseDelay:         MinCloseDelay,
			MaximumCloseDelay:         MaxCloseDelay,
			ChannelOpenBaseFee:        feeSchedule.ChannelOpenFee,
			ChannelStorageFeePerByte:  feeSchedule.StorageByteFee,
			ChannelTombstoneRetention: DefaultReplayHorizon,
		},
		Conditional: PaymentConditionalGovernanceParams{
			MaximumActivePromisesPerChannel:    MaxConditionsPerState,
			MaximumPromiseAmountRatioBps:       DefaultMaximumPromiseAmountRatioBps,
			MinimumTimeoutMargin:               DefaultTimeoutMargin,
			MaximumPromiseLifetime:             DefaultMaximumPromiseLifetime,
			BatchResolutionMaximumSize:         MaxSettlementBatchOps,
			PromiseStorageFee:                  feeSchedule.ConditionalPromiseSettlementFee,
			ExpiredPromiseCleanupLimitPerBlock: DefaultExpiredPromiseCleanupLimitPerBlock,
		},
	}
}

func (p PaymentChannelGovernanceParams) Normalize(defaults PaymentChannelGovernanceParams) PaymentChannelGovernanceParams {
	p.MinimumChannelCollateral = normalizeAmountOrDefault(p.MinimumChannelCollateral, defaults.MinimumChannelCollateral)
	p.MaximumChannelCollateral = normalizeAmountOrDefault(p.MaximumChannelCollateral, defaults.MaximumChannelCollateral)
	if p.MinimumChallengePeriod == 0 {
		p.MinimumChallengePeriod = defaults.MinimumChallengePeriod
	}
	if p.MaximumChallengePeriod == 0 {
		p.MaximumChallengePeriod = defaults.MaximumChallengePeriod
	}
	if p.DefaultChallengePeriod == 0 {
		p.DefaultChallengePeriod = defaults.DefaultChallengePeriod
	}
	if p.MinimumCloseDelay == 0 {
		p.MinimumCloseDelay = defaults.MinimumCloseDelay
	}
	if p.MaximumCloseDelay == 0 {
		p.MaximumCloseDelay = defaults.MaximumCloseDelay
	}
	p.ChannelOpenBaseFee = normalizeAmountOrDefault(p.ChannelOpenBaseFee, defaults.ChannelOpenBaseFee)
	p.ChannelStorageFeePerByte = normalizeAmountOrDefault(p.ChannelStorageFeePerByte, defaults.ChannelStorageFeePerByte)
	if p.ChannelTombstoneRetention == 0 {
		p.ChannelTombstoneRetention = defaults.ChannelTombstoneRetention
	}
	return p
}

func (p PaymentChannelGovernanceParams) Validate() error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Channel)
	minCollateral, err := parsePositiveInt("payments governance minimum channel collateral", params.MinimumChannelCollateral)
	if err != nil {
		return err
	}
	maxCollateral, err := parsePositiveInt("payments governance maximum channel collateral", params.MaximumChannelCollateral)
	if err != nil {
		return err
	}
	if maxCollateral.LT(minCollateral) {
		return errors.New("payments governance maximum channel collateral must be >= minimum")
	}
	if params.MinimumChallengePeriod < MinChallengePeriod || params.MaximumChallengePeriod > MaxChallengePeriod || params.MinimumChallengePeriod > params.MaximumChallengePeriod {
		return fmt.Errorf("payments governance challenge period bounds must be within %d and %d", MinChallengePeriod, MaxChallengePeriod)
	}
	if params.DefaultChallengePeriod < params.MinimumChallengePeriod || params.DefaultChallengePeriod > params.MaximumChallengePeriod {
		return errors.New("payments governance default challenge period must fit bounds")
	}
	if params.MinimumCloseDelay < MinCloseDelay || params.MaximumCloseDelay > MaxCloseDelay || params.MinimumCloseDelay > params.MaximumCloseDelay {
		return fmt.Errorf("payments governance close delay bounds must be within %d and %d", MinCloseDelay, MaxCloseDelay)
	}
	if err := validatePositiveInt("payments governance channel open base fee", params.ChannelOpenBaseFee); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments governance channel storage fee per byte", params.ChannelStorageFeePerByte); err != nil {
		return err
	}
	if params.ChannelTombstoneRetention == 0 {
		return errors.New("payments governance tombstone retention must be positive")
	}
	return nil
}

func (p PaymentChannelGovernanceParams) ValidateOpenCollateral(collateralText string) error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Channel)
	collateral, err := parsePositiveInt("payments governance open collateral", collateralText)
	if err != nil {
		return err
	}
	minCollateral, err := parsePositiveInt("payments governance minimum channel collateral", params.MinimumChannelCollateral)
	if err != nil {
		return err
	}
	maxCollateral, err := parsePositiveInt("payments governance maximum channel collateral", params.MaximumChannelCollateral)
	if err != nil {
		return err
	}
	if collateral.LT(minCollateral) {
		return errors.New("payments open collateral below governance minimum")
	}
	if collateral.GT(maxCollateral) {
		return errors.New("payments open collateral above governance maximum")
	}
	return nil
}

func (p PaymentConditionalGovernanceParams) Normalize(defaults PaymentConditionalGovernanceParams) PaymentConditionalGovernanceParams {
	if p.MaximumActivePromisesPerChannel == 0 {
		p.MaximumActivePromisesPerChannel = defaults.MaximumActivePromisesPerChannel
	}
	if p.MaximumPromiseAmountRatioBps == 0 {
		p.MaximumPromiseAmountRatioBps = defaults.MaximumPromiseAmountRatioBps
	}
	if p.MinimumTimeoutMargin == 0 {
		p.MinimumTimeoutMargin = defaults.MinimumTimeoutMargin
	}
	if p.MaximumPromiseLifetime == 0 {
		p.MaximumPromiseLifetime = defaults.MaximumPromiseLifetime
	}
	if p.BatchResolutionMaximumSize == 0 {
		p.BatchResolutionMaximumSize = defaults.BatchResolutionMaximumSize
	}
	p.PromiseStorageFee = normalizeAmountOrDefault(p.PromiseStorageFee, defaults.PromiseStorageFee)
	if p.ExpiredPromiseCleanupLimitPerBlock == 0 {
		p.ExpiredPromiseCleanupLimitPerBlock = defaults.ExpiredPromiseCleanupLimitPerBlock
	}
	return p
}

func (p PaymentConditionalGovernanceParams) Validate() error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Conditional)
	if params.MaximumActivePromisesPerChannel == 0 || params.MaximumActivePromisesPerChannel > MaxConditionsPerState {
		return fmt.Errorf("payments governance active promises per channel must be between 1 and %d", MaxConditionsPerState)
	}
	if params.MaximumPromiseAmountRatioBps == 0 || params.MaximumPromiseAmountRatioBps > 10_000 {
		return errors.New("payments governance promise amount ratio must be between 1 and 10000 bps")
	}
	if params.MinimumTimeoutMargin == 0 {
		return errors.New("payments governance minimum timeout margin must be positive")
	}
	if params.MaximumPromiseLifetime < params.MinimumTimeoutMargin {
		return errors.New("payments governance maximum promise lifetime must cover timeout margin")
	}
	if params.BatchResolutionMaximumSize == 0 || params.BatchResolutionMaximumSize > MaxSettlementBatchOps {
		return fmt.Errorf("payments governance batch resolution maximum size must be between 1 and %d", MaxSettlementBatchOps)
	}
	if err := validateNonNegativeInt("payments governance promise storage fee", params.PromiseStorageFee); err != nil {
		return err
	}
	if params.ExpiredPromiseCleanupLimitPerBlock == 0 {
		return errors.New("payments governance expired promise cleanup limit must be positive")
	}
	return nil
}

func (p PaymentConditionalGovernanceParams) ValidatePromiseWindow(channel ChannelRecord, promise ConditionalPromise) error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Conditional)
	channel = channel.Normalize()
	promise = promise.Normalize()
	if promise.TimeoutHeight <= channel.OpenHeight {
		return errors.New("payments governance promise timeout must be after channel open")
	}
	lifetime := promise.TimeoutHeight - channel.OpenHeight
	if lifetime > params.MaximumPromiseLifetime {
		return errors.New("payments promise lifetime exceeds governance maximum")
	}
	maxHeight := channel.LatestState.Normalize().TimeoutHeight
	if maxHeight == 0 {
		maxHeight = channel.OpenHeight + channel.CloseDelay + channel.DisputePeriod
	}
	if promise.TimeoutHeight+params.MinimumTimeoutMargin < promise.TimeoutHeight || promise.TimeoutHeight+params.MinimumTimeoutMargin > maxHeight {
		return errors.New("payments promise timeout does not leave governance timeout margin")
	}
	return nil
}

func normalizeAmountOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return strings.TrimSpace(fallback)
	}
	return value
}
