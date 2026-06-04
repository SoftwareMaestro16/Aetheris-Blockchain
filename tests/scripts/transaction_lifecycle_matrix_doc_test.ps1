param(
  [string]$Matrix = "docs\transaction-lifecycle-matrix.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$MatrixPath = if ([System.IO.Path]::IsPathRooted($Matrix)) { $Matrix } else { Join-Path $RepoRoot $Matrix }

function Assert-Contains {
  param(
    [string]$Text,
    [string]$Pattern,
    [string]$Message
  )
  if ($Text -notmatch $Pattern) {
    throw $Message
  }
}

$text = Get-Content -Raw -LiteralPath $MatrixPath

foreach ($row in @(
    'Bank send',
    'Staking delegate',
    'Tokenfactory create denom',
    'Tokenfactory mint',
    'Tokenfactory burn',
    'Tokenfactory change admin',
    'Fees update params',
    'DEX create pool',
    'DEX add liquidity',
    'DEX remove liquidity',
    'DEX swap exact in'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($row)) -Message "transaction lifecycle matrix missing row: $row"
}

foreach ($column in @(
    '\| Tx \| Actor \| Signer \| CLI input \| Funds and fee \| State writes \| Observable events \| Verification queries \| Tests \|',
    '\| Tx \| Signer/auth failures \| Field and denom failures \| Balance/state failures \| Replay/sequence coverage \| Scale/scan note \|'
  )) {
  Assert-Contains -Text $text -Pattern $column -Message "transaction lifecycle matrix missing required table shape: $column"
}

foreach ($securityTerm in @(
    'signer mismatch',
    'invalid Bech32',
    'invalid or spoofed denoms',
    'insufficient funds',
    'duplicate state',
    'replay/sequence failure',
    'ABCI panic',
    'No transaction path may require a full store scan'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($securityTerm)) -Message "transaction lifecycle matrix missing security term: $securityTerm"
}

foreach ($evidence in @(
    'x/fees/keeper/ante_test.go',
    'x/tokenfactory/keeper/msg_server_test.go',
    'x/dex/keeper/msg_server_test.go',
    'app/pos_test.go',
    'tests/e2e/tokenfactory_smoke.ps1',
    'tests/e2e/dex_smoke.ps1',
    'tests/e2e/pos_smoke.ps1',
    'tests/e2e/prototype_acceptance.ps1'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($evidence)) -Message "transaction lifecycle matrix missing evidence link: $evidence"
}

foreach ($gap in @(
    'MUST FIX before public release',
    'reusable signed-tx replay/sequence e2e helper',
    'SHOULD FIX for stronger operator observability',
    'Emit explicit domain events'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($gap)) -Message "transaction lifecycle matrix missing gap marker: $gap"
}

Write-Host "transaction lifecycle matrix doc test passed"
