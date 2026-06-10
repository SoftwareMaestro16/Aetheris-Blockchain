# Additional Modules

Track 7 defines supporting modules that harden execution, authorization, and
queries without changing consensus wiring yet. Current packages are executable
specifications only; they do not register SDK stores, keepers, module accounts,
genesis, CLI, or ABCI hooks.

## x/compute

Purpose:

- measure CPU/compute usage separately from simple tx gas;
- price expensive computation;
- protect validators from CPU abuse.

State:

- compute unit schedule;
- per-op cost;
- per-contract compute stats;
- per-block compute budget.

Tests:

- expensive contract charged more;
- compute cap enforced;
- compute accounting deterministic.

`x/compute/types` defines a compute unit schedule, block compute meter,
per-contract compute cap, per-block compute budget, deterministic contract
stats, and validation that rejects zero contract addresses.

## x/permissions

Purpose:

- ACL system for contracts and modules;
- resolver delegates;
- domain managers;
- contract extension permissions;
- governance-controlled permissions.

Rules:

- all permissions have owner, scope, expiry, and revocation path;
- permission checks are deterministic;
- no hidden superuser outside governance/emergency policy.

`x/permissions/types` defines grant, check, expiry, and revoke behavior for
contract extension, resolver delegate, domain manager, module ACL, governance,
and emergency scopes. Governance and emergency are explicit scopes, not
implicit bypasses.

## x/indexer

Purpose:

- fast query layer;
- state search;
- event search;
- memo search;
- domain lookup;
- asset discovery.

Rule:

- indexer must never be required for consensus.

`x/indexer/types` defines deterministic projection records, canonical fields,
bounded query results, and search indexes for state, event, memo, domain,
and asset records. ConsensusRequired() returns false to keep the consensus
boundary explicit.

## x/market

Purpose:

- market for compute, storage, and execution priority;
- bounded, deterministic, and non-extractive.

Rules:

- cannot replace base `naet` fee;
- cannot let wealthy users fully starve normal users;
- must be capped by scheduler fairness.

`x/market/types` defines a capped premium selection model for compute,
storage, and priority orders. Market premiums are optional and bounded; every
accepted order must still prove the base `naet` fee is paid. normal-user
reserved slots, per-account share caps, and scheduler fairness caps prevent
premium orders from fully starving ordinary traffic.

## x/scheduler-v2

Purpose:

- DAG execution engine;
- parallel tx scheduling;
- async actor mailbox planning.

Requirements:

- deterministic read/write set;
- deterministic conflict resolution;
- replayable schedule;
- identical result across validators.

`x/scheduler-v2/types` defines a replayable DAG planner with deterministic
topological batches, conflict serialization, replay hash calculation, and
actor mailbox planning. It remains an executable specification only until
parallel state access, proof replay, and validator-equivalence tests are wired
into the execution pipeline.
