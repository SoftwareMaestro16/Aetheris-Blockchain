# Manual Security Audit Checklist

Use this checklist for every PR that changes consensus, funds, governance,
genesis, localnet, CI, release, or dependency behavior.

## Required Record

Every security-relevant PR must record:

- change scope and affected modules;
- reviewer name;
- linked test evidence or workflow run;
- list of findings with severity;
- triage decision for every `Critical`, `High`, and `Medium` finding;
- explicit statement that no untriaged `Critical` or `High` finding remains.

## Cosmos Consensus Review

- Address, signer, authority, admin, recipient, and module-account fields reject
  empty, malformed, zero, and unauthorized values.
- Native `naet` fee, staking, mint, and bank assumptions cannot drift through
  genesis, params, tx, migration, or export/import paths.
- Keeper writes are deterministic and do not depend on wall time, randomness,
  map iteration order, goroutine races, external APIs, local filesystem state,
  or platform-specific serialization.
- Malformed tx, query, params, and genesis inputs return errors instead of
  ABCI panics where the SDK interface allows error returns.
- Bank movements propagate SDK bank errors and do not leave partial custom
  module state.
- DEX reserves, module balances, and LP supply remain synchronized after every
  pool, liquidity, swap, export, and import path.
- Tokenfactory mint, burn, create denom, and admin transfer paths require the
  current admin or configured authority.
- Query/list endpoints use pagination, direct lookups, or explicit caps.
- Migrations validate exportable state before and after layout changes.
- Localnet and diagnostics scripts do not print or package keyrings, validator
  private keys, mnemonics, database URLs, API tokens, or environment dumps.

## Contract Standards Review

- AW-5 wallet commands reject replayed `seqno`, wrong `wallet_id`, expired
  `valid_until`, invalid signatures, unauthorized extensions, and non-`naet`
  protocol fee paths before state mutation.
- AFT-44 token master/wallet accounting rejects token supply divergence,
  non-admin mint/admin takeover, replayed wallet messages, malformed
  metadata, native AET/naet spoofing, and non-`naet` protocol fees.
- ANFT-66 item transfers require current owner authorization and reject
  metadata spoofing, malformed collection/item addresses, and unbounded batch
  minting.
- ASBT-67 transfer attempts are rejected, SBT revocation requires authority, and
  revoke paths do not transfer ownership.
- Async contract execution rejects queue DoS, malformed envelopes, duplicate
  contract addresses, duplicate queue sequences, queue `next_sequence` drift,
  and bounce/refund paths that could double-spend value.

## Automated Evidence

Attach or link successful runs for:

- `go test ./...`
- `go vet ./...`
- `buf lint`
- generated protobuf verification
- `govulncheck`
- `gosec`
- `gitleaks`
- CodeQL
- dependency review for dependency-changing PRs

## Manual Decision

Before merge, the reviewer must confirm:

```text
No untriaged Critical or High security findings remain.
```

If that statement is false, the PR must stay unmerged until the finding is
fixed, downgraded with rationale, or accepted with owner and tracked issue.
