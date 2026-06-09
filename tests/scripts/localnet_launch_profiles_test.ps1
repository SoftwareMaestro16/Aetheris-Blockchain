param(
  [string]$Preflight = "scripts\testnet\public-testnet-preflight.ps1",
  [string]$Smoke = "tests\e2e\localnet_smoke.ps1",
  [string]$Diagnostics = "scripts\localnet\diagnostics.ps1",
  [string]$WaitHeight = "scripts\localnet\wait-height.ps1",
  [string]$ValidatorSet = "scripts\localnet\query-validator-set.ps1",
  [string]$ReadinessWorkflow = ".github\workflows\testnet-readiness.yml"
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

function Assert-NotContains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -match $Pattern) { throw $Message }
}

$preflightText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Preflight)
$smokeText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Smoke)
$diagnosticsText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Diagnostics)
$waitText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $WaitHeight)
$validatorSetText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $ValidatorSet)
$workflowText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $ReadinessWorkflow)

foreach ($term in @(
    'ValidateSet("3", "4", "5", "All")',
    '@(3, 4, 5)',
    'localnet_smoke.ps1',
    'diagnostics.ps1',
    'public-preflight-$validators',
    'aetra-local-1',
    'ValidatorCount $validators'
  )) {
  Assert-Contains -Text $preflightText -Pattern ([regex]::Escape($term)) -Message "preflight missing launch profile term: $term"
}

foreach ($term in @(
    'init.ps1',
    'validate-genesis.ps1',
    'start.ps1',
    'Wait-LocalnetHeight',
    'Wait-LocalnetValidators',
    'stop.ps1',
    'invalid validator count',
    'missing binary',
    'occupied RPC port'
  )) {
  Assert-Contains -Text $smokeText -Pattern ([regex]::Escape($term)) -Message "localnet smoke missing behavior: $term"
}

foreach ($term in @(
    'Wait-LocalnetHeight',
    'TargetHeight',
    'Assert-LocalnetWorkspacePath'
  )) {
  Assert-Contains -Text $waitText -Pattern ([regex]::Escape($term)) -Message "wait-height script missing: $term"
}

foreach ($term in @(
    'Wait-LocalnetValidators',
    'ExpectedCount',
    'validators?per_page=100',
    'voting_power'
  )) {
  Assert-Contains -Text $validatorSetText -Pattern ([regex]::Escape($term)) -Message "validator set script missing: $term"
}

foreach ($term in @(
    'priv_validator_key.json',
    'priv_validator_state.json',
    'node_key.json',
    'keyring data',
    'ConvertTo-LocalnetRedactedText'
  )) {
  Assert-Contains -Text $diagnosticsText -Pattern ([regex]::Escape($term)) -Message "diagnostics missing secret exclusion: $term"
}

foreach ($term in @(
    'localnet-smoke',
    'ValidatorCount 3',
    'localnet-rehearsal-4',
    'ValidatorProfile 4',
    'localnet-rehearsal-5',
    'ValidatorProfile 5',
    "github.event_name == 'workflow_dispatch'",
    'Upload diagnostics'
  )) {
  Assert-Contains -Text $workflowText -Pattern ([regex]::Escape($term)) -Message "readiness workflow missing launch profile gate: $term"
}

foreach ($bad in @(
    'ValidateSet("3", "5", "10", "All")',
    '@(3, 5, 10)',
    'ValidatorCounts 3,5,10',
    'prototype_acceptance.ps1',
    'cosmwasm_smoke.ps1'
  )) {
  Assert-NotContains -Text $preflightText -Pattern ([regex]::Escape($bad)) -Message "preflight still contains prototype-era localnet gate: $bad"
}

Write-Host "localnet launch profiles test passed"
