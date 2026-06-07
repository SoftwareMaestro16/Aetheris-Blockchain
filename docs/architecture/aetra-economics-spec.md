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
