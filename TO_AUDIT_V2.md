# TO_AUDIT_V2: Aetra L1 Layer Review And Readiness Plan

## 1. Purpose

This file is the working checklist for a complete Aetra L1 review pass.
It translates `architecture.md`, `UPDATE.md`, existing review reports, module
code, proto surfaces, app wiring, docs, and tests into a sequence of concrete
verification tasks.

The review is local-only and repository-only. It does not include third-party
systems, public endpoints, live networks, or any action outside this workspace.

Every task must produce:

- evidence inspected;
- checks performed;
- tests or commands run;
- result;
- remaining gap, if any.

Recommended status labels:

- `PASS`: evidence is implemented and tested.
- `PARTIAL`: some evidence exists, but one or more production gates are missing.
- `GAP`: required evidence was not found.
- `N/A`: not applicable for the current release boundary.

## 2. Source Material To Read First

Primary planning documents:

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

Existing review examples and gates:

- `TO_AUDIT.md`
- `docs/security/manual-audit-checklist.md`
- `docs/security/dex-audit-report.md`
- `docs/security/module-bank-movement-audit.md`
- `docs/security/prototype-audit-gate.md`
- `docs/security/refactor-audit-report.md`
- `docs/security/security-audit-pack.md`
- `docs/architecture/*.md`
- `tests/scripts/*.ps1`

Implementation surfaces:

- `app/`
- `cmd/`
- `proto/`
- `x/`
- `tests/`
- `observability/`
- `.github/`

## 3. Review Method

For every layer:

1. Map expected behavior from `architecture.md` and `UPDATE.md`.
2. Map current files, modules, keepers, params, genesis, proto, tests, and docs.
3. Check deterministic state behavior.
4. Check integer accounting and bounded math.
5. Check authority and signer rules.
6. Check genesis validation and export/import behavior.
7. Check query, event, CLI, gRPC, REST, and observability coverage.
8. Check unit, integration, localnet, and documentation tests.
9. Record status in `AUDIT_RESULTS_V2.md`.

## 4. Repository Inventory Tasks

### 4.1 File And Module Inventory

Checks:

- Count repository files by top-level directory.
- List all `x/` modules.
- Count Go, proto, Markdown, script, and CI files.
- Count test files under `app/`, `x/`, and `tests/`.
- Identify generated files and avoid treating generated gateway code as runtime logic.

Commands:

```powershell
rg --files
Get-ChildItem -Path x -Directory
Get-ChildItem -Path app,x,tests -Recurse -Filter *_test.go -File
```

Acceptance:

- Layer map is complete.
- Generated code is separated from hand-written logic.
- Results include current git commit and dirty worktree status.

### 4.2 Planning Document Alignment

Checks:

- Confirm whether `UPDATES.md` or `UPDATE.md` exists.
- Extract headings from `architecture.md`.
- Extract headings from `UPDATE.md`.
- Identify contradictions between old and new VM direction.
- Identify outdated documents that still describe removed runtime modules.

Acceptance:

- Missing or renamed planning files are called out.
- AVM-vs-CosmWasm release boundary is explicit.
- Deprecated docs are listed as documentation debt, not implementation truth.

## 5. Layer 1: Base Chain And App Wiring

Expected from architecture:

- Cosmos SDK + CometBFT base chain.
- Native denom `naet`, display denom `AET`.
- Module account permission matrix.
- Reserved system addresses.
- Genesis validation.
- Export/import stability.
- App startup with all active modules.

Checks:

- Review `app/app.go`, `app/modules.go`, `app/wiring`, `app/keeperwiring`,
  `app/accounts`, `app/genesisconfig`, `app/genesisvalidation`.
- Verify all active store keys are mounted.
- Verify active modules are ordered for preblock, begin, end, init, and export.
- Verify system addresses and blocked addresses.
- Verify module account permissions are minimal.
- Verify default genesis includes required active modules.

Tests:

```powershell
go test ./app/...
go test ./cmd/l1d/...
```

Acceptance:

- App boots in tests.
- Genesis validation rejects invalid defaults.
- Export/import tests exist for active modules.
- Module account permissions are tested.

## 6. Layer 2: Consensus, Finality, And Determinism

Expected from architecture:

- 100-300 validators over time.
- 5-8 second target block time.
- Normal finality 5-15 seconds.
- Degraded-but-healthy finality target below 120 seconds.
- No non-deterministic inputs in state transitions.
- No wall-clock, random, filesystem, network, or unordered map effects in
  consensus paths.

Checks:

- Review ABCI handlers, lifecycle hooks, proposal handling, params, and app hash
  related code.
- Search for time, random, goroutine, select, float, unsorted map iteration,
  external IO, panics, and must helpers in app/x hand-written runtime code.
- Classify each match as runtime, generated, test, doc, or deterministic helper.
- Verify deterministic sorting where maps or slices affect hashes, roots, or
  stored state.

Commands:

```powershell
rg -n "time\.Now\(|math/rand|rand\.|go func|select \{|float32|float64|for .*range .*map|range .*map\[" app x --glob "*.go" --glob "!*.pb.go" --glob "!*.pb.gw.go" --glob "!**/*_test.go"
rg -n "panic\(|Must[A-Z][A-Za-z0-9_]*\(" app x --glob "*.go" --glob "!*.pb.go" --glob "!*.pb.gw.go" --glob "!**/*_test.go"
```

Acceptance:

- Every consensus-sensitive match has a safe explanation or follow-up item.
- Generated gateway code is not counted as consensus path.
- Tests include deterministic app behavior and restart/export checks.

## 7. Layer 3: Staking, Validators, Delegation, And Pools

Expected from architecture and update plan:

- Validator set grows from 100-128 to 150-200 to 250-300.
- Moderate self-bond and validator entry requirements.
- Delegation or pool-based participation.
- Nomination pools with deterministic share accounting.
- Unbonding period around 14-21 days.
- Delegators inherit validator penalties.
- Pool operator cannot move user principal outside defined rules.

Implementation surfaces:

- `x/nominator-pool`
- `x/single-nominator-pool`
- `x/validator-election`
- `x/validator-registry`
- `x/validator-insurance`
- `x/delegator-protection`
- `x/pos`
- `app/params/nomination_pool_spec.go`
- `docs/architecture/pos-staking-correctness.md`
- `docs/architecture/nomination-pool-spec.md`

Checks:

- Verify pool deposits, shares, rewards, withdrawals, unbonding, and rounding.
- Verify validator selection and registry state are deterministic.
- Verify validator entry stake and pool split docs match code gates.
- Verify pool proofs and export/import behavior.
- Verify direct delegation policy if disabled in current architecture.

Tests:

```powershell
go test ./x/nominator-pool/...
go test ./x/single-nominator-pool/...
go test ./x/validator-election/...
go test ./x/validator-registry/...
powershell -ExecutionPolicy Bypass -File tests/scripts/nomination_pool_spec_doc_test.ps1
powershell -ExecutionPolicy Bypass -File tests/scripts/pos_staking_correctness_doc_test.ps1
```

Acceptance:

- Pool accounting conserves value.
- Withdrawal and unbonding behavior is tested.
- Validator entry policy is documented and test-gated.
- Dirty or in-progress pool changes are explicitly called out.

## 8. Layer 4: Anti-Concentration And Commission Policy

Expected from architecture:

- Effective power cap schedule:
  - up to 150 validators: 3.0 percent;
  - 151-250 validators: 2.5 percent;
  - above 250 validators: 2.0 percent.
- Overflow stake does not create extra influence.
- Overflow rewards are reduced or zero.
- Commission floor, maximum, and daily change bounds.
- Top-N concentration monitoring.
- Delegation warnings.

Implementation surfaces:

- `x/aetra-staking-policy`
- `x/stake-concentration`
- `x/dynamic-commission`
- `app/params/aetra_staking_policy_spec.go`
- `docs/architecture/staking-policy.md`

Checks:

- Verify cap math for 100, 150, 200, 250, 300 validators.
- Verify raw stake, effective stake, overflow stake, and top-N math.
- Verify commission bounds.
- Verify governance or authority checks.
- Verify whether cap affects rewards only or actual CometBFT voting power.
- Verify active app wiring and proto/API exposure.

Tests:

```powershell
go test ./x/aetra-staking-policy/...
go test ./x/stake-concentration/...
go test ./x/dynamic-commission/...
powershell -ExecutionPolicy Bypass -File tests/scripts/aetra_staking_policy_spec_doc_test.ps1
```

Acceptance:

- Cap calculations are deterministic and tested.
- Cap cannot exceed configured values.
- Top-N query is available or gap is recorded.
- If not wired to actual validator updates, status must be `PARTIAL`.

## 9. Layer 5: Penalties, Evidence, And Validator Accountability

Expected from architecture:

- Base Cosmos SDK slashing and evidence flow.
- Severe double-sign handling.
- Downtime handling.
- Progressive downtime only if implemented safely.
- Timestamp and proposal rules remain deterministic.
- No subjective penalty path.

Implementation surfaces:

- SDK `x/slashing` and `x/evidence` wiring.
- `x/evidence`
- `x/reporter`
- `x/performance`
- `x/reputation`
- `x/aetra-validator-score`
- `docs/architecture/slashing-system.md`

Checks:

- Verify evidence module wiring and lifecycle.
- Verify custom evidence state does not replace SDK evidence unsafely.
- Verify downtime and jail records.
- Verify score inputs are chain-state based.
- Verify no manual operator judgement changes consensus.
- Verify unbonding/evidence timing is documented and tested.

Tests:

```powershell
go test ./x/evidence/...
go test ./x/reporter/...
go test ./x/performance/...
go test ./x/reputation/...
go test ./x/aetra-validator-score/...
powershell -ExecutionPolicy Bypass -File tests/scripts/slashing_system_doc_test.ps1
powershell -ExecutionPolicy Bypass -File tests/scripts/aetra_validator_score_spec_doc_test.ps1
```

Acceptance:

- Evidence and score records validate deterministically.
- Score remains informational unless objective reward modifiers are explicitly enabled and tested.
- Progressive downtime status is clear: implemented, partial, or deferred.

## 10. Layer 6: Economics, Fees, Burn, Treasury, And Supply

Expected from architecture:

- Inflation in low/moderate bounds.
- Target bonded ratio around 55-65 percent.
- Delegator APR estimate around 5-8 percent.
- Fee split: burn, rewards, treasury.
- Burned amount is queryable.
- Treasury accounting is queryable.
- Supply invariants hold.

Implementation surfaces:

- `x/aetra-economics`
- `x/fees`
- `x/fee-collector`
- `x/burn`
- `x/treasury`
- `x/emissions`
- `x/mint-authority`
- `app/params/aetra_economics_spec.go`
- `docs/genesis-params.md`

Checks:

- Verify fee split sums to 10000 bps.
- Verify burn, reward, and treasury accounting.
- Verify mint caps and mint authority.
- Verify estimated APR labels are estimates.
- Verify dynamic inflation bounds and epoch smoothing.
- Verify active app wiring and proto/API exposure.

Tests:

```powershell
go test ./x/aetra-economics/...
go test ./x/fees/...
go test ./x/fee-collector/...
go test ./x/burn/...
go test ./x/treasury/...
go test ./x/emissions/...
go test ./x/mint-authority/...
powershell -ExecutionPolicy Bypass -File tests/scripts/aetra_economics_spec_doc_test.ps1
```

Acceptance:

- Accounting uses integers only.
- Supply cannot underflow or overflow in tested paths.
- Fee split and burn are tested.
- If economics model is standalone and not active app flow, status must be `PARTIAL`.

## 11. Layer 7: Native Accounts, Identity, Reputation, And Proofs

Expected from `UPDATE.md`:

- Native wallet/address policy.
- Virtual accounts.
- Activation and auth policy.
- Account state and storage rent.
- Domains and identity.
- Reputation and proof/event/receipt surfaces.

Implementation surfaces:

- `app/addressing`
- `app/accounts`
- `x/native-account`
- `x/identity`
- `x/identity-root`
- `x/reputation`
- `x/proofregistry`
- `x/events`
- `x/storage-rent`

Checks:

- Verify address codec and reserved address rules.
- Verify zero, system, and user address behavior.
- Verify identity record ordering and proof roots.
- Verify reputation scoring and query surfaces.
- Verify storage rent accounting and reserve rules.
- Verify docs and tests match native account staking direction.

Tests:

```powershell
go test ./app/addressing ./app/accounts
go test ./x/native-account/...
go test ./x/identity/...
go test ./x/identity-root/...
go test ./x/reputation/...
go test ./x/storage-rent/...
powershell -ExecutionPolicy Bypass -File tests/scripts/native_account_staking_doc_test.ps1
```

Acceptance:

- Address and account rules are test-gated.
- Identity/proof state is canonical.
- Storage rent cannot bypass base accounting.

## 12. Layer 8: VM And Contract Execution

Expected current direction:

- AVM is genesis VM.
- CosmWasm is not genesis default.
- CosmWasm is optional, gated, and disabled by default.
- Contract standards are AVM-native first.
- Gas, storage, async, events, receipts, export/import, and standards are tested.

Implementation surfaces:

- `x/aetravm/avm`
- `x/aetravm/async`
- `x/aetravm/standards/aft`
- `x/aetravm/standards/anft`
- `x/aetravm/standards/aw`
- `x/vm`
- `x/contracts`
- `x/avm-scheduler`
- `app/wasmconfig`
- `docs/architecture/vm-direction.md`

Checks:

- Verify AVM bytecode validation, gas metering, storage, async queue, receipts,
  and contract standards.
- Verify CosmWasm disabled-by-default policy.
- Verify optional CosmWasm limits and gate config.
- Verify contract state export/import and canonical roots.
- Verify events and receipts are stable.

Tests:

```powershell
go test ./app/wasmconfig
go test ./x/aetravm/...
go test ./x/vm/...
go test ./x/contracts/...
go test ./x/avm-scheduler/...
powershell -ExecutionPolicy Bypass -File tests/scripts/vm_direction_doc_test.ps1
```

Acceptance:

- AVM tests pass.
- CosmWasm remains gated.
- Any production execution gaps are listed separately from executable specs.

## 13. Layer 9: Routing, Messaging, Scheduler, Zones, And Sharding R&D

Expected:

- Deterministic routing.
- Explicit zones.
- Message queues and receipts.
- Scheduler and AVM scheduler.
- Sharding remains R&D until gates are complete.

Implementation surfaces:

- `x/routing`
- `x/messages`
- `x/messaging`
- `x/queue`
- `x/scheduler`
- `x/schedulerv2`
- `x/avm-scheduler`
- `x/zones`
- `x/sharding`
- `x/sharding-coordinator`
- `x/mesh`
- `x/networking`

Checks:

- Verify canonical ordering of routing tables, zones, queues, receipts, and roots.
- Verify scheduler limits and deterministic execution.
- Verify sharding code is not silently treated as mainnet-ready.
- Verify networking code is not inside consensus state transitions unless deterministic.

Tests:

```powershell
go test ./x/routing/...
go test ./x/messages/...
go test ./x/messaging/...
go test ./x/scheduler/...
go test ./x/zones/...
go test ./x/sharding/...
powershell -ExecutionPolicy Bypass -File tests/scripts/sharding_rd_doc_test.ps1
```

Acceptance:

- Queue and routing state is deterministic.
- Sharding status is clearly gated.
- Scheduler state has export/import or documented limits.

## 14. Layer 10: Governance And Parameter Safety

Expected:

- Governance controls critical params within bounds.
- Unsafe params are rejected.
- Critical params may activate at epoch boundary.
- Events are emitted for param changes.

Implementation surfaces:

- SDK governance wiring.
- `x/config`
- `x/config-voting`
- `x/constitution`
- `x/system-registry`
- `app/params/governance_parameters.go`

Checks:

- Verify parameter specs define type, default, min, max, authority, activation.
- Verify invalid updates are rejected.
- Verify events are emitted where implemented.
- Verify app/genesis validation includes governance parameter catalog.

Tests:

```powershell
go test ./x/config/...
go test ./x/config-voting/...
go test ./x/constitution/...
go test ./x/system-registry/...
powershell -ExecutionPolicy Bypass -File tests/scripts/governance_parameters_doc_test.ps1
```

Acceptance:

- Critical params cannot be set outside bounds.
- Parameter changes have test evidence.
- Missing event surfaces are listed.

## 15. Layer 11: API, CLI, Proto, Events, And Observability

Expected:

- CLI queries and tx commands.
- gRPC services.
- REST gateway where supported.
- Stable events.
- Prometheus metrics.
- Explorer/indexer-friendly surfaces.

Implementation surfaces:

- `proto/`
- `cmd/l1d`
- `app/api_services.go`
- `app/services`
- `x/*/client/cli`
- `x/indexer`
- `observability/`
- `docs/event-contract.md`

Checks:

- Count proto query and msg services.
- Verify REST annotations for modules that need them.
- Verify CLI command coverage.
- Verify event names and attributes are stable.
- Verify metrics include finality, block time, fees, burn, treasury, validator
  behavior, contract gas, and sync status.

Tests:

```powershell
go test ./cmd/l1d/...
go test ./observability/...
powershell -ExecutionPolicy Bypass -File tests/scripts/api_cli_query_event_surface_doc_test.ps1
powershell -ExecutionPolicy Bypass -File tests/scripts/observability_public_metrics_doc_test.ps1
```

Acceptance:

- Public query surface is documented.
- CLI examples exist for user-facing flows.
- Observability docs and tests pass.

## 16. Layer 12: Migration, Export/Import, And Upgrade Safety

Expected:

- Store key decision for every stateful module.
- Genesis import/export.
- Migration handlers for state-breaking changes.
- Version map updates.
- Upgrade tests.
- Operator instructions.

Implementation surfaces:

- `app/upgrades`
- `app/exporting`
- `x/*/module.go`
- `x/*/migrations`
- `docs/genesis-migrations.md`
- `docs/upgrade-migrations.md`

Checks:

- Verify migration registration.
- Verify modules with persistent state export/import correctly.
- Verify version map and app upgrades.
- Verify old genesis import tests where available.

Tests:

```powershell
go test ./app/exporting ./app/upgrades
powershell -ExecutionPolicy Bypass -File tests/scripts/data_migration_upgrade_strategy_doc_test.ps1
```

Acceptance:

- Every active persistent module has migration posture.
- Export/import errors are handled predictably.
- Missing migration tests are listed.

## 17. Layer 13: Test And CI Readiness

Expected:

- Unit tests.
- Integration tests.
- E2E/localnet smoke tests.
- Edge-case tests.
- Performance/load profiles.
- CI gate for critical subset.

Checks:

- Run `go test ./...`.
- Run selected doc gates for architecture-critical surfaces.
- Verify `.github/workflows`.
- Verify local PowerShell scripts remain usable.
- Verify Linux CI is primary for production confidence.

Commands:

```powershell
go test ./...
Get-ChildItem tests/scripts -Filter *.ps1 -File
```

Acceptance:

- Full Go test suite passes or failures are recorded.
- Critical doc gates pass or failures are recorded.
- Missing localnet/load tests are listed.

## 18. Result File Requirements

`AUDIT_RESULTS_V2.md` must include:

- date;
- git commit;
- dirty worktree list;
- files inspected;
- commands run;
- layer-by-layer status table;
- detailed findings;
- tests run and results;
- prioritized next work.

Findings must be actionable:

- evidence path;
- why it matters;
- required next step;
- suggested test.

Do not mark a layer production-ready only because unit tests pass. Production
readiness requires implementation, app wiring, params, genesis validation,
queries, events, docs, export/import, and test coverage.
