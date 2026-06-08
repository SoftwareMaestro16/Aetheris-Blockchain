# Aetra Testnet Completion Backlog V2

This file is the practical backlog for turning Aetra into a runnable, testable,
validator-understandable public testnet candidate.

The goal of V2 is not to add more native modules. The goal is:

- one stable node binary;
- deterministic genesis;
- 4-5 node localnet;
- export/import restart;
- upgrade rehearsal;
- validator join guide;
- pool-based PoS that matches the product model;
- AVM v1 that can run real contracts;
- minimal, boring, validator-grade infrastructure.

## Non-Negotiable Direction

- Keep the runnable chain small.
- Do not introduce native DEX/token/NFT modules for application assets.
- Token, NFT, DEX, wallet standards, and domains must be AVM contracts or
  registries unless explicitly listed as system state.
- Normal user staking goes through the official liquid staking/nominator pool
  path. Direct user validator selection is not the normal UX.
- `AE...` remains the only user-facing address format.
- `4:...` remains raw/internal/proof format.
- Private keys, seed phrases, validator private keys, keyrings, and localnet
  secrets never appear in genesis, events, exports, logs, docs examples, or
  release artifacts.
- BeginBlock/EndBlock and block lifecycle must not scan all users/contracts.
- Export/import must preserve every consensus state field needed for restart.

## Current Repo Observations

This backlog was written after inspecting the repo state around:

- `cmd/l1d/cmd`;
- `scripts/localnet`;
- `scripts/testnet`;
- `scripts/release`;
- `.github/workflows/testnet-readiness.yml`;
- `x/contracts`;
- `x/aetravm/avm`;
- `x/aetravm/async`;
- `x/aetravm/standards`;
- `x/nominator-pool`;
- validator and testnet docs.

Observed status:

- Testnet readiness CI exists, but it must become the canonical release gate,
  not only an auxiliary workflow.
- Exit codes already exist under `x/contracts/types/exit_codes.go` and are small
  stable values. Keep them small; do not introduce huge opaque error numbers.
- AVM has a stack VM, gas table, host registry, storage ABI, async messages, and
  receipts/events work in progress.
- AVM has `EntryQuery`, but contract get methods are not yet a full user-facing,
  read-only contract query workflow.
- AVM storage currently looks more like bounded key/value storage. The target
  testnet AVM model should move persistent contract data toward immutable
  content-addressed Chunks and ChunkMap indexes.
- Docs still contain historical tokenfactory/DEX language in places. For V2,
  keep those as future AVM standards or remove from testnet launch docs.
- Docker image, seed/peer publication, Cosmovisor guide, and top-level
  `docs/VALIDATOR.md` / `docs/TESTNET.md` / `docs/COSMOVISOR.md` are not yet the
  single canonical operator path.

## Parallel Workstream Ownership

Use these ownership boundaries so multiple Codex chats can work in parallel.

### CHAT A - Testnet Core And Release Gate

Owned paths:

- `.github/workflows/*`;
- `cmd/l1d/cmd`;
- `scripts/localnet/*`;
- `scripts/testnet/*`;
- `scripts/release/*`;
- `app/genesis*`;
- `app/upgrades*`;
- `app/params/testnet_readiness*`;
- top-level release/testnet docs.

Do not touch:

- AVM instruction implementation except build/CLI integration;
- PoS accounting internals except smoke-test integration.

### CHAT B - PoS, Nominator Pool, Slashing, Validator Reputation

Owned paths:

- `x/nominator-pool`;
- `x/validator-registry`;
- `x/validator-election`;
- `x/validator-insurance`;
- `x/aetra-validator-score`;
- `x/aetra-staking-policy`;
- staking/slashing docs and e2e smoke tests.

Do not touch:

- AVM runtime implementation;
- release workflow except adding job commands agreed with CHAT A.

### CHAT C - AVM Runtime V1

Owned paths:

- `x/aetravm/avm`;
- `x/aetravm/async`;
- `x/aetravm/messageabi`;
- `x/aetravm/standards`;
- `x/contracts`;
- AVM examples and contract docs.

Do not touch:

- PoS allocation math;
- validator registry/election internals;
- app-level release workflow except adding AVM tests commands agreed with CHAT A.

### CHAT D - Infrastructure And Operator Docs

Owned paths:

- `docs/VALIDATOR.md`;
- `docs/TESTNET.md`;
- `docs/COSMOVISOR.md`;
- `docs/AVM.md`;
- `docs/HEALTH.md`;
- Docker files;
- release artifact documentation;
- peer/seed list publication templates.

Do not touch:

- consensus state internals;
- AVM opcode semantics;
- PoS accounting internals.

### CHAT E - Noise Reduction And Launch Scope

Owned paths:

- stale docs mentioning native DEX/token/NFT;
- module-boundary docs;
- app module wiring tests for "no native app asset modules";
- future standards docs under `x/aetravm/standards`.

Do not touch:

- active PoS/AVM implementation unless removing stale references requires tests.

## Phase 0 - Launch Scope Freeze

### Task 0.1 - Define Testnet Kernel

Implementation:

- Create a short `docs/TESTNET.md` that states the testnet kernel:
  - Cosmos SDK + CometBFT node;
  - native bank balance layer;
  - native account/auth/freeze/rent only where already wired;
  - pool-based staking;
  - AVM contracts;
  - no native token/NFT/DEX app modules for application assets.
- Add a machine-checkable launch scope test.
- Make the release workflow fail if docs teach normal users to stake by choosing
  validators directly.

Tests:

- Static doc test: no user-facing `aevaloper` / `aevalcons`.
- Static doc test: no native DEX/token/NFT launch instructions.
- Static doc test: user staking examples mention official pool deposit.

Done:

- Everyone can read one page and know what the testnet actually launches.

### Task 0.2 - Remove Or Quarantine Prototype Noise

Implementation:

- Identify modules/docs that are prototype-only or future-only.
- Move future concepts into `docs/future/` or mark them as AVM standards, not
  launch-critical modules.
- Ensure app wiring does not include native DEX/token/NFT modules.
- Keep `x/aetravm/standards/aft`, `anft`, `adex` as future AVM standards only
  unless they are runnable as contracts.

Tests:

- App module wiring test rejects `tokenfactory`, `dex`, `nft`, `market` native
  asset modules in launch profile.
- Release docs test rejects launch instructions for native application-asset
  modules.

Done:

- Testnet scope is small and not confused by prototype-era docs.

## Phase 1 - Runnable Testnet Core

### Task 1.1 - One Stable Binary

Implementation:

- Build one binary: `aetrad` / `aetrad.exe`.
- Ensure `aetrad version --long --output json` includes:
  - app name;
  - version;
  - git commit;
  - build date;
  - dirty flag;
  - Cosmos SDK version;
  - CometBFT version;
  - AVM version.
- Add release ldflags for version metadata.
- Release artifact must include the binary and checksums.

Tests:

- CLI test parses version JSON.
- Release script test verifies binary checksum file.
- CI job runs built binary version command.

Done:

- Validator can download/build one binary and verify what it is.

### Task 1.2 - Deterministic Genesis

Implementation:

- Make localnet genesis generation deterministic within one run across all nodes.
- Chain-id must pass `ValidateAetraTestnetChainID`.
- Genesis must reject:
  - malformed chain-id;
  - secrets;
  - wrong denom;
  - missing staking params;
  - invalid module params;
  - validator count outside local/testnet profile rules.
- Genesis validation command:
  - `aetrad genesis validate-genesis <path>`;
  - `scripts/localnet/validate-genesis.ps1`.

Tests:

- Genesis validate CI job.
- Golden test for default app genesis shape.
- Localnet script test: all node genesis files have identical hash.
- Negative tests: malformed chain-id, secret-like field, wrong denom.

Done:

- A launch operator can validate published genesis before starting.

### Task 1.3 - 4-5 Node Localnet

Implementation:

- Add canonical 4-node and 5-node localnet profiles.
- Keep 3-node smoke for fast CI; 4/5-node is launch rehearsal.
- Scripts:
  - init;
  - validate genesis;
  - start;
  - wait for height;
  - query validator set;
  - stop;
  - collect diagnostics.
- No secret material in diagnostics artifacts.

Tests:

- CI fast: 3-node smoke.
- Manual/release gate: 4-node and 5-node localnet smoke.
- Negative: occupied ports, missing binary, invalid validator count.

Done:

- Local testnet can be booted repeatedly with clear commands.

### Task 1.4 - Export/Import Roundtrip

Implementation:

- Export state after blocks and committed txs.
- Validate exported genesis.
- Import into fresh home.
- Restart from imported state.
- Verify app hash/critical state consistency where feasible.

Tests:

- `tests/e2e/export_import_smoke.ps1`.
- App-level export/import restart test.
- Negative: corrupted exported state is rejected.

Done:

- Runtime mutations survive restart/export/import.

### Task 1.5 - Upgrade Rehearsal

Implementation:

- Add a canonical no-op upgrade handler for rehearsal.
- Add upgrade dry-run:
  - pre-upgrade state;
  - version map;
  - handler execution;
  - post-upgrade export validation.
- Document Cosmovisor upgrade path.

Tests:

- `go test ./app/upgrades ./app -run Upgrade`.
- Localnet upgrade smoke:
  - schedule upgrade;
  - stop at height;
  - swap binary or no-op handler;
  - restart;
  - export validate.

Done:

- Testnet can rehearse a coordinated upgrade before public launch.

## Phase 2 - Validator-Grade Infrastructure

### Task 2.1 - Canonical Operator Docs

Create:

- `docs/VALIDATOR.md`;
- `docs/TESTNET.md`;
- `docs/COSMOVISOR.md`;
- `docs/HEALTH.md`;
- `docs/AVM.md`.

`docs/VALIDATOR.md` must cover:

- hardware;
- OS;
- build/download binary;
- version verification;
- chain-id;
- genesis validation;
- keyring;
- validator key safety;
- state sync;
- snapshots;
- create validator;
- monitor;
- restart;
- upgrade;
- incident response.

`docs/TESTNET.md` must cover:

- chain-id;
- genesis URL/checksum placeholder;
- seed nodes;
- persistent peers;
- RPC endpoints;
- faucet path if enabled;
- minimum fees;
- expected block time;
- launch profile;
- known non-goals.

`docs/COSMOVISOR.md` must cover:

- install;
- directory layout;
- current binary;
- upgrades directory;
- environment;
- upgrade handler naming;
- rollback policy.

Tests:

- Static doc coverage test for all required sections.
- Release package includes all docs.

Done:

- A validator can join without reading source code.

### Task 2.2 - Docker Image

Implementation:

- Add Dockerfile for `aetrad`.
- Add minimal runtime image.
- Add non-root user.
- Add healthcheck command.
- Add build args for version/commit.
- Add docker-compose localnet sample only if it does not replace existing
  PowerShell localnet scripts.

Tests:

- Docker build CI job.
- Container `aetrad version --long --output json` passes.
- Healthcheck command returns healthy against a local node.

Done:

- Release can publish a validator-ready image.

### Task 2.3 - Health Checks And Peer Lists

Implementation:

- Define health endpoints/commands:
  - process alive;
  - RPC status;
  - latest height increasing;
  - catching_up false;
  - peer count;
  - validator signing info;
  - app invariant command/test.
- Add `docs/HEALTH.md`.
- Add seed/persistent peer list templates:
  - `docs/testnet/peers.example.json`;
  - `docs/testnet/seeds.example.txt`.

Tests:

- Localnet health script validates 3-node/5-node profile.
- Peer list parser rejects malformed node IDs/endpoints.

Done:

- Operators can monitor and join peers with published data.

### Task 2.4 - Release Workflow

Implementation:

- Make testnet release workflow run:
  - `go test ./...`;
  - `go vet ./...`;
  - `buf lint`;
  - genesis validate;
  - localnet smoke;
  - export/import smoke;
  - invariants;
  - release artifact build;
  - binary version command;
  - Docker build.
- Upload:
  - binary;
  - checksums;
  - docs;
  - readiness report;
  - localnet diagnostics on failure.

Tests:

- Workflow static test contains all required jobs.
- Release package script test verifies required docs and checksums.

Done:

- A release candidate cannot be published without the runnable gates.

## Phase 3 - PoS V1 Completion

### Task 3.1 - Pool-Based User Staking

Implementation:

- Normal user staking path:
  - User `AE...`;
  - official liquid staking contract/pool;
  - pool shares/receipt;
  - allocation to validators.
- Normal CLI/API must not ask user for a validator address.
- Direct SDK `MsgDelegate` must be:
  - disabled for normal user path; or
  - explicitly guarded as operator-only/advanced governance-enabled path.
- Pool deposits must support small users above anti-spam minimum.

Tests:

- User deposits into pool without validator address.
- Deposit below minimum rejected.
- Direct user delegation rejected before staking mutation.
- Pool shares deterministic.
- Export/import preserves shares and pool totals.

Done:

- User staking UX is pool/index-based.

### Task 3.2 - Nominator Pool Accounting

Implementation:

- Pool state must live in KV/prefix records, not runtime-only memory.
- Mutating methods read/write KV.
- Export reads KV and emits deterministic sorted state.
- Import writes prefix records.
- No full scans in block lifecycle.
- Pool allocations update only touched keys.

Tests:

- mutate -> export -> import -> query same state.
- pagination bounded.
- deterministic order.
- storage rent debt preserved.

Done:

- Pool state survives restart and scales past prototype size.

### Task 3.3 - Reward Policy V1

Implementation:

- Define simple reward policy:
  - fees/inflation source;
  - pool share accounting;
  - validator commission;
  - lazy reward index;
  - rounding rule;
  - cap by collected/emitted rewards.
- Rewards distributed by pool shares, not manual validator choice.
- Reputation cannot increase without stake-time exposure.

Tests:

- reward cap invariant;
- deterministic reward index;
- claim idempotency;
- export/import after rewards;
- jailed/slashed validator does not produce positive bonus.

Done:

- Rewards are understandable and not open-ended.

### Task 3.4 - Slashing V1

Implementation:

- Ensure slashing params are genesis/governance params.
- Wire downtime/double-sign policy to validator status and pool exposure.
- Pool users inherit validator slashing exposure through allocation accounting.
- Slashed state export/import stable.

Tests:

- downtime slash fixture;
- double-sign/tombstone fixture if available;
- pool allocation principal decreases or records exposure;
- cannot recover slashed stake through export/import/migration.

Done:

- Validators and pool participants have deterministic slashing risk.

### Task 3.5 - Validator Score/Reputation V1

Implementation:

- Minimal deterministic score:
  - uptime;
  - missed blocks;
  - commission;
  - slashing risk;
  - stake efficiency;
  - pool allocation limit;
  - reputation accumulator.
- Score output drives allocation engine weights.
- No nondeterministic inputs.

Tests:

- same input -> same weights;
- score changes with uptime/commission/slashing;
- inactive/ineligible validators rejected;
- export/import preserves scores and snapshots.

Done:

- Allocation engine has a transparent v1 score.

## Phase 4 - AVM V1 Completion

### AVM Direction

Aetra VM should become a deterministic, stack-based, message-driven VM using
immutable content-addressed Chunks as the persistent data model.

Important constraint:

- Do not copy TON slice/builder exactly.
- Do not expose raw byte fiddling as the main developer experience.
- Prefer typed Reader/Writer/Codec over manual bit parsing.

### Task 4.1 - Small Stable Exit Codes

Current status:

- Small exit codes exist in `x/contracts/types/exit_codes.go`.

Implementation:

- Keep exit codes under `100` for core VM/contract errors.
- Define one canonical list:
  - ok;
  - validation failed;
  - unauthorized;
  - inactive/frozen;
  - code rejected;
  - out of gas;
  - storage limit;
  - storage rent debt;
  - message expired;
  - queue limit;
  - execution failed;
  - internal bounce;
  - forbidden host call;
  - contract abort.
- Map AVM runtime exit codes to contract receipt exit codes.
- Add `ExitCodeName` coverage for every runtime code.

Tests:

- golden exit code list;
- all codes `< 100`;
- unknown returns `unknown`;
- receipt stores code and name/proof metadata.

Done:

- Operators and contract developers can understand failures.

### Task 4.2 - Chunk Core

Implementation:

Define AVM Chunk:

```text
Chunk {
  data_bits <= 2048
  refs <= 8
  type_hash optional
  hash = BLAKE3(canonical_chunk_encoding)
}
```

Rules:

- immutable;
- content-addressed;
- DAG only;
- cycles rejected;
- identical chunk -> identical hash;
- hash is stable across export/import;
- refs are ordered and bounded.

Tests:

- hash golden vectors;
- data over 2048 bits rejected;
- refs over 8 rejected;
- cycle rejected;
- canonical encoding stable;
- export/import preserves hash/root.

Done:

- AVM state can be represented as content-addressed immutable chunks.

### Task 4.3 - Type System And Codec

Runtime primitive types:

- `bool`;
- `uint8`, `uint16`, `uint32`, `uint64`, `uint128`, `uint256`;
- `int8`, `int16`, `int32`, `int64`, `int128`, `int256`;
- `address`;
- `hash`;
- `coins = uint128`;
- `timestamp = uint64`;
- `null`;
- `tuple`;
- `chunk`;
- `execution_frame`.

Compile-time/developer types:

- `struct`;
- `Option<T>`;
- `Tuple<T...>`;
- `Map<K,V>` compiled to ChunkMap;
- bounded UTF-8 `string`;
- bytes.

Implementation:

- VM stores universal values, not runtime generics.
- Add optional `type_hash = BLAKE3(schema_descriptor)`.
- Add `Reader<T>` as read cursor over Chunk data.
- Add `Writer<T>` as immutable Chunk constructor.
- Add `Codec<T>` descriptors:
  - canonical schema string;
  - encode;
  - decode;
  - gas cost;
  - max encoded size.
- Strings:
  - UTF-8;
  - byte-length bounded;
  - encoded as length-prefixed bytes;
  - no unbounded string concatenation.

Tests:

- primitive encoding golden vectors;
- string UTF-8 valid/invalid;
- Option null/value encoding;
- type_hash stable;
- invalid decode reverts;
- same typed value -> same chunk hash.

Done:

- Contracts can be written with short typed code, not manual byte parsing.

### Task 4.4 - ChunkMap

Do not implement EVM-style mutable mapping.

Implementation:

Define persistent Chunk Trie Map:

```text
ChunkMap {
  root: Chunk
  fanout: 8
  key_hash = BLAKE3(canonical_key)
  path = deterministic nibbles/buckets from key_hash
  leaf = value Chunk
}
```

Rules:

- immutable tree;
- lazy node creation;
- update copies only changed path;
- no global mutable hashmap;
- no full scan for lookup/update;
- deterministic iteration only through bounded proof/index APIs;
- parallel-friendly buckets.

Tests:

- put/get/delete;
- update changes only path root;
- same operations -> same root;
- different buckets can be proven independent;
- key collision handling;
- export/import preserves root;
- gas depends on depth and encoded bytes.

Done:

- Domains, token ownership, NFT ownership, balances inside contracts, and other
  contract maps use ChunkMap, not global storage slots.

### Task 4.5 - AVM Execution Model

Execution phases:

1. Storage Phase - load state Chunks.
2. Credit Phase - apply attached value.
3. Compute Phase - execute VM.
4. Action Phase - emit outgoing messages/events.
5. Finalization Phase - commit new Chunk roots.

Implementation:

- ExecutionFrame:
  - instruction pointer;
  - stack snapshot;
  - local context;
  - pending calls/messages;
  - error handler/abort state.
- Stack values are typed VM values.
- No classic RAM or linear mutable memory.
- All state updates produce new Chunks.
- Out-of-gas reverts state changes but records receipt.
- Per-message and per-block gas limits enforced.

Tests:

- deploy;
- execute external;
- execute internal;
- out-of-gas rollback;
- abort exit code;
- same input/state/code -> same root/gas/receipt;
- forbidden nondeterministic opcode rejected.

Done:

- AVM can execute deterministic contracts safely.

### Task 4.6 - Host Functions V1

Allowed:

- hash SHA256;
- hash BLAKE3;
- verify ed25519;
- parse/format Aetra address;
- read storage chunk;
- write storage chunk;
- delete storage chunk;
- emit event;
- send internal message;
- get block height;
- get chain id;
- get contract address;
- get caller/source;
- get attached value;
- abort with exit code.

Careful:

- Wall-clock time is forbidden.
- If `time.now()` is added, it must mean consensus block time from header, not
  local process clock.
- Randomness must not be process randomness. If `secure_random()` is added, it
  must be a deterministic/verifiable block entropy value, e.g.:

```text
random = BLAKE3(previous_state_root || block_entropy || message_hash || domain)
```

and only after the chain defines block entropy/proof rules.

Forbidden:

- filesystem;
- network;
- floating point;
- goroutines/threads;
- process/env;
- nondeterministic map iteration;
- local wall-clock;
- unverified randomness.

Tests:

- each allowed host has gas cost;
- unknown host rejected;
- forbidden host rejected;
- host storage respects Chunk/ChunkMap limits;
- send_internal respects queue limits;
- abort returns contract-defined small exit code.

Done:

- Host surface is deterministic and auditable.

### Task 4.7 - Get Methods

Implementation:

- Add contract get methods as read-only AVM entrypoints.
- Get method call:
  - loads code and state root;
  - executes with query gas limit;
  - cannot write storage;
  - cannot send internal messages;
  - cannot emit consensus events;
  - can return typed value bytes/Chunk;
  - can include proof metadata.
- CLI/API:
  - `aetrad query contracts get <contract> <method> <args-json>`;
  - gRPC query endpoint;
  - bounded response size.

Tests:

- get method reads state;
- attempted write rejected;
- attempted send message rejected;
- gas limit enforced;
- response deterministic;
- malformed args rejected;
- proof query stable.

Done:

- Contracts have practical read APIs without state mutation.

### Task 4.8 - Minimal Contract Examples

Examples must be real AVM examples, not native modules:

- counter contract;
- key/value ChunkMap contract;
- domain registry sample:
  - name string;
  - owner AE address;
  - resolver records map;
- token sample using AFT standard;
- NFT sample using ANFT standard;
- pool deposit adapter sample for official staking if applicable.

Tests:

- examples compile/encode;
- deploy;
- execute;
- get methods;
- export/import;
- receipts/events.

Done:

- A developer can write and run a minimal contract.

## Phase 5 - Contract Standards Instead Of Native App Assets

### Task 5.1 - Move Token/NFT/DEX To AVM Standards

Implementation:

- Keep standards:
  - AFT for fungible tokens;
  - ANFT for NFTs/SBTs;
  - ADEX as future DEX standard.
- Do not wire native app modules for these.
- Launch docs should say:
  - token/NFT/DEX are contract standards;
  - DEX is not required for initial public testnet unless AVM contract example is
    ready.

Tests:

- App wiring rejects native asset modules.
- Docs do not teach native tokenfactory/DEX launch flows.
- Standards registry has deterministic descriptors.

Done:

- Asset story is clear and not split across native and AVM paths.

### Task 5.2 - Domain Registry As AVM Or Minimal Identity Registry

Implementation options:

Option A for testnet:

- minimal native identity registry if already wired and bounded;
- domain ownership separate from wallet account state;
- resolver records not stored in account state.

Option B preferred for AVM maturity:

- domain registry contract backed by ChunkMap.

Rules:

- owner is `AE...`;
- name is bounded UTF-8 string;
- records are bounded ChunkMap entries;
- storage rent applies.

Tests:

- register;
- transfer;
- set resolver;
- get owner;
- export/import;
- storage rent accounting.

Done:

- Domain feature does not pollute wallet state.

## Phase 6 - App Wiring And Invariants

### Task 6.1 - Runtime Invariants

Implementation:

- Register global app invariants for:
  - bank/module accounting;
  - rewards cap;
  - rent reserve/runway;
  - pool shares;
  - validator entry;
  - direct delegation rejection;
  - AVM receipt/queue/state root;
  - no native app asset modules.
- Add CLI/test command or CI target that runs them against app state.

Tests:

- registry includes every required invariant;
- each invariant has a failing fixture;
- default app state passes;
- post-core-flow app state passes.

Done:

- Testnet release can prove app-level invariant coverage.

### Task 6.2 - No Full Scans In Block Lifecycle

Implementation:

- Audit BeginBlock/EndBlock/PreBlock.
- Replace all-account/all-contract scans with:
  - bounded queues;
  - touched-key updates;
  - epoch snapshots;
  - explicit indexes.

Tests:

- static scan test for risky iteration paths;
- block lifecycle bounded-iteration test with large fixture;
- no query pagination bypass.

Done:

- Chain can scale beyond toy localnet.

## Phase 7 - Final Public Testnet Gate

### Task 7.1 - Testnet Release Candidate Checklist

A release candidate is not ready until all pass:

- `go test ./...`;
- `go vet ./...`;
- `buf lint`;
- binary build;
- version command;
- genesis validate;
- 4-node localnet;
- 5-node localnet;
- export/import restart;
- upgrade rehearsal;
- app invariants;
- AVM deploy/execute/get method smoke;
- pool staking smoke;
- slashing/reputation v1 smoke;
- release package build;
- Docker image build;
- validator docs coverage.

Done:

- The release artifact has enough evidence for a public testnet launch decision.

### Task 7.2 - Public Testnet Launch Artifacts

Create release artifacts:

- binary archives for Linux amd64/arm64 and Windows dev tooling;
- Docker image;
- checksums;
- genesis JSON;
- genesis checksum;
- seed nodes list;
- persistent peers list;
- RPC endpoint list;
- faucet instructions if enabled;
- validator guide;
- Cosmovisor guide;
- health guide;
- incident response guide.

Done:

- External validators can join from published artifacts only.

## Recommended Execution Order

1. CHAT A: stabilize binary/version/genesis/localnet/export-import CI.
2. CHAT D: create canonical `docs/VALIDATOR.md`, `docs/TESTNET.md`,
   `docs/COSMOVISOR.md`, Dockerfile, health docs.
3. CHAT B: finish pool-based staking, direct delegation rejection, reward policy,
   slashing, validator score v1.
4. CHAT C: finish AVM exit-code mapping, Chunk core, typed Codec, ChunkMap, get
   methods, examples.
5. CHAT E: remove stale native DEX/token/NFT launch docs and enforce future AVM
   standards boundary.
6. CHAT A: final app wiring, 4/5-node localnet, upgrade rehearsal, release
   workflow.
7. All chats: run full testnet release candidate checklist.

## Definition Of Done For V2

Aetra is V2 testnet-ready only when:

- a clean checkout can build one `aetrad` binary;
- genesis can be generated, validated, and published;
- 4-5 node localnet reaches height and has stable validators;
- export/import restart works;
- upgrade rehearsal works;
- normal staking uses the pool path;
- direct user validator delegation is not the default user path;
- reward/slashing/reputation v1 are understandable and tested;
- AVM can deploy, execute external/internal messages, charge gas, store state,
  emit receipts/events, return small exit codes, and run get methods;
- AVM persistent data uses Chunk/ChunkMap design or has a clearly documented
  migration path to it before public contract developer use;
- token/NFT/DEX are AVM standards/contracts, not native app asset modules;
- validator docs, testnet docs, Cosmovisor docs, Docker image, health checks,
  seed/peer templates, and release artifacts exist;
- CI gates enforce the above.
