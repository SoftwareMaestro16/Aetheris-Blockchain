param(
  [string]$Spec = "docs\architecture\sharding-rd.md",
  [string]$Boundaries = "docs\module-boundaries.md",
  [string]$Matrix = "docs\test-matrix.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$SpecPath = if ([System.IO.Path]::IsPathRooted($Spec)) { $Spec } else { Join-Path $RepoRoot $Spec }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }
$MatrixPath = if ([System.IO.Path]::IsPathRooted($Matrix)) { $Matrix } else { Join-Path $RepoRoot $Matrix }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$specText = Get-Content -Raw -LiteralPath $SpecPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath
$matrixText = Get-Content -Raw -LiteralPath $MatrixPath

foreach ($term in @(
    'Production Sharding And Partitioning R&D',
    'must not claim production sharding',
    'masterchain',
    'workchain',
    'shardchain',
    'cross-shard message',
    'Finality model',
    'Validator assignment',
    'Randomness source',
    'Shard header commitments',
    'Data availability',
    'Fraud/equivocation evidence',
    'Slashing rules',
    'Cross-shard message ordering',
    'Cross-shard replay protection',
    'Masterchain State',
    'Workchain Model',
    'Shardchain Model',
    'Simulator Before Implementation',
    'x/sharding/sim',
    'Prototype Only After Simulator',
    'duplicate cross-shard receipt',
    'missing receipt',
    'invalid shard proof',
    'stale shard header',
    'wrong destination shard',
    'replayed message',
    'validator equivocation',
    'data unavailable shard block',
    'consensus-safety proof',
    'No production sharding claim'
  )) {
  Assert-Contains -Text $specText -Pattern ([regex]::Escape($term)) -Message "sharding R&D spec missing: $term"
}

foreach ($term in @(
    'x/sharding/sim',
    'sharding R&D simulator',
    'No production sharding claim',
    'must not register SDK stores'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing sharding term: $term"
}

foreach ($term in @(
    'x/sharding/sim',
    'sharding-rd.md',
    'BenchmarkRoutingTableLookup',
    'BenchmarkCrossShardProofVerification',
    'BenchmarkShardSplitMerge',
    'BenchmarkShardedStateExportImport'
  )) {
  Assert-Contains -Text $matrixText -Pattern ([regex]::Escape($term)) -Message "test matrix missing sharding term: $term"
}

Write-Host "sharding R&D doc test passed"
