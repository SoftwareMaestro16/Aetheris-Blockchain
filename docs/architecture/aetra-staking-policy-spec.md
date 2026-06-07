# x/aetra-staking-policy Module Specification

Purpose: control effective voting power, delegation overflow, commission policy, and anti-concentration incentives.

This module is the central anti-centralization module of Aetra.

## Responsibilities

The module must:

- calculate raw validator stake;
- calculate effective validator stake;
- calculate overflow stake;
- enforce or expose effective voting power cap;
- calculate reward multiplier for overflow stake;
- expose delegation concentration warnings;
- enforce commission floor;
- enforce max commission;
- enforce max commission change rate;
- expose top-N concentration metrics;
- validate governance param changes;
- emit events for cap/overflow/commission policy changes;
- remain deterministic and export/import safe.

## Production Rule

`x/aetra-staking-policy` is not complete when only cap math exists. The production definition is:

```text
staking policy = stake math + cap enforcement/exposure + overflow accounting + commission policy + concentration metrics + governance params + events + export/import safety + tests + docs
```

Every responsibility must be represented in code, genesis/governance parameter validation, query surface, events where state changes, and tests. If a responsibility is temporarily query-only instead of enforcement, the behavior must be explicit and covered by tests so the chain does not silently present fake anti-centralization guarantees.

## Implementation Contract

The implementation gate is `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `AetraStakingPolicyModuleName` must be `x/aetra-staking-policy`;
- `DefaultAetraStakingPolicySpecEvidence` must cover the module purpose and central anti-centralization role;
- `BuildAetraStakingPolicySpecReport` must require all responsibilities from this document;
- `ValidateAetraStakingPolicySpec` must reject incomplete evidence;
- missing stake math, cap, overflow, reward multiplier, concentration warnings, commission controls, top-N metrics, governance validation, policy-change events, or deterministic export/import safety must fail validation;
- module identity must fail validation when missing or not equal to `x/aetra-staking-policy`.

## Required Tests

The module specification tests must prove:

- the default evidence covers all responsibilities;
- raw stake, effective stake, overflow stake, effective voting power cap, and overflow reward multiplier are mandatory;
- delegation concentration warnings, commission floor, max commission, max commission change rate, top-N concentration metrics, governance param validation, policy events, and export/import safety are mandatory;
- purpose and central anti-centralization role are mandatory;
- wrong or missing module identity is rejected.

## 22.2 State

Suggested state:

```text
Params:
  MaxValidatorsSoftTarget
  ValidatorPowerCapBps
  ValidatorPowerCapSchedule
  OverflowRewardMultiplierBps
  CommissionFloorBps
  CommissionMaxBps
  CommissionMaxDailyChangeBps
  Top10TargetBps
  Top20TargetBps
  Top33TargetBps
  MinSelfBond
  MinValidatorBond
  WarningThresholdBps

ValidatorPolicy:
  OperatorAddress
  RawBondedTokens
  EffectiveBondedTokens
  OverflowBondedTokens
  EffectivePowerBps
  IsOverCap
  RewardMultiplierBps
  LastCommissionChangeTime
  LastCommissionRateBps

ConcentrationSnapshot:
  Height
  BondedRatio
  ActiveValidators
  Top10Bps
  Top20Bps
  Top33Bps
  NakamotoCoefficientEstimate
```

All decimal values should use integer basis points or SDK decimal types consistently. Avoid floating point.

### State Requirements

`Params` must contain all governance-controlled and genesis-validated knobs needed for validator set sizing, effective power cap, overflow reward treatment, commission bounds, concentration targets, bond minimums, and warning thresholds.

`ValidatorPolicy` must be derived deterministically from staking state and policy params. It must expose raw bonded tokens, effective bonded tokens, overflow bonded tokens, effective power in basis points, over-cap status, reward multiplier, and commission-change tracking.

`ConcentrationSnapshot` must be a deterministic network-level view for observability, queries, dashboards, governance alerts, and export/import. It must record height, bonded ratio, active validators, top-10/top-20/top-33 concentration, and a Nakamoto coefficient estimate.

### State Implementation Contract

The state gate is `BuildAetraStakingPolicyStateSpecReport` in `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `DefaultAetraStakingPolicyStateSpecEvidence` must list every required `Params`, `ValidatorPolicy`, and `ConcentrationSnapshot` field;
- missing required fields must fail validation;
- duplicate fields must fail validation;
- unexpected fields must fail validation;
- module identity must be `x/aetra-staking-policy`;
- decimal accounting must explicitly use integer basis points or SDK decimal types;
- floating point accounting must fail validation.
