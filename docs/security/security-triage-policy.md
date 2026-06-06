# Security Triage Policy

This policy defines how Aetra handles scanner, dependency, CodeQL, secret,
and manual audit findings.

## Merge Rule

PRs must not merge while any `Critical` or `High` security finding is
untriaged.

A `Critical` or `High` finding is triaged only when it has one of:

- a merged fix with regression test or scanner evidence;
- an explicit downgrade with technical rationale, owner, and review date;
- an accepted risk record with owner, issue link, mitigation, and target date.

`Medium` findings require an owner and target milestone. `Low` findings may be
batched but must remain visible.

## Severity Sources

Use the highest severity reported by any source:

- GitHub CodeQL alert severity;
- GitHub Dependency Review advisory severity;
- `govulncheck`/OSV advisory severity;
- `gosec` severity;
- `gitleaks` confirmed secret exposure;
- manual Cosmos audit severity from
  [manual-audit-checklist.md](manual-audit-checklist.md).

When a tool lacks severity, classify reachable fund-safety, signer/authority,
consensus determinism, ABCI panic, secret exposure, and native-denom bypass
findings as `High` until reviewed.

## Required CI Gates

The Security Gate workflow must be required for protected branches:

- `govulncheck`
- `gosec high severity`
- `gitleaks secrets`
- `dependency review`
- `CodeQL`

GitHub branch protection must require these checks before PR merge. Direct
pushes to `main` are reserved for repository-owner emergency maintenance and
must still leave the tree passing the same commands locally or in CI.

`govulncheck` findings that are already documented and owner-triaged are listed
in `.github/security/govulncheck-triage.txt`. Any new `GO-*` advisory ID emitted
by the workflow is untriaged by default and fails the PR until this policy's
finding record is completed or the dependency is fixed.

## Finding Record Format

Record triage in an issue, release report, or security gate document using:

| Field | Required |
| --- | --- |
| Finding | Tool ID, advisory ID, file/line, or manual finding title |
| Severity | Critical, High, Medium, or Low |
| Reachability | Reachable, package-only, generated-code-only, local-only, or unknown |
| Decision | Fix, downgrade, accepted risk, duplicate, or false positive |
| Rationale | Concrete technical reason |
| Owner | Person or team |
| Target | Date, milestone, or linked issue |
| Evidence | Test, scanner output, commit, or mitigation doc |

## Blockers

These findings are never silently accepted:

- consensus-state nondeterminism;
- unauthorized mint, burn, admin, governance, signer, or authority action;
- wrong fee or staking denom accepted outside explicit tests;
- bank supply, DEX reserve, LP supply, or module balance corruption;
- ABCI panic from malformed public tx/query/genesis input where an error return
  is possible;
- committed secret, mnemonic, private key, token, or validator private material;
- dependency advisory with confirmed reachable `Critical` or `High` path.

## Public Testnet Gate

Public testnet cannot proceed while any untriaged `Critical` or `High`
fund-safety, consensus-safety, or secret-leak finding remains. This applies to
manual review findings and to `govulncheck`, `gosec`, CodeQL, gitleaks,
Dependency Review, determinism-gate, and prototype-audit outputs.
