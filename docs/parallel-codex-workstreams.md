# Parallel Codex Workstreams

This document tracks the repository-level coordination rules for parallel Codex
work on the native account, staking, rent, reputation, proof, docs, and app
wiring plan. Each chat must pick exactly one workstream, stay inside its
ownership boundary, and expose cross-workstream needs through narrow
interfaces, fixtures, or TODOs instead of editing another workstream's package.

## Parallelization Rules

Start every chat by reading:

- `UPDATE.md`;
- `architecture.md`;
- `docs/cosmos-l1-skills.md`;
- the packages listed under the selected workstream.

Global rules:

- Do not start by editing `app.go` or global module wiring unless the selected
  workstream explicitly owns integration.
- Do not change address derivation, `AE...` format, `4:...` format, sequence
  semantics, or signature domains outside the address/account workstream.
- Sequence semantics and signature domains are address/account workstream only.
- Do not reintroduce user direct delegation to validators.
- Do not add native token/NFT/DEX modules; those remain contracts.
- Every workstream must add tests in the same change set as code.
- Every workstream must keep export/import and genesis validation in mind, even
  if final wiring is done by another workstream.
- Prefer small files split by `types`, `keys`, `keeper`, `messages`, `queries`,
  `events`, `genesis`, `migrations`, and tests.
- Shared structs must live in the owning module's `types` package; other
  modules consume interfaces or query methods.
- Avoid circular keeper dependencies. Cross-module access must use explicit
  interfaces.
- Run targeted package tests first, then `go test ./...` after integration work.

Branch/worktree rule:

```text
Use one branch per workstream, for example:
  codex/native-account
  codex/storage-rent
  codex/liquid-staking-pool
```

Do not commit unrelated dirty files. Merge small stable packages as they pass
tests; broad app wiring is owned by W14 after W0-W13 APIs stabilize.

## Shared Contracts

Address contract:

```text
user-facing account/validator/consensus/pool address = AE...
internal raw address = 4:...
AE... <-> 4:... roundtrip must be stable
```

Validator entry contract:

```text
active validators: 100-300 outside explicit testnet override
minimum validator entry: 1_000_000 AET
solo validator self-stake: 1_000_000 AET
pool-backed validator self-stake: >= 400_000 AET
pool-backed nominator/pool stake toward minimum entry: <= 600_000 AET
direct user validator delegation: disabled
```

User staking contract:

```text
User -> Liquid Staking Contract -> Pool Contract -> Validators
normal user chooses pool/index, not a validator
MinPoolDeposit = 10 AET
UnbondingPeriod = 18 days
```

Fee/rent contract:

```text
MinTxFee = 0.003 AET = 3_000_000 naet
StorageRentRate = 1 naet per byte-second
storage_size = code_bytes + data_bytes
effective_fee = gas_fee + storage_rent_delta + unpaid_storage_debt
zero balance + no state = free
zero balance + persistent state = debt + freeze, not delete
system/critical state = protocol-paid + no freeze
```

Frozen-state contract:

```text
frozen = recoverable, state intact, balance intact
archived = reduced state, recoverable only if enough metadata/proofs remain
deleted = state removed, not normally recoverable
```

## Dependency Graph

Independent chat workstream groups:

```text
CHAT 1 - Repository Baseline And Guardrails:
  W0 Address compatibility
  W1 Governance params schema
  W2 Native account/auth
  W3 Storage rent core

CHAT 2 - Validator Registry And Official Pool Entry:
  W4 Validator registry/policy
  W5 Liquid staking pool state
  W6 Contract capability hooks

CHAT 3 - Allocation, Rewards, Reputation, Proofs:
  W7 Allocation engine
  W8 Pool rewards
  W9 Stake reputation
  W10 Proofs/events

CHAT 4 - Genesis, Invariants, Docs, Final Wiring:
  W11 Genesis/migrations/export-import
  W12 Scalability/invariants
  W13 Docs/CLI/query surface
  W14 Final app wiring
```

Parallel execution rule:

```text
All independent groups can start at once.
```

Workstream rule:

```text
each workstream owns its packages
each workstream can add temporary local interfaces/fixtures
no workstream edits another workstream's owned files
final app wiring happens after feature package APIs stabilize
```

Integration/merge strategy:

```text
Merge stable leaf packages as they pass tests.
Merge shared interfaces before code that consumes them.
Leave broad app wiring to W14 after W0-W13 APIs are stable.
```

## Workstream Ownership

CHAT 1 owns foundational address, params, account/auth, and storage-rent work:

- W0 Address Compatibility owns `app/addressing`, address validation tests, and
  address docs snippets.
- W1 Governance Params Schema owns params structs, genesis param validation, and
  docs tables for genesis/governance params.
- W2 Native Account And Auth owns native account types/state, activation, auth
  policy, account queries, and account genesis.
- W3 Storage Rent Core owns storage-rent types/keeper, rent debt accounting,
  system storage reserve, and normal account/contract freeze interfaces.

CHAT 2 owns validator registry and official pool entry:

- W4 Validator Registry And Staking Policy owns validator registry/state,
  staking policy params, validator admission, commission bounds, and self-stake
  or pool-backed entry checks.
- W5 Liquid Staking Pool State owns pool state, pool share state, unbonding
  state, pool contract address pair, and pool deposit/unbond/withdraw message
  skeletons.
- W6 Contract Capability Hooks owns official liquid staking contract
  registration, native staking injection hooks, unauthorized contract rejection,
  and frozen/frozen_limited capability checks.

CHAT 3 owns allocation, rewards, reputation, proofs, and events:

- W7 Allocation Engine owns deterministic validator scoring, weights,
  allocation records, rebalancing, and bounded allocation loops.
- W8 Pool Rewards owns lazy reward index, reward claims, validator commission,
  pool fee deduction, and reward export/import.
- W9 Stake Reputation owns stake-time accumulators, reputation claims, and
  account-owned non-transferable reputation.
- W10 Proofs And Events owns proof metadata structs, deterministic staking
  events, event receipts, and no-secret event checks.

CHAT 4 owns genesis, hardening, docs, and final integration.

CHAT 4 goal: make the independently built pieces production-usable through
export/import, migrations, invariants, docs, query/CLI surface, and final app
wiring.

CHAT 4 outputs:

- deterministic genesis/export/import;
- versioned migrations;
- scalability checks;
- invariant registry;
- docs and examples;
- final keeper/app/module wiring;
- full test pass.

CHAT 4 workstreams must not rewrite:

- address derivation;
- account auth semantics;
- allocation math;
- reward math;
- storage rent semantics.

CHAT 4 owned workstreams:

- W11 Genesis, Migration, Export/Import owns deterministic full-state
  export/import, versioned migrations, lazy migration, and malformed genesis
  rejection before writes.
- W12 Scalability And Invariants owns bounded-iteration checks, million-user
  simulations, invariant registration, and failure fixtures.
- W13 Docs, CLI, Query Surface owns docs, query/CLI examples, static doc tests,
  and required examples validation.
- W14 Final App Wiring owns app module wiring, keeper injection, module order,
  integration tests, app boot, export/import restart, and final `go test ./...`.

## W11 Genesis, Migration, Export/Import Requirements

W11 ownership:

- genesis state for new modules;
- export/import validation;
- versioned migrations;
- lazy migration.

W11 tasks:

- Add deterministic export/import for accounts, pools, allocations, rewards,
  reputation, rent, and validator policy.
- Add versioned account/pool migration.
- Reject malformed duplicate state before writes.
- Preserve mixed account versions.

W11 depends on W2/W3/W4/W5/W7/W8/W9 types.

W11 must not touch business logic except migration handlers.

Required W11 tests:

- full export/import preserves all new state;
- duplicate account/pool/share/allocation rejected;
- unsupported version rejected safely;
- lazy migration preserves address and sequence.

## W12 Scalability And Invariants Requirements

W12 ownership:

- invariant registration;
- bounded-iteration tests;
- benchmarks/simulations.

W12 tasks:

- Assert no O(N users) BeginBlock/EndBlock paths.
- Add invariant tests for bank/module accounting, rewards cap, rent, pool
  shares, validator entry, and direct delegation rejection.
- Add million-user style simulation for pool shares and reward claims.

W12 depends on all core state modules.

W12 must not touch feature implementation except small instrumentation hooks.

Required W12 tests:

- BeginBlock/EndBlock bounded;
- reward claim bounded;
- reputation claim bounded;
- rent charge bounded;
- invariant failure fixtures.

## W13 Docs, CLI, Query Surface Requirements

W13 ownership:

- docs;
- CLI examples;
- query docs;
- static doc tests.

W13 tasks:

- Update docs to say normal users stake only through official liquid staking
  pools.
- Document validator entry:
  - `1_000_000 AET`;
  - solo full self-stake;
  - pool-backed `400_000/600_000`.
- Document `100-300` validator range.
- Document unbonding `18 days`.
- Document min tx fee `0.003 AET`.
- Document storage rent and recoverable freeze/unfreeze.

W13 depends on final names from W1/W2/W5.

W13 must not touch keeper logic.

Required W13 tests:

- static doc tests for required terms;
- command examples compile or are validated.

## W0 Address Compatibility Requirements

W0 owns:

- `app/addressing`;
- address validation tests;
- address docs snippets.

W0 tasks:

- Freeze `AE...` and `4:...` golden vectors.
- Add pool address helpers if missing: `FormatPoolAddress`,
  `ParsePoolAddress`, or explicit reuse of the account codec.
- Ensure only `AE...` and `4:...` address formats are used in user-facing account, validator,
  consensus, and pool APIs.
- Add stable `AE... <-> 4:...` roundtrip tests for accounts, validators,
  consensus addresses, and pools.

W0 must not touch staking keeper logic, storage rent accounting, or broad app
module wiring except codec registration if required.

W0 must not touch broad app module wiring.

Required W0 tests:

- account `AE...` roundtrip;
- validator `AE...` roundtrip;
- consensus `AE...` roundtrip;
- pool `AE...` roundtrip;
- raw `4:...` roundtrip;
- malformed legacy prefixes rejected.

## Safety Rules

A workstream may add local mock interfaces for another workstream, but must name
them clearly as temporary fixtures. A workstream must not edit another
workstream's owned packages to make tests pass. Cross-workstream changes should
be merged through explicit interfaces in the owner package. W14 is the only
workstream that should do broad `app` wiring after W0-W13 APIs are stable.
