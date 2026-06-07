param(
  [string]$Doc = "docs\architecture\data-migration-upgrade-strategy.md"
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

foreach ($term in @(
    'Data Migration and Upgrade Strategy',
    'Aetra is expected to evolve. Upgrade safety is part of the architecture.',
    'module consensus versions are explicit',
    'migrations are registered before public testnet',
    'deterministic state transformations',
    'validates old state before writing new state',
    'validates new state after writing',
    'export/import after upgrade must produce valid genesis',
    'app hash after restart must remain stable',
    'dry-run upgrade procedure must be documented',
    'rollback limits must be documented',
    'unsafe downgrade behavior must be rejected',
    'upgrade handlers must emit stable events',
    'tests must cover migration success and failure paths',
    'current consensus version',
    'migration handlers from every supported previous version',
    'version map sanity checks',
    'genesis validation for migrated state',
    'export/import compatibility tests',
    'deterministic replay tests',
    'invariant tests after migration',
    'upgrade name',
    'target height',
    'binary version',
    'expected module version map before upgrade',
    'expected module version map after upgrade',
    'operator runbook',
    'post-upgrade smoke tests',
    'public rollback boundary',
    'missing module version rejected',
    'future module version rejected',
    'app hash stability after migration',
    'event emission for upgrade handler execution'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "data migration upgrade strategy doc missing: $term"
}

Write-Host "data migration upgrade strategy doc test passed"
