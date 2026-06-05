# Public Testnet And Production Gates

This file is the release gate ledger for Aetheris public testnet and later
production readiness. It is stricter than the prototype acceptance suite and
does not replace module-specific security checklists.

## Public Testnet Gate

Public testnet cannot open until all required items are green or explicitly
triaged with owner, severity, mitigation, and target milestone.

Required checks:

- `go test ./...` passes.
- `go vet ./...` passes.
- `buf lint` passes.
- Security scans pass or findings are triaged:
  - `govulncheck`
  - `gosec`
  - CodeQL
  - gitleaks
  - dependency review
- Deterministic execution gate passes:
  - `scripts\security\determinism-gate.ps1`
- 3-validator and 5-validator localnet profiles pass:
  - `scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 3`
  - `scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 5`
  - `scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile All`
- Snapshot and state-sync work from published trust height, trust hash, and at
  least two RPC servers.
- Validator onboarding docs are clean and tested from a fresh machine or clean
  home directory.
- Faucet plan is implemented or explicitly deferred.
- Explorer/indexer plan is implemented or explicitly deferred.
- Incident response and rollback docs are tested.
- CosmWasm smoke passes if CosmWasm is enabled:
  - `tests\e2e\cosmwasm_smoke.ps1 -EnableWasm`
- AVM smoke passes if AVM is enabled:
  - `go test ./x/aetherisvm/avm ./x/aetherisvm/async`
- Contract standard smoke passes if async contracts are enabled:
  - `go test ./x/aetherisvm/standards/...`

Blocking rule:

- Any untriaged `Critical` or `High` fund-safety, consensus-safety, or
  secret-leak finding blocks public testnet.

## Production Gate

Production cannot be claimed until all public testnet gates remain green over
a long-running public testnet and the additional production requirements are
met.

Required production evidence:

- Long-running public testnet has no untriaged consensus-safety or fund-safety
  issues.
- Validator set can upgrade safely.
- Staking, fees, DEX, AVM, and contract standards have adversarial tests.
- State export/import is deterministic.
- Independent audit is completed and high/critical findings are fixed or
  explicitly accepted by governance with public rationale.
- Emergency governance and halt/restart process is tested.
- Snapshot/state-sync restore produces the same expected app hash.
- Public RPC, explorer/indexer, faucet, validator docs, incident response, and
  rollback process have owners and operational runbooks.

Production exclusions:

- Sharding remains `sharding R&D` or `experimental sharding` until
  `docs/architecture/sharding-rd.md` production gate is complete.
- CosmWasm remains disabled unless explicitly enabled by config and gate tests.
- AVM remains non-production until keeper wiring, adversarial tests, fuzz
  tests, export/import, and audit gates are complete.

## Immediate Build Order

1. Finish base-chain safety and Phase 2 helper cleanup.
2. Finish PoS/staking production hardening.
3. Build deterministic async queue without AVM first.
4. Build minimal AVM with a counter contract.
5. Implement AW-5 wallet.
6. Implement AFT-44 token master/wallet.
7. Implement ANFT-66 NFT collection/item.
8. Implement ASBT-67 soulbound item.
9. Gate CosmWasm behind explicit config and tests.
10. Start sharding simulator and spec.
11. Only after simulator and audit, prototype masterchain/workchain/shardchain.

## Evidence Map

| Gate Area | Evidence |
| --- | --- |
| Base chain tests | `go test ./...`, `go vet ./...`, `buf lint` |
| Security scans | `docs/security/security-audit-pack.md`, `.github/workflows/security.yml` |
| Determinism | `scripts\security\determinism-gate.ps1`, `docs/security/prototype-audit-gate.md` |
| Localnet 3/5 profiles | `scripts\testnet\public-testnet-preflight.ps1` |
| Snapshot/state-sync | `docs/public-testnet-preparation.md`, `scripts\localnet\snapshot.ps1`, `scripts\localnet\statesync.ps1` |
| Validator onboarding | `docs/validator-onboarding.md` |
| Faucet | `docs/public-testnet-preparation.md#faucet-plan` |
| Explorer/indexer | `docs/public-testnet-preparation.md#explorer-and-indexer-plan` |
| Incident/rollback | `docs/testnet-incident-response.md`, `docs/public-testnet-preparation.md#rollback-and-restart-procedure` |
| CosmWasm | `app/wasmconfig`, `tests\e2e\cosmwasm_smoke.ps1`, `docs/security/cosmwasm-readiness.md` |
| AVM | `x/aetherisvm/avm`, `docs/architecture/avm.md` |
| Contract standards | `x/aetherisvm/standards/...`, `docs/standards` |
| Sharding R&D | `x/sharding/sim`, `docs/architecture/sharding-rd.md` |

## Decision Record

Before public testnet launch, operators must publish:

- release commit,
- binary checksum,
- genesis hash,
- seed/persistent peers,
- chain id,
- expected native denom `naet`,
- public RPC endpoints,
- snapshot/state-sync trust data,
- faucet status: implemented or explicitly deferred,
- explorer/indexer status: implemented or explicitly deferred,
- enabled experimental features, if any.
