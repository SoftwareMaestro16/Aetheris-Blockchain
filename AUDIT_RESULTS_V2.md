# AUDIT_RESULTS_V2: Aetra L1 Layer Review Results

## 1. Review Metadata

Date: 2026-06-08

Workspace: `C:\Users\Ryzen\Desktop\L1`

Git commit reviewed: `8d31c3b39`

Branch: `main`

Worktree status at review time:

```text
## main...origin/main
 M app/params/engineering_priorities.go
 M app/params/initial_release_non_goals.go
 M app/params/initial_release_non_goals_test.go
 M app/params/network_profile.go
 M app/params/network_profile_test.go
 M docs/architecture/engineering-priorities.md
 M docs/architecture/initial-release-non-goals.md
 M docs/module-boundaries.md
 M tests/scripts/engineering_priorities_doc_test.ps1
 M tests/scripts/initial_release_non_goals_doc_test.ps1
 M tests/scripts/native_account_staking_doc_test.ps1
 M x/nominator-pool/types/allocation_plan.go
 M x/nominator-pool/types/chat3_allocation_rewards_test.go
 M x/zones/types/zones_test.go
```

Note: these files were already modified before this review file was written.
They were not changed by this review pass.

Requested `UPDATES.md` status: file not found. The repository has `UPDATE.md`,
which was used as the update-plan source.

## 2. Files And Evidence Reviewed

Primary docs:

- `architecture.md`
- `UPDATE.md`
- `AVM.md`
- `BLOCKCHAIN.md`
- `COMPONENTS.md`
- `ECONOMICS.md`
- `NETWORKING.md`
- `POS.md`
- `README.md`
- `NEXT_ARCHITECTURE.md`

Review examples and gates:

- `TO_AUDIT.md`
- `docs/security/manual-audit-checklist.md`
- `docs/security/dex-audit-report.md`
- `docs/security/module-bank-movement-audit.md`
- `docs/security/prototype-audit-gate.md`
- `docs/security/refactor-audit-report.md`
- `docs/security/security-audit-pack.md`
- `tests/scripts/*.ps1`

Code surfaces:

- `app/`
- `cmd/`
- `proto/`
- `x/`
- `tests/`
- `observability/`

## 3. Repository Inventory

File count by top-level area from `rg --files`:

```text
.github: 5
app: 211
docs: 112
observability: 11
proto: 91
scripts: 32
tests: 97
x: 1048
```

Main extension count:

```text
.go: 1268
.ps1: 124
.md: 119
.proto: 89
.yaml: 1
.lock: 1
```

Test file count:

```text
app: 104
x: 369
tests: 2
total app/x/tests *_test.go: 475
```

`x/` module directories found:

```text
actor-registry, actors, aetracore, aetra-economics, aetra-staking-policy,
aetra-validator-score, aetravm, avm-scheduler, bridge-hub, burn, compute,
config, config-voting, constitution, contracts, cross-chain-registry,
delegator-protection, dynamic-commission, emissions, epoch, events, evidence,
execution, fee-collector, fees, identity, identity-root, indexer, internal,
load, market, memo, mesh, messages, messaging, mint-authority, native-account,
networking, nominator-pool, payments, performance, permissions, pos,
proofregistry, queue, reporter, reputation, routing, scheduler, schedulerv2,
services, sharding, sharding-coordinator, single-nominator-pool,
stake-concentration, storage, storage-rent, system-registry, taskgroups,
treasury, validator-economy, validator-election, validator-insurance,
validator-registry, vm, workflow, zones
```

## 4. Commands Run

Full Go test suite:

```powershell
go test ./...
```

Result: `PASS`

Duration observed: about 402.6 seconds.

Selected doc gates:

```powershell
powershell -ExecutionPolicy Bypass -File tests\scripts\engineering_priorities_doc_test.ps1
powershell -ExecutionPolicy Bypass -File tests\scripts\initial_release_non_goals_doc_test.ps1
powershell -ExecutionPolicy Bypass -File tests\scripts\native_account_staking_doc_test.ps1
powershell -ExecutionPolicy Bypass -File tests\scripts\aetra_staking_policy_spec_doc_test.ps1
powershell -ExecutionPolicy Bypass -File tests\scripts\aetra_economics_spec_doc_test.ps1
powershell -ExecutionPolicy Bypass -File tests\scripts\aetra_validator_score_spec_doc_test.ps1
powershell -ExecutionPolicy Bypass -File tests\scripts\vm_direction_doc_test.ps1
```

Result: `PASS` for all seven selected doc gates.

Static pattern sweeps:

```powershell
rg -n "time\.Now\(|math/rand|rand\.|go func|select \{|float32|float64|for .*range .*map|range .*map\[" app x --glob "*.go" --glob "!*.pb.go" --glob "!*.pb.gw.go" --glob "!**/*_test.go"
rg -n "panic\(|Must[A-Z][A-Za-z0-9_]*\(" app x --glob "*.go" --glob "!*.pb.go" --glob "!*.pb.gw.go" --glob "!**/*_test.go"
rg -n "TODO|not implemented|stub|placeholder" app x proto docs --glob "!*.pb.go" --glob "!*.pb.gw.go"
```

Observed counts:

```text
determinism-sensitive pattern matches in hand-written non-test app/x Go: 47
panic/Must matches in hand-written non-test app/x Go: 284
TODO/not implemented/stub/placeholder text matches outside generated pb/gw: 12
```

Important classification note:

- Generated gateway code produced many `go func` and `net/http` matches before
  generated files were excluded.
- The 47 deterministic-pattern matches are mostly map iteration validation
  helpers, roadmap/spec helpers, canonicalization logic, and a few keeper/type
  loops that require manual classification before mainnet readiness.
- The 284 `panic/Must` matches include normal Cosmos module registration and
  startup fail-fast paths, but also need classification before launch gates.

## 5. Layer Status Summary

| Layer | Status | Summary |
| --- | --- | --- |
| Base chain and app wiring | PASS/PARTIAL | App tests pass. Active module graph is broad and tested. Some architecture-target modules remain standalone or spec-level. |
| Consensus, finality, determinism | PARTIAL | Determinism test posture exists, but 47 static matches need manual classification and finality/load evidence is not complete. |
| Staking, validators, pools | PARTIAL | `x/nominator-pool`, `x/single-nominator-pool`, validator modules exist and test. Current dirty pool files must be reviewed separately. |
| Anti-concentration and commission | PARTIAL | `x/aetra-staking-policy`, `x/stake-concentration`, `x/dynamic-commission` exist and test. Main gap: actual validator update integration and public proto/API for `x/aetra-staking-policy`. |
| Evidence and validator accountability | PARTIAL | `x/evidence`, `x/reporter`, `x/performance`, `x/reputation`, `x/aetra-validator-score` exist and test. Progressive downtime as full production behavior needs app-level evidence. |
| Economics, fees, burn, treasury | PARTIAL | `x/aetra-economics`, `x/fees`, `x/fee-collector`, `x/burn`, `x/treasury`, `x/emissions`, `x/mint-authority` exist and test. Main gap: standalone `x/aetra-economics` not wired like active fee modules. |
| Native accounts, identity, proofs | PASS/PARTIAL | Addressing, native-account, identity, identity-root, reputation, storage-rent have strong tests. Full public workflow/localnet proof still needed. |
| VM and contracts | PASS/PARTIAL | AVM packages and VM direction gate pass. CosmWasm is documented as gated optional. Production AVM enablement still needs load/state-growth evidence. |
| Routing, scheduler, zones, sharding R&D | PARTIAL | Many executable specs and tests exist. Sharding remains R&D and should stay gated. |
| Governance and params | PASS/PARTIAL | Param spec docs and app/params tests are strong. Need full event and epoch-activation evidence per module. |
| API, CLI, proto, observability | PARTIAL | Proto and REST surfaces exist for many active modules. `x/aetra-*` modules lack proto service definitions. |
| Migration/export/import/upgrades | PARTIAL | Many modules register migrations and app export tests pass. Standalone `x/aetra-*` modules need production store migration posture if wired. |
| Tests and CI readiness | PASS/PARTIAL | `go test ./...` passes. Many doc gates exist. Full localnet/load/finality profile remains future gate. |

## 6. Detailed Findings

### F1. `architecture.md` conflicts with current VM direction

Status: `PARTIAL`

Evidence:

- `architecture.md` still describes `CosmWasm first`.
- `docs/architecture/vm-direction.md` states AVM is the genesis smart contract
  runtime and CosmWasm is optional and gated.
- User clarified that Aetra uses AVM for smart contracts.

Why it matters:

- The top-level architecture file can mislead future implementation work.
- Tests and docs now point to AVM-first, while the old top-level text still
  points to CosmWasm-first.

Required next step:

- Update `architecture.md` to make AVM the genesis VM.
- Move CosmWasm to optional gated compatibility.
- Keep explicit tests that CosmWasm remains disabled by default.

Suggested tests:

- `powershell -ExecutionPolicy Bypass -File tests\scripts\vm_direction_doc_test.ps1`
- `go test ./app/wasmconfig ./x/aetravm/... ./x/vm/...`

### F2. `UPDATES.md` requested by task does not exist

Status: `GAP`

Evidence:

- Repository contains `UPDATE.md`.
- No `UPDATES.md` was found in root file inventory.

Why it matters:

- Automation and future agents may read the wrong file name or skip update-plan
  context.

Required next step:

- Either rename `UPDATE.md` to `UPDATES.md`, add a small redirect file, or
  standardize all references on `UPDATE.md`.

Suggested test:

- Add a docs path test that verifies the chosen update-plan filename exists.

### F3. Three requested Aetra modules exist, but are not fully production-wired

Status: `PARTIAL`

Evidence:

- `x/aetra-staking-policy`
- `x/aetra-economics`
- `x/aetra-validator-score`
- All three have keeper/types logic and tests.
- Search found module names in their own packages and `app/params` specs, but
  not in active app store-key wiring, module order, persistent keeper wiring,
  or proto service files.

Why it matters:

- The logic is useful and tested, but it currently behaves like a standalone
  executable specification rather than a full Cosmos SDK runtime module.
- Architecture requires params, genesis validation, queries, events, tests,
  docs, and app wiring for production behavior.

Required next step:

- Decide whether each `x/aetra-*` module replaces, wraps, or remains a spec for
  existing active modules:
  - `x/aetra-staking-policy` vs `x/stake-concentration` and
    `x/dynamic-commission`;
  - `x/aetra-economics` vs `x/fees`, `x/fee-collector`, `x/burn`,
    `x/treasury`, `x/emissions`;
  - `x/aetra-validator-score` vs `x/reputation` and `x/performance`.
- If production modules, add proto, app module, store key, persistent keeper,
  genesis, migrations, query service, tx service, CLI, events, and integration
  tests.

Suggested tests:

- App startup test contains each module in `ModuleManager.Modules`.
- Store-key test contains each module store key.
- Default genesis includes each module if active.
- Export/import preserves each module state.

### F4. Effective validator power cap currently appears reward/spec-level

Status: `PARTIAL`

Evidence:

- `x/aetra-staking-policy/types/state.go` computes raw stake, effective stake,
  overflow stake, reward multiplier, warnings, top-10/top-20/top-33, and cap
  schedule.
- It does not appear wired to the Cosmos SDK staking keeper validator update
  flow that sends voting power updates to CometBFT.

Why it matters:

- Architecture explicitly requires a staged decision:
  - Stage 1: rewards/warnings only.
  - Stage 2: actual validator voting power.
- Current evidence supports Stage 1-style policy logic. Stage 2 requires deeper
  staking integration and heavy tests.

Required next step:

- Explicitly mark current behavior as Stage 1 or implement Stage 2.
- If Stage 2 is chosen, add staking keeper integration tests proving CometBFT
  validator updates use capped power while raw stake and penalty accounting
  remain correct.

Suggested tests:

- Validator with 5 percent raw stake and 3 percent cap produces 3 percent
  effective voting power.
- Delegation and unbonding shares stay correct.
- Penalty accounting uses raw stake where required.

### F5. Economics model exists, but active accounting path is split across modules

Status: `PARTIAL`

Evidence:

- `x/aetra-economics` has dynamic inflation, fee split, APR estimate, burned
  supply, treasury balance, epoch summaries, params validation, keeper tests.
- Active app wiring references `x/fees`, `x/fee-collector`, `x/burn`,
  `x/treasury`, `x/emissions`, and `x/mint-authority`.
- `docs/genesis-params.md` documents fee split defaults: 5000 burn, 3500
  rewards, 1500 treasury.

Why it matters:

- There are two layers: an Aetra-specific economics spec module and the active
  native fee/economics modules.
- Production readiness requires one authoritative accounting path.

Required next step:

- Define whether `x/aetra-economics` becomes active or remains a specification
  layer.
- Add cross-module tests that run fee collection, burn, treasury allocation,
  emissions, mint caps, and APR estimate through the active app path.

Suggested tests:

- Fee split sums to exactly 10000 bps.
- Burned fees reduce accounting supply according to active module rules.
- Treasury receives exact amount.
- Rewards are deterministic over many epochs.

### F6. Validator score model is deterministic but not app-integrated as primary accountability layer

Status: `PARTIAL`

Evidence:

- `x/aetra-validator-score` computes scores from chain-state-style inputs:
  uptime, missed blocks, jail events, slash events, self-bond, commission,
  governance participation, concentration, identity metadata.
- `ConsensusOverrideEnabled` defaults to false.
- The module is not active in app wiring from the searched evidence.

Why it matters:

- This matches the desired principle: score is informational first and reward
  modifiers only from objective data.
- Production accountability still needs integration with evidence, slashing,
  performance, reputation, rewards, and public queries.

Required next step:

- Keep consensus override disabled.
- Wire score output to explorer/query surface first.
- Only wire reward modifier after cross-module tests prove deterministic inputs.

Suggested tests:

- Score recomputation is stable after export/import.
- Reward modifier cannot exceed configured bounds.
- Score cannot override consensus participation.

### F7. Top-level architecture and generated architecture docs are not fully synchronized

Status: `PARTIAL`

Evidence:

- `architecture.md` is a long master spec.
- `docs/architecture/*.md` contains many more precise per-section gates.
- Several docs include deprecation notes for removed native asset-factory or
  native exchange modules.
- `architecture.md` text appears to have encoding damage in Russian sections.

Why it matters:

- Different docs may point agents toward different implementation directions.
- Encoding damage makes important requirements harder to review and quote.

Required next step:

- Normalize `architecture.md` encoding and update stale VM/economics/staking
  sections.
- Treat `docs/architecture/*.md` as section-specific gates when they are newer.
- Add a doc consistency script for master spec vs generated section docs.

Suggested tests:

- Doc test that verifies AVM-first language.
- Doc test that rejects reintroducing removed native asset-factory or native
  exchange modules as active runtime modules.

### F8. Static deterministic-pattern sweep needs manual classification

Status: `PARTIAL`

Evidence:

- 47 matches in hand-written non-test app/x Go after excluding generated pb/gw.
- Examples include map iteration in validation helpers and genesis/state helpers:
  - `app/params/validator_reputation.go`
  - `app/params/inflation_activity.go`
  - `app/params/adaptive_inflation.go`
  - `x/fee-collector/types/genesis.go`
  - `x/emissions/types/genesis.go`
  - `x/treasury/types/genesis.go`
  - `x/dynamic-commission/types/genesis.go`

Why it matters:

- Some map iterations only build error messages or validate docs. Those may be
  acceptable.
- Any map iteration that affects stored state, hashes, roots, genesis output,
  or app hash must be sorted or otherwise canonical.

Required next step:

- Add a deterministic-sweep allowlist with per-line justification.
- Fix or sort any runtime state-affecting map iteration.

Suggested tests:

- Deterministic re-run test for genesis validation and export/import.
- Canonical ordering tests for affected modules.

### F9. `panic/Must` usage is broad and should be classified

Status: `PARTIAL`

Evidence:

- 284 matches in hand-written non-test app/x Go after excluding generated pb/gw.
- Many are normal startup fail-fast paths in Cosmos module registration, genesis
  marshal/unmarshal, and migration registration.

Why it matters:

- Startup fail-fast can be acceptable.
- Runtime block processing or user-triggered paths should return errors instead
  of stopping the node process.

Required next step:

- Classify matches into:
  - startup-only;
  - test/helper;
  - exported pure helper with validated inputs;
  - runtime path needing error return.
- Convert any runtime user/state path to error-return behavior.

Suggested tests:

- Negative input tests for any converted runtime path.

### F10. Test posture is strong, but localnet/load/finality evidence remains a launch gate

Status: `PARTIAL`

Evidence:

- `go test ./...` passed.
- 475 Go test files exist under app/x/tests.
- 60+ PowerShell doc/test scripts exist.
- Architecture requires 100, 150-200, and 250-300 validator profiling, finality
  measurements, degraded validator scenarios, snapshot/state sync, and
  localnet reports.

Why it matters:

- Unit and integration tests cannot prove the 100-300 validator operational
  target alone.
- Finality target below 120 seconds needs measurement reports.

Required next step:

- Add or run public-testnet/localnet profile scripts for 100 validators first.
- Record block time, commit latency, missed blocks, state growth, snapshot time,
  restart time, and memory use.

Suggested tests:

- 100-validator localnet smoke.
- Finality measurement script.
- Partial validator offline scenario with healthy majority.

### F11. Documentation and code have strong doc-gate culture

Status: `PASS/PARTIAL`

Evidence:

- Selected doc gates passed:
  - engineering priorities;
  - initial release non-goals;
  - native account staking;
  - Aetra staking policy spec;
  - Aetra economics spec;
  - Aetra validator score spec;
  - VM direction.
- Many more scripts exist under `tests/scripts`.

Why it matters:

- This is a good structure for a large L1 project.
- The remaining work is to ensure doc gates are linked to active implementation,
  not only specs.

Required next step:

- Add CI grouping:
  - fast unit;
  - app wiring;
  - doc gates;
  - integration;
  - localnet/profile.

Suggested tests:

- CI job that runs selected critical doc gates on every push.

## 7. Per-Layer Notes

### Base Chain And App Wiring

Result: `PASS/PARTIAL`

Positive evidence:

- `go test ./app/...` passed as part of `go test ./...`.
- App wiring includes many active native modules.
- Store keys, module order, genesis, account permissions, and runtime services
  have tests.

Gaps:

- `x/aetra-*` modules are not active app modules from searched evidence.
- Some architecture-target features are represented in `app/params` as
  spec/report gates rather than runtime behavior.

### Consensus And Determinism

Result: `PARTIAL`

Positive evidence:

- Deterministic canonicalization appears in many type packages.
- Static generated-code noise was separated from hand-written runtime code.

Gaps:

- 47 static matches need line-by-line classification.
- No completed 100-300 validator finality report was found during this pass.

### Staking And Pools

Result: `PARTIAL`

Positive evidence:

- `x/nominator-pool` has keeper and type tests.
- `x/single-nominator-pool` has keeper coverage.
- `app/params/nomination_pool_spec.go` exists and is tested.

Gaps:

- Dirty `x/nominator-pool` files require separate review before relying on pool
  results.
- Need end-to-end staking flow through active app and localnet.

### Anti-Concentration

Result: `PARTIAL`

Positive evidence:

- `x/aetra-staking-policy` implements cap schedule, overflow stake, reward
  multiplier, warnings, top-N concentration, identity metadata, warning
  acknowledgement, params validation, genesis marshal/unmarshal, and tests.
- `x/stake-concentration` and `x/dynamic-commission` are active-style modules
  with proto surfaces.

Gaps:

- Actual CometBFT voting power cap integration was not found.
- `x/aetra-staking-policy` lacks proto/gRPC/app wiring from searched evidence.

### Economics

Result: `PARTIAL`

Positive evidence:

- `x/aetra-economics` implements the target economic model as code.
- Active fee/burn/treasury/emissions/mint-authority modules exist and test.
- Fee split defaults match architecture: 50 percent burn, 35 percent rewards,
  15 percent treasury.

Gaps:

- Need one authoritative production path for economic accounting.
- Need cross-module supply invariant evidence through the active app path.

### Validator Accountability

Result: `PARTIAL`

Positive evidence:

- `x/aetra-validator-score` has deterministic score logic and tests.
- Consensus override is disabled by default.
- Reputation/performance/reporter modules exist.

Gaps:

- Score module is standalone from searched evidence.
- Progressive downtime as full app behavior needs more evidence.

### VM And Contracts

Result: `PASS/PARTIAL`

Positive evidence:

- `docs/architecture/vm-direction.md` is clear: AVM is genesis VM.
- `go test ./x/aetravm/...` passed via full suite.
- `vm_direction_doc_test.ps1` passed.
- CosmWasm is gated and disabled by default per docs and `app/wasmconfig` tests.

Gaps:

- Top-level `architecture.md` needs AVM-first update.
- Production AVM still needs state-growth/load/finality evidence.

### Routing, Messaging, Zones, Sharding

Result: `PARTIAL`

Positive evidence:

- Routing, scheduler, zones, sharding coordinator, networking, and AVM scheduler
  packages exist and test.
- Sharding is documented as R&D, not default production behavior.

Gaps:

- Need clear public-testnet gate for sharding-related code.
- Need load and state growth evidence for scheduler/zone paths.

### Governance And Params

Result: `PASS/PARTIAL`

Positive evidence:

- `app/params` contains many architecture spec gates and tests.
- Governance parameter docs include bounded values.

Gaps:

- Need full event-surface proof for all critical param changes.
- Need epoch-delayed activation evidence for critical params where required.

### API, Proto, CLI, Events, Observability

Result: `PARTIAL`

Positive evidence:

- `proto/l1` contains many Query and Msg services with REST annotations for
  several modules.
- `cmd/l1d` tests passed.
- `observability` tests passed.

Gaps:

- `x/aetra-staking-policy`, `x/aetra-economics`, and
  `x/aetra-validator-score` do not appear to have proto service surfaces.
- Event matrices should be checked against active module events.

### Migration And Export/Import

Result: `PARTIAL`

Positive evidence:

- Many active modules register migrations.
- App export tests passed in the full Go test suite.

Gaps:

- Standalone Aetra modules only have JSON marshal/unmarshal helpers at this
  stage; if promoted to active modules, they need store migrations.

## 8. Priority Work List

P0:

1. Update `architecture.md` to AVM-first and CosmWasm-gated.
2. Resolve `UPDATE.md` vs `UPDATES.md` naming.
3. Decide production status of `x/aetra-staking-policy`,
   `x/aetra-economics`, and `x/aetra-validator-score`.
4. Classify 47 deterministic-pattern matches.
5. Classify 284 `panic/Must` matches.

P1:

1. If `x/aetra-*` modules become production modules, add proto, app wiring,
   store keys, persistent keepers, migrations, CLI, events, and integration
   tests.
2. Add cross-module economics accounting tests through active app modules.
3. Add Stage 1/Stage 2 decision doc for validator power cap.
4. Add score integration plan with consensus override disabled.

P2:

1. Add 100-validator localnet profile and finality report.
2. Add snapshot/state-sync/restart report.
3. Add AVM state-growth and gas profile report.
4. Add event matrix verification for explorers/indexers.

## 9. Final Conclusion

Aetra has a large and unusually broad L1 codebase with strong unit-test and
doc-gate culture. The full Go test suite passes, and the AVM direction is
already reflected in newer architecture docs and tests.

The main readiness gap is not compilation or simple unit coverage. The main gap
is production integration: several core Aetra-specific modules implement the
desired model as standalone keeper/types logic, while the active app graph uses
other native modules for fees, burn, treasury, emissions, reputation,
performance, stake concentration, and commission behavior.

Before public testnet, the project needs a clear decision for each such module:
promote to active runtime module, merge into existing active modules, or keep as
spec-only validation. After that decision, each production path needs app
wiring, proto/API, events, export/import, migration posture, integration tests,
and localnet/finality evidence.
