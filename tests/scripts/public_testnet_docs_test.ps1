param(
  [string]$Preparation = "docs\public-testnet-preparation.md",
  [string]$Onboarding = "docs\validator-onboarding.md",
  [string]$Incident = "docs\testnet-incident-response.md"
)

$ErrorActionPreference = "Stop"

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

$prepText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Preparation)
$onboardingText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Onboarding)
$incidentText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Incident)

foreach ($term in @(
    "public-testnet-preflight.ps1",
    "ValidatorProfile All",
    "ValidatorProfile 10",
    "Faucet Plan",
    "Explorer And Indexer Plan",
    "Minimum Hardware",
    "Snapshot And State-Sync Plan",
    "CosmWasm Test Contract",
    "Validator Onboarding",
    "Incident response"
  )) {
  Assert-Contains -Text $prepText -Pattern ([regex]::Escape($term)) -Message "public testnet prep missing: $term"
}

foreach ($term in @(
    "genesis validate-genesis",
    "keys add",
    "keyring-backend os",
    "staking create-validator",
    "query staking validators",
    "cosmwasm_smoke.ps1"
  )) {
  Assert-Contains -Text $onboardingText -Pattern ([regex]::Escape($term)) -Message "validator onboarding missing: $term"
}

foreach ($term in @(
    "Severity",
    "Consensus Halt",
    "Suspected Fund Or Admin Exploit",
    "Faucet Incident",
    "Snapshot Or State-Sync Incident",
    "postmortem"
  )) {
  Assert-Contains -Text $incidentText -Pattern ([regex]::Escape($term)) -Message "incident response missing: $term"
}

Write-Host "public testnet docs test passed"
