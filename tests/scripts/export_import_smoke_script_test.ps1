param(
  [string]$Script = "tests\e2e\export_import_smoke.ps1",
  [string]$Workflow = ".github\workflows\testnet-readiness.yml"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Path))
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$scriptText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Script)
$workflowText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Workflow)

foreach ($term in @(
    '$ImportDir',
    'Initialize-ImportedLocalnetHomes',
    'Copy-Item -LiteralPath $GenesisPath',
    'genesis", "validate-genesis"',
    'start.ps1',
    'NoInit',
    'Wait-LocalnetHeight',
    'Wait-LocalnetValidators',
    'latest_app_hash',
    'Get-LocalnetBankBalance',
    'node0-export-corrupt.json',
    'not-an-int'
  )) {
  Assert-Contains -Text $scriptText -Pattern ([regex]::Escape($term)) -Message "export/import smoke missing: $term"
}

foreach ($term in @(
    'export-import-roundtrip',
    'tests/e2e/export_import_smoke.ps1',
    'ValidatorCount 3'
  )) {
  Assert-Contains -Text $workflowText -Pattern ([regex]::Escape($term)) -Message "readiness workflow missing export/import term: $term"
}

Write-Host "export/import smoke script test passed"
