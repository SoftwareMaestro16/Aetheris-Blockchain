# x/aetra-economics Module Specification

Purpose: low/moderate inflation, fee burn, treasury allocation, reward smoothing, and transparent APR model.

This module is the economic-control module of Aetra. It must avoid high-APR inflation traps while keeping validator/delegator rewards understandable, bounded, and auditable.

## Responsibilities

The module must:

- implement low/moderate inflation;
- implement fee burn;
- implement treasury allocation;
- implement reward smoothing;
- expose a transparent APR model.

## Implementation Contract

The implementation gate is `app/params/aetra_economics_spec.go`.

Required catalog properties:

- `AetraEconomicsModuleName` must be `x/aetra-economics`;
- `DefaultAetraEconomicsSpecEvidence` must cover low/moderate inflation, fee burn, treasury allocation, reward smoothing, and transparent APR model;
- `BuildAetraEconomicsSpecReport` must require all purpose components from this document;
- `ValidateAetraEconomicsSpec` must reject incomplete evidence;
- missing purpose components must fail validation;
- wrong or missing module identity must fail validation.

## 23.1 Responsibilities

The module must:

- calculate dynamic inflation;
- track bonded ratio;
- estimate staking APR;
- split fees;
- burn configured fee share;
- send configured share to distribution/rewards;
- send configured share to treasury;
- smooth reward changes;
- expose economic metrics;
- protect supply invariants.

### Responsibilities Implementation Contract

Required catalog properties:

- `AetraEconomicsResponsibilitiesEvidence` must represent every responsibility listed in section 23.1;
- `DefaultAetraEconomicsResponsibilitiesEvidence` must enable all required responsibilities;
- `BuildAetraEconomicsResponsibilitiesReport` must require all ten responsibilities;
- `ValidateAetraEconomicsResponsibilities` must reject missing responsibilities;
- wrong or missing module identity must fail validation;
- the responsibility catalog must be tested as a production requirement, not treated as narrative-only documentation.

## 23.2 State

Suggested state:

```text
Params:
  InflationMinBps
  InflationMaxBps
  InflationChangeRateBps
  TargetBondedRatioBps
  BurnFeeShareBps
  RewardFeeShareBps
  TreasuryFeeShareBps
  RewardSmoothingEpochs
  AprTargetMinBps
  AprTargetMaxBps

EpochEconomics:
  EpochNumber
  StartHeight
  EndHeight
  BondedRatioBps
  InflationBps
  EstimatedAprBps
  FeesCollected
  FeesBurned
  FeesToRewards
  FeesToTreasury
  MintedRewards

SupplyStats:
  TotalMinted
  TotalBurned
  NetIssuance
```

State requirements:

- `Params` must hold governance-controlled economic bounds and fee split configuration;
- `EpochEconomics` must expose deterministic epoch-level accounting for inflation, APR estimate, fees, burn, treasury, and minted rewards;
- `SupplyStats` must track minted, burned, and net issuance totals;
- all balances and rates must use deterministic integer or SDK decimal accounting, never floating point;
- state must be export/import safe and query-friendly for explorers, dashboards, wallets, and audit tooling.

### State Implementation Contract

Required catalog properties:

- `AetraEconomicsStateSpecEvidence` must list all required `Params`, `EpochEconomics`, and `SupplyStats` fields;
- `DefaultAetraEconomicsStateSpecEvidence` must include all twenty-four state fields from section 23.2;
- `BuildAetraEconomicsStateSpecReport` must reject missing, duplicate, and unexpected state fields;
- `ValidateAetraEconomicsStateSpec` must reject wrong module identity and incomplete state catalogs;
- state field coverage must be tested before implementation work is considered complete.

## 23.3 Inflation curve

Inflation should respond to bonded ratio:

```text
if bonded_ratio < target:
  increase inflation gradually

if bonded_ratio > target:
  decrease inflation gradually
```

Hard requirements:

- inflation never below min;
- inflation never above max;
- inflation change per epoch bounded;
- no floating point;
- no per-block instability;
- all calculations deterministic.

### Inflation Curve Implementation Contract

Required catalog properties:

- `AetraEconomicsInflationCurveEvidence` must represent every hard requirement listed in section 23.3;
- `DefaultAetraEconomicsInflationCurveEvidence` must enable all inflation curve requirements;
- `BuildAetraEconomicsInflationCurveReport` must require directionality, min/max bounds, bounded epoch changes, integer accounting, epoch-level stability, and determinism;
- `ValidateAetraEconomicsInflationCurve` must reject missing requirements;
- `x/aetra-economics/types.ComputeInflationBps` must calculate the deterministic bonded-ratio target without floating point;
- `x/aetra-economics/types.ComputeNextInflationBps` must move toward the target by no more than `InflationChangeRateBps` per epoch;
- `x/aetra-economics/types.ApplyEpoch` must use the bounded next-inflation calculation, not an unbounded direct jump to target.

## 23.4 Fee split rules

Fee split must always sum to 100%.

Recommended initial range:

```text
BurnFeeShareBps: 3000-6000
RewardFeeShareBps: 2000-4000
TreasuryFeeShareBps: 1000-2000
```

Example:

```text
50% burn
35% validators/delegators
15% treasury
```

The module must reject fee split params if:

- sum != 10000 bps;
- any share is negative;
- burn share exceeds max governance bound;
- treasury share exceeds max governance bound;
- rewards share is zero unless explicitly permitted by emergency governance.

### Fee Split Implementation Contract

Required catalog properties:

- `AetraEconomicsFeeSplitRulesEvidence` must represent every fee split rule listed in section 23.4;
- `DefaultAetraEconomicsFeeSplitRulesEvidence` must enable all fee split requirements;
- `BuildAetraEconomicsFeeSplitRulesReport` must require the 10000 bps sum rule, initial governance ranges, invalid sum rejection, negative share rejection, burn/treasury max-bound rejection, and emergency-only zero rewards;
- `ValidateAetraEconomicsFeeSplitRules` must reject missing fee split requirements;
- `x/aetra-economics/types.Params.Validate` must reject invalid fee split params before keeper state is updated;
- `x/aetra-economics/types.ComputeFeeSplit` must allocate burn, validators/delegators, and treasury from validated bps only;
- `x/aetra-economics/keeper.QueryFeeSplitParams` must expose current shares and governance bounds.

## 23.5 APR query

APR query must clearly distinguish:

- inflation-only APR;
- fee-adjusted APR;
- validator commission impact;
- estimated delegator APR;
- estimated validator gross APR;
- estimated validator net APR.

APR must be labeled as estimate, not guaranteed return.

### APR Query Implementation Contract

Required catalog properties:

- `AetraEconomicsAPRQueryEvidence` must represent every APR query output listed in section 23.5;
- `DefaultAetraEconomicsAPRQueryEvidence` must enable all APR query requirements;
- `BuildAetraEconomicsAPRQueryReport` must require inflation-only APR, fee-adjusted APR, validator commission impact, estimated delegator APR, estimated validator gross APR, estimated validator net APR, and estimate labeling;
- `ValidateAetraEconomicsAPRQuery` must reject missing APR query requirements;
- `x/aetra-economics/types.EstimateAPRBreakdown` must calculate APR values deterministically without floating point;
- `x/aetra-economics/keeper.QueryEstimatedAPR` must expose the full APR breakdown;
- APR query responses must set `IsEstimate` and `EstimateLabel`, and must not present APR as guaranteed return.

## 23.6 Tests

Required tests:

- inflation increases when bonded ratio below target;
- inflation decreases when bonded ratio above target;
- inflation remains within min/max;
- inflation change rate bounded;
- fee split exact accounting;
- burn accounting;
- treasury accounting;
- rewards accounting;
- APR estimate math;
- zero-fee block handling;
- high-fee block handling;
- export/import economics state;
- supply invariant after many epochs;
- governance invalid params rejected.

### Tests Implementation Contract

Required catalog properties:

- `AetraEconomicsTestingRequirementsEvidence` must represent every test requirement listed in section 23.6;
- `DefaultAetraEconomicsTestingRequirementsEvidence` must enable all required tests;
- `BuildAetraEconomicsTestingRequirementsReport` must require all fourteen test categories;
- `ValidateAetraEconomicsTestingRequirements` must reject missing test coverage;
- `x/aetra-economics/types` tests must cover inflation, fee split, burn, treasury, rewards, APR math, zero/high fee blocks, and many-epoch supply invariants;
- `x/aetra-economics/keeper` tests must cover export/import economics state and governance invalid params rejected.
