# Core Module Architecture

Date: 2026-06-05

Aetra uses Cosmos SDK core modules where they are already consensus-safe and
wraps them with Aetra-specific address, denom, fee, and zero-address policy.
Do not duplicate SDK core modules under new names unless a future design proves
that replacement is safer than bounded integration.

## CORE

### `x/auth`

Responsibilities:

- account model;
- signature verification;
- sequence and nonce replay protection;
- transaction validation;
- signer extraction;
- address validation;
- zero-address rejection before state mutation.

Aetra policy:

- account addresses are formatted through `app/addressing`;
- valid public account formats are raw `4:` and userfriendly `AE...`;
- old public `0:`, `orb1`, and `ORB` formats are rejected outside explicit
  migration tooling;
- identical signed tx bytes must not execute twice;
- wrong chain-id signatures fail before balance or sequence mutation;
- invalid signers fail before message state mutation;
- account metadata for indexers is exposed through `app/indexer` helpers using
  Aetra address formatting.

### `x/bank`

Responsibilities:

- native `AET` transfers;
- `naet` balances;
- module account accounting;
- mint/burn permission enforcement;
- multi-asset readiness for non-native user assets.

Aetra policy:

- native transfer denom is `naet`;
- protocol fees are not bank-transfer assets and remain enforced by `x/fees`;
- user-created denoms, IBC assets, LP denoms, NFT/SBT assets, and display denom
  `AET` cannot satisfy base-chain protocol fees;
- deterministic transfer events are regression-tested;
- resolver-based transfer targets and memo metadata are future execution-layer
  features and must resolve/validate before funds move.

### `x/staking`

Responsibilities:

- validator set;
- delegation;
- unbonding;
- redelegation;
- voting power updates;
- staking reward eligibility.

Aetra policy:

- bond denom is fixed to `naet`;
- validator commission bounds are modeled in `app/params`:
  - min commission `1%`;
  - max commission `20%`;
  - max daily commission change `1%`;
- validator lifecycle, CometBFT validator updates, restart/export/import, and
  snapshot/state-sync behavior are public-testnet gates.

### `x/slashing`

Responsibilities:

- downtime penalties;
- double-sign penalties;
- jailing and tombstone behavior;
- consensus security enforcement.

Aetra policy:

- slash fractions and missed-block windows must be positive and bounded;
- downtime recovery must be explicit and tested;
- slashed supply accounting must be deterministic;
- burn-vs-redistribution for slashed stake is an explicit economic policy
  decision, not an implicit bank movement.

### `x/gov`

Responsibilities:

- parameter updates;
- software upgrades;
- treasury controls;
- emergency controls.

Aetra policy:

- governance can update soft params only within hard-coded protocol bounds;
- hard economic bounds, such as inflation caps and fee caps, require software
  upgrade to change;
- governance authority tests are required for fees, token/economy params,
  domain params, reputation params, VM params, and emergency procedures.

### `x/distribution`

Responsibilities:

- reward distribution;
- validator commission;
- delegator reward accounting;
- community pool and treasury integration.

Aetra policy:

- mint rewards provide the validator baseline security budget;
- fee rewards provide activity-linked upside;
- deterministic integer rounding must preserve total accounting;
- reward withdrawal and validator commission behavior must remain covered by
  unit and integration tests.

## Test Evidence

The current regression surface includes:

- address codec tests for `4:` and `AE...`;
- replay, wrong-chain, invalid signer, malformed tx, fee-denom, and insufficient
  funds integration tests;
- deterministic bank transfer event test;
- staking create/delegate/unbond/redelegate/reward/slashing-param tests;
- `app/indexer` account metadata tests;
- fees governance and hard-bound tests;
- export/import and deterministic genesis tests.

