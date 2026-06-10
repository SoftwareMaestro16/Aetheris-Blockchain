param(
  [string]$Doc = "docs\parallel-codex-workstreams.md",
  [string]$Skills = "docs\cosmos-l1-skills.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) { return $Path }
  return Join-Path $RepoRoot $Path
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Doc)
$skillsText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Skills)

foreach ($term in @(
    'Parallel Codex Workstreams',
    'Each chat must pick exactly one workstream',
    'UPDATE.md',
    'architecture.md',
    'docs/cosmos-l1-skills.md',
    'Do not start by editing `app.go`',
    'Do not change address derivation',
    '`AE...` format',
    '`4:...` format',
    'sequence semantics',
    'signature domains',
    'Do not reintroduce user direct delegation to validators',
    'Do not add native token/NFT/DEX modules',
    'Every workstream must add tests',
    'export/import and genesis validation',
    'types`, `keys`, `keeper`, `messages`, `queries`,',
    'Avoid circular keeper dependencies',
    'Run targeted package tests first',
    'go test ./...',
    'Use one branch per workstream',
    'codex/native-account',
    'codex/storage-rent',
    'codex/liquid-staking-pool',
    'Do not commit unrelated dirty files'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "parallel workstreams doc missing rule: $term"
}

foreach ($term in @(
    'user-facing account/validator/consensus/pool address = AE...',
    'internal raw address = 4:...',
    'AE... <-> 4:... roundtrip must be stable',
    'active validators: 100-300 outside explicit testnet override',
    'minimum validator entry: 1_000_000 AET',
    'solo validator self-stake: 1_000_000 AET',
    'pool-backed validator self-stake: >= 400_000 AET',
    'pool-backed nominator/pool stake toward minimum entry: <= 600_000 AET',
    'direct user validator delegation: disabled',
    'User -> Liquid Staking Contract -> Pool Contract -> Validators',
    'normal user chooses pool/index, not a validator',
    'MinPoolDeposit = 10 AET',
    'UnbondingPeriod = 18 days',
    'MinTxFee = 0.003 AET = 3_000_000 naet',
    'StorageRentRate = 1 naet per byte-second',
    'storage_size = code_bytes + data_bytes',
    'effective_fee = gas_fee + storage_rent_delta + unpaid_storage_debt',
    'zero balance + no state = free',
    'zero balance + persistent state = debt + freeze, not delete',
    'system/critical state = protocol-paid + no freeze',
    'frozen = recoverable, state intact, balance intact',
    'archived = reduced state',
    'deleted = state removed'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "parallel workstreams doc missing shared contract: $term"
}

foreach ($term in @(
    'CHAT 1 - Repository Baseline And Guardrails',
    'W0 Address compatibility',
    'W1 Governance params schema',
    'W2 Native account/auth',
    'W3 Storage rent core',
    'CHAT 2 - Validator Registry And Official Pool Entry',
    'W4 Validator registry/policy',
    'W5 Liquid staking pool state',
    'W6 Contract capability hooks',
    'CHAT 3 - Allocation, Rewards, Reputation, Proofs',
    'W7 Allocation engine',
    'W8 Pool rewards',
    'W9 Stake reputation',
    'W10 Proofs/events',
    'CHAT 4 - Genesis, Invariants, Docs, Final Wiring',
    'W11 Genesis/migrations/export-import',
    'W12 Scalability/invariants',
    'W13 Docs/CLI/query surface',
    'W14 Final app wiring',
    'All independent groups can start at once',
    'final app wiring happens after feature package APIs stabilize',
    'Leave broad app wiring to W14 after W0-W13 APIs are stable'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "parallel workstreams doc missing dependency graph item: $term"
}

foreach ($term in @(
    'CHAT 4 goal: make the independently built pieces production-usable',
    'export/import, migrations, invariants, docs, query/CLI surface, and final app',
    'deterministic genesis/export/import',
    'versioned migrations',
    'scalability checks',
    'invariant registry',
    'docs and examples',
    'final keeper/app/module wiring',
    'full test pass',
    'CHAT 4 workstreams must not rewrite',
    'address derivation',
    'account auth semantics',
    'allocation math',
    'reward math',
    'storage rent semantics',
    'CHAT 4 owned workstreams'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "parallel workstreams doc missing CHAT4 item: $term"
}

foreach ($term in @(
    'W11 Genesis, Migration, Export/Import Requirements',
    'W11 ownership:',
    'genesis state for new modules',
    'export/import validation',
    'versioned migrations',
    'lazy migration',
    'W11 tasks:',
    'deterministic export/import for accounts, pools, allocations, rewards',
    'reputation, rent, and validator policy',
    'versioned account/pool migration',
    'Reject malformed duplicate state before writes',
    'Preserve mixed account versions',
    'W11 depends on W2/W3/W4/W5/W7/W8/W9 types',
    'W11 must not touch business logic except migration handlers',
    'Required W11 tests:',
    'full export/import preserves all new state',
    'duplicate account/pool/share/allocation rejected',
    'unsupported version rejected safely',
    'lazy migration preserves address and sequence'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "parallel workstreams doc missing W11 item: $term"
}

foreach ($term in @(
    'W12 Scalability And Invariants Requirements',
    'W12 ownership:',
    'invariant registration',
    'bounded-iteration tests',
    'benchmarks/simulations',
    'Assert no O(N users) BeginBlock/EndBlock paths',
    'bank/module accounting, rewards cap, rent, pool',
    'shares, validator entry, and direct delegation rejection',
    'million-user style simulation for pool shares and reward claims',
    'W12 depends on all core state modules',
    'W12 must not touch feature implementation except small instrumentation hooks',
    'Required W12 tests:',
    'BeginBlock/EndBlock bounded',
    'reward claim bounded',
    'reputation claim bounded',
    'rent charge bounded',
    'invariant failure fixtures'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "parallel workstreams doc missing W12 item: $term"
}

foreach ($term in @(
    'W13 Docs, CLI, Query Surface Requirements',
    'W13 ownership:',
    'docs',
    'CLI examples',
    'query docs',
    'static doc tests',
    'normal users stake only through official liquid staking',
    'Document validator entry:',
    '`1_000_000 AET`',
    'solo full self-stake',
    'pool-backed `400_000/600_000`',
    'Document `100-300` validator range',
    'Document unbonding `18 days`',
    'Document min tx fee `0.003 AET`',
    'Document storage rent and recoverable freeze/unfreeze',
    'W13 depends on final names from W1/W2/W5',
    'W13 must not touch keeper logic',
    'Required W13 tests:',
    'static doc tests for required terms',
    'command examples compile or are validated'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "parallel workstreams doc missing W13 item: $term"
}

foreach ($term in @(
    'W0 owns:',
    '`app/addressing`',
    'address validation tests',
    'address docs snippets',
    'Freeze `AE...` and `4:...` golden vectors',
    '`FormatPoolAddress`',
    '`ParsePoolAddress`',
    'account `AE...` roundtrip',
    'validator `AE...` roundtrip',
    'consensus `AE...` roundtrip',
    'pool `AE...` roundtrip',
    'raw `4:...` roundtrip',
    'malformed legacy prefixes rejected',
    'must not touch staking keeper logic',
    'storage rent accounting',
    'broad app module wiring'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "parallel workstreams doc missing W0 item: $term"
}

foreach ($term in @(
    'A Cosmos SDK L1 is a deterministic state machine',
    'Cross-module access uses explicit keeper interfaces',
    'Avoid global mutable state',
    'Floating-point math in consensus logic',
    'Missing genesis validation and missing migration path'
  )) {
  Assert-Contains -Text $skillsText -Pattern ([regex]::Escape($term)) -Message "cosmos skills doc missing prerequisite: $term"
}

Write-Host "parallel codex workstreams doc test passed"
