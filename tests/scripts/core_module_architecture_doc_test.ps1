$ErrorActionPreference = "Stop"

$repo = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$doc = Join-Path $repo "docs\architecture\core-module-architecture.md"

if (-not (Test-Path $doc)) {
  throw "missing core module architecture doc"
}

$text = Get-Content -Raw $doc
$required = @(
  'x/auth',
  'x/bank',
  'x/staking',
  'x/slashing',
  'x/gov',
  'x/distribution',
  'app/addressing',
  'app/indexer',
  '4:',
  'AE...',
  'naet',
  'wrong chain-id',
  'identical signed tx bytes',
  'deterministic transfer events',
  'min commission',
  'max commission',
  'software upgrade'
)

foreach ($needle in $required) {
  if ($text -notlike "*$needle*") {
    throw "core module architecture doc missing: $needle"
  }
}

Write-Host "core module architecture doc test passed"
