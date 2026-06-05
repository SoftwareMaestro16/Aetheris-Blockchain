# Module Boundaries

## `x/tokenfactory`

Purpose: create and manage custom denoms without EVM dependency.

State:
- Denom registry keyed by full denom.
- Admin record per denom.
- Optional metadata record per denom.
- Module params.

Minimal Msg surface:
- `MsgCreateDenom`
- `MsgMint`
- `MsgBurn`
- `MsgChangeAdmin`
- `MsgUpdateParams`

Keeper dependencies:
- Bank keeper interface for mint, burn, send, and metadata operations.
- Account/address codec where required by the scaffolded SDK version.

Security invariants:
- Only authorized admins can mint, burn, or transfer admin rights.
- Total supply changes must match bank keeper mint/burn results.
- Subdenom length bounds and mint/burn/create emergency flags are governance-controlled.
- Governance cannot seize an existing factory denom admin role.

## `x/dex`

Purpose: deterministic constant-product AMM.

Current direction:
- `x/dex` remains a native module while async contract execution and VM
  selection mature.
- Future contract-based pools/routers must treat this module as the reference
  implementation or migration bridge until audited.

State:
- Pool registry keyed by pool ID.
- Asset pair index.
- LP share accounting.
- Fee accumulator references.
- Module params.

Minimal Msg surface:
- `MsgCreatePool`
- `MsgAddLiquidity`
- `MsgRemoveLiquidity`
- `MsgSwapExactIn`
- `MsgUpdateParams`

Keeper dependencies:
- Bank keeper interface for escrow, pool balances, and LP share movement.
- Fees keeper interface for protocol fee accounting.

Security invariants:
- Integer math only.
- No pool operation can create value.
- LP shares must remain backed by pool reserves.
- User-provided min-out values protect against slippage.
- Governance-controlled DEX params are bounded and cannot mutate reserves or LP supply.
- Duplicate pair lookup uses a deterministic pair index instead of scanning all pools.
- Native AET metadata cannot be spoofed through display denoms, factory denoms,
  or arbitrary LP denoms.
- Recorded reserves must match module account balances, and LP supply must
  match pool shares after every DEX state transition.

## `x/fees`

Purpose: centralize protocol fee policy and distribution.

State:
- Fee collector module account reference.
- Distribution weights.
- Accrued fee records where needed.
- Module params.

Minimal Msg surface:
- `MsgUpdateParams`
- Future fee claim/distribution messages only if they cannot be handled by hooks.

Keeper dependencies:
- Bank keeper interface for balances and transfers.
- Distribution or auth module interfaces only when explicitly required.

Security invariants:
- Distribution weights must sum to the configured denominator.
- Governance authority controls params.
- Fee collection must be idempotent for repeated block execution inputs.

## `x/aetherisvm/standards`

Purpose: define Aetheris-native contract standards before the VM runtime is
wired into the app. Standards are VM-independent executable specifications with
async/AVM-compatible conformance handlers.

State:
- No chain state.
- Standard packages are executable specifications, validation helpers, message
  codecs, and async conformance handlers only.

Current standards:
- `aft`: AFT-44 fungible token master/wallet contract model.
- `anft`: ANFT-66 NFT collection/item model and ASBT-67 soulbound extension.
- `aw`: AW-5 replay-safe contract wallet model.

Security invariants:
- Standards must not register SDK modules, stores, keepers, or protocol fee
  denoms by themselves.
- Standards must define explicit storage schema, inbound messages, outbound
  messages, getters, unknown-message policy, bounce behavior, fee behavior, and
  deployment behavior before runtime wiring.
- Standard `AsyncHandler` conformance paths must execute through the bounded
  async queue and must not bypass native-only fee validation.
- Native `AET`/`naet` is not a user token standard instance.
- User token balances cannot satisfy base-chain protocol fees.
- Contract address derivation and supply accounting rules must be deterministic
  and regression-tested before VM wiring.
- NFT collection/item membership and SBT non-transferability must be
  deterministic and regression-tested before VM wiring.
- Wallet `seqno`, `wallet_id`, `valid_until`, signature, extension auth, and
  native-only fee rules must fail before state mutation.

## `x/aetherisvm/async`

Purpose: define deterministic asynchronous contract message semantics before
keeper and VM runtime wiring.

State:
- Contract account model.
- Bounded contract state bytes.
- Message envelope model.
- Global message queue.
- Per-contract inbox/outbox views.
- Execution receipts and observability counters.

Current status:
- Pure Go executable specification only.
- No SDK stores, keepers, module accounts, or ABCI hooks are registered yet.
- Cosmos SDK delivered transactions remain synchronous; async semantics are
  modeled as deterministic queue processing inside blocks.
- Production partitioning or sharding is a later R&D track, not part of this
  module boundary.

Security invariants:
- Contract address derivation must be deterministic.
- Queue ordering must use tx index, message index, source logical time,
  destination key, and assigned sequence tie-breaker.
- Bounce and refund behavior must be deterministic and preserve failed-state
  rollback rules.
- Bounce/refund service messages must not create double-refund loops.
- All protocol fee and message value accounting is native `naet` only.
- Per-message gas limit and forward fee validation must be explicit.
- Per-tx, per-block, recursion, body, state, emitted-message, storage-write,
  and deploy limits must be bounded before VM wiring.
- Export/import must preserve queued messages, inbox, outbox, receipts, and
  metrics exactly, including `next_sequence`, `next_tx_index`, logical time,
  and ordering metadata.

## `x/aetherisvm/avm`

Purpose: define the native Aetheris Virtual Machine before keeper and runtime
wiring.

State:
- Deterministic bytecode format.
- Module verifier.
- Local runner.
- Storage snapshot ABI.
- Gas schedule.
- Host function allowlist.
- Async handler adapter.

Current status:
- Pure Go executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.
- AVM cannot mutate production chain state until the base-chain safety gate,
  async queue semantics, security scans, determinism gate, and adversarial
  audit are green.

Security invariants:
- Bytecode serialization must be deterministic.
- Code hash must be computed from encoded bytecode.
- Message ABI must use the async `MessageEnvelope`.
- Storage ABI must use deterministic key/value snapshots with bounded memory.
- Host functions must be allowlisted.
- Wall-clock time, host randomness, filesystem/network access, floating point,
  unbounded iteration, and nondeterministic map iteration are forbidden.
- Gas, code size, memory, stack/register, import, and instruction limits must be
  bounded before keeper wiring.
- AVM must not bypass address validation, zero-address rejection, `naet` fee
  policy, signer checks, malformed transaction handling, or genesis validation.

## `x/sharding/sim`

Purpose: provide the sharding R&D simulator before any production sharding or
partitioning implementation.

State:
- In-memory masterchain state model.
- In-memory workchain registry.
- In-memory shardchain registry.
- Cross-shard message and receipt model.
- Equivocation evidence model.

Current status:
- Pure Go simulator only.
- No SDK stores, keepers, module accounts, genesis, ABCI hooks, consensus
  changes, or network partitioning are registered.
- No production sharding claim is allowed.

Security invariants:
- The simulator must not register SDK stores or mutate production chain state.
- Public wording must say sharding R&D or experimental sharding until the
  production gate passes.
- Masterchain state must commit validator set, staking snapshot, workchain
  registry, shard headers, cross-shard receipt roots, config updates, and
  equivocation evidence.
- Workchains must keep explicit VM set, address format, genesis hash, upgrade
  policy, and native `naet` fee policy.
- Shardchains must commit state root, message queue root, receipt root,
  validator subset, data availability status, and split/merge references.
- Cross-shard messages must reject duplicate receipts, missing receipts,
  invalid shard proofs, stale shard headers, wrong destination shards, replayed
  messages, validator equivocation, and data-unavailable shard blocks.
- Prototype keepers may begin only after simulator tests, fuzz tests,
  adversarial tests, long-run testnet, independent audit, and
  consensus-safety proof are complete.

## `app/wasmconfig`

Purpose: keep CosmWasm readiness gated until explicit app wiring is requested.

State:
- No chain state.
- Policy constants and validation helpers only.

Current status:
- CosmWasm is the near-term gated VM candidate.
- CosmWasm remains disabled by default.
- Enabling CosmWasm requires explicit config or feature gate.
- Upload, instantiate, admin/migration, gas, contract size, memory/cache, and
  query limits are defined before keeper wiring.
- Pinned code is disabled by default and governance-only if enabled later.
- Governance authority for enabling/disabling CosmWasm is explicit and must be
  a non-zero authority address.

Security invariants:
- CosmWasm readiness must not add a `wasm` store key, module account, genesis
  state, CLI tx surface, or keeper wiring by default.
- CosmWasm cannot bypass `naet` fee policy, address policy, zero-address
  policy, or genesis validation.
- Upload, instantiate, execute, query, migrate, pinning, code size, gas, and
  query response/depth limits must be enforced before state mutation when the
  real `x/wasm` keeper is wired.
- A future Aetheris VM requires a written binary serialization spec, message
  ABI, storage ABI, gas schedule, deterministic execution proof, fuzz tests,
  upgrade/migration policy, and adversarial audit before implementation.
- Contract standards must remain testable independent of CosmWasm or a future
  Aetheris VM choice.

## `x/bridge`

Purpose: future interoperability. It remains out of scope for the first scaffold.

Activation requires a separate design covering light-client verification, replay domains, validator or relayer trust assumptions, finality, rate limits, and emergency controls.
