$ErrorActionPreference = "Stop"

$repo = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$doc = Join-Path $repo "docs\architecture\economy-interop-architecture.md"

if (-not (Test-Path $doc)) {
  throw "missing economy/interop architecture doc"
}

$text = Get-Content -Raw $doc
$required = @(
  'x/fees',
  'x/token',
  'x/ibc',
  'x/bridge',
  'dynamic fee calculation',
  'current network load',
  'estimated fee',
  'MsgUpdateParams',
  'naet',
  'AET',
  'mint accumulator',
  'burn accumulator',
  'IBC assets cannot pay Aetra protocol fees',
  'bridge assets cannot become fee denoms',
  'emergency pause',
  'proof verification tests'
)

foreach ($needle in $required) {
  if ($text -notlike "*$needle*") {
    throw "economy/interop architecture doc missing: $needle"
  }
}

Write-Host "economy/interop architecture doc test passed"
