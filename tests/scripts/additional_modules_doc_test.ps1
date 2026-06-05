param(
  [string]$Doc = "docs\architecture\additional-modules.md",
  [string]$Boundaries = "docs\module-boundaries.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath

foreach ($term in @(
    'Additional Modules',
    'x/compute',
    'measure CPU/compute usage separately from simple tx gas',
    'price expensive computation',
    'protect validators from CPU abuse',
    'compute unit schedule',
    'per-op cost',
    'per-contract compute stats',
    'per-block compute budget',
    'expensive contract charged more',
    'compute cap enforced',
    'compute accounting deterministic',
    'x/permissions',
    'ACL system for contracts and modules',
    'resolver delegates',
    'domain managers',
    'contract extension permissions',
    'governance-controlled permissions',
    'all permissions have owner, scope, expiry, and revocation path',
    'permission checks are deterministic',
    'no hidden superuser outside governance/emergency policy',
    'x/indexer',
    'fast query layer',
    'state search',
    'event search',
    'memo search',
    'domain lookup',
    'token/NFT discovery',
    'indexer must never be required for consensus',
    'ConsensusRequired() returns false',
    'x/market',
    'market for compute, storage, and execution priority',
    'bounded, deterministic, and non-extractive',
    'cannot replace base `naet` fee',
    'cannot let wealthy users fully starve normal users',
    'must be capped by scheduler fairness',
    'normal-user',
    'reserved slots',
    'per-account share caps',
    'x/scheduler-v2',
    'DAG execution engine',
    'parallel tx scheduling',
    'async actor mailbox planning',
    'deterministic read/write set',
    'deterministic conflict resolution',
    'replayable schedule',
    'identical result across validators',
    'replayable DAG planner'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "additional modules doc missing: $term"
}

foreach ($term in @(
    '`x/compute`',
    'Compute unit schedule',
    'Per-op cost',
    'Per-contract compute stats',
    'Per-block compute budget',
    '`x/permissions`',
    'All permissions have owner, scope, expiry, and revocation path',
    'There is no hidden superuser outside explicit governance/emergency policy',
    '`x/indexer`',
    'Indexer must never be required for consensus',
    'Query limits are required for bounded result sets',
    '`x/market`',
    'Market premiums cannot replace the base `naet` fee',
    'Premiums are capped and priority score is capped by scheduler fairness',
    '`x/scheduler-v2`',
    'DAG execution engine',
    'Read/write sets are required and deterministic',
    'Schedule replay hash is stable across input order'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing additional module: $term"
}

Write-Host "additional modules doc test passed"
