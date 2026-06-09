param(
  [string]$Script = "tests\e2e\upgrade_rehearsal_smoke.ps1",
  [string]$Playbook = "docs\upgrade-playbook.md"
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
$playbookText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Playbook)

foreach ($term in @(
    'rehearsal-noop',
    'upgrade-info.json',
    'Wait-LocalnetHeight',
    'stop.ps1',
    'export-genesis.ps1',
    'validate-genesis',
    'NoInit'
  )) {
  Assert-Contains -Text $scriptText -Pattern ([regex]::Escape($term)) -Message "upgrade rehearsal smoke missing: $term"
}

foreach ($term in @(
    'Cosmovisor',
    'same home directory',
    'upgrade-info.json',
    'rehearsal-noop',
    'export the state after the upgrade'
  )) {
  Assert-Contains -Text $playbookText -Pattern ([regex]::Escape($term)) -Message "upgrade playbook missing: $term"
}

Write-Host "upgrade rehearsal smoke script test passed"
