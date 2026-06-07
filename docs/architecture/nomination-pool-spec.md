# Nomination Pool Detailed Specification

Nomination pools are important for accessibility, but they introduce accounting
and centralization risks.

The implementation gate is `app/params/nomination_pool_spec.go`. A feature is
not complete unless the pool model, delegator model, risk acknowledgements,
queries, genesis validation, export/import safety, and tests are present.

## 26. Nomination Pool Detailed Specification

Purpose: allow users to participate in staking without running validator
infrastructure while preserving deterministic accounting and limiting
centralization risk.

Nomination pools must not become an unbounded operator cartel layer. Pool
operators can coordinate deposits and validator selection, but the protocol must
expose pool accounting, commission, validator target, unbonding state, and
delegator share state.

## 26.1 Pool Model

Each pool should have:

```text
Pool:
  PoolId
  OperatorAddress
  ValidatorAddress
  TotalBonded
  TotalShares
  CommissionBps
  Status
  CreatedHeight
  UnbondingEntries
```

Delegator state:

```text
PoolDelegation:
  DelegatorAddress
  PoolId
  Shares
  PrincipalEstimate
  RewardsAccrued
```

### Model Contract

Required catalog properties:

- `AetraNominationPoolModuleName` must be `x/nominator-pool`;
- `DefaultAetraNominationPoolModelEvidence` must include all required `Pool`
  and `PoolDelegation` fields from section 26.1;
- `BuildAetraNominationPoolModelReport` must reject missing, duplicate, and
  unexpected fields;
- `ValidateAetraNominationPoolModel` must reject wrong module identity and
  incomplete risk acknowledgement;
- model coverage must acknowledge accessibility, deterministic accounting, and
  centralization risks.

### Current Implementation Mapping

Current `x/nominator-pool/types.NominatorPool` uses more explicit names while
preserving the section 26.1 model:

- `PoolId` maps to `PoolID`;
- `OperatorAddress` maps to `PoolOperator`;
- `ValidatorAddress` maps to `ValidatorTarget`;
- `TotalBonded` maps to `TotalBondedStake`;
- `TotalShares` maps to `TotalShares`;
- `CommissionBps` maps to `PoolCommissionBps`;
- `Status` maps to `Status`;
- `CreatedHeight` should be represented by pool creation height and must become
  explicit in stored state before production readiness if it is not already
  persisted by keeper event/state history;
- `UnbondingEntries` maps to `UnbondingQueue`.

Current `x/nominator-pool/types.DelegatorShare` maps to `PoolDelegation`:

- `DelegatorAddress` maps to `Delegator`;
- `PoolId` is implied by the owning pool and should be explicit in query
  responses;
- `Shares` maps to `Shares`;
- `PrincipalEstimate` must be derived deterministically from
  `shares * TotalBondedStake / TotalShares`;
- `RewardsAccrued` maps to `PendingRewards` plus deterministic reward index
  accounting.

### Accounting Requirements

Pool accounting must be deterministic:

- `TotalShares` must equal the sum of all delegator shares;
- deposits mint shares using deterministic integer math;
- withdrawals burn shares and create unbonding entries;
- `PrincipalEstimate` is an estimate derived from current pool accounting, not
  a guaranteed redemption amount;
- `RewardsAccrued` must be derived from reward index checkpoints and pending
  rewards;
- slash losses must reduce pool bonded value without corrupting share supply;
- export/import must preserve sorted pools, delegations, and unbonding entries.

### Centralization Risk Requirements

Pool operators must not hide concentration risk:

- pool operator address must be queryable;
- validator target must be queryable;
- pool commission must be bounded by governance params;
- pool status must be explicit;
- pool delegations must be visible enough for wallets and explorers to warn
  users about overloaded pools or validator targets;
- no mandatory KYC should be embedded into consensus pool admission.
