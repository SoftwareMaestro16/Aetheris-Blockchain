param(
  [string]$GenesisDoc = "docs\genesis-migrations.md",
  [string]$UpgradeDoc = "docs\upgrade-migrations.md",
  [string]$AsyncDoc = "docs\architecture\async-smart-contract-execution.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$GenesisPath = if ([System.IO.Path]::IsPathRooted($GenesisDoc)) { $GenesisDoc } else { Join-Path $RepoRoot $GenesisDoc }
$UpgradePath = if ([System.IO.Path]::IsPathRooted($UpgradeDoc)) { $UpgradeDoc } else { Join-Path $RepoRoot $UpgradeDoc }
$AsyncPath = if ([System.IO.Path]::IsPathRooted($AsyncDoc)) { $AsyncDoc } else { Join-Path $RepoRoot $AsyncDoc }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$genesisText = Get-Content -Raw -LiteralPath $GenesisPath
$upgradeText = Get-Content -Raw -LiteralPath $UpgradePath
$asyncText = Get-Content -Raw -LiteralPath $AsyncPath

foreach ($term in @(
    'Genesis Export And Migration Contract',
    'DefaultGenesis -> InitChain/InitGenesis -> ExportAppStateAndValidators -> ValidateGenesis',
    'duplicate auth accounts',
    'zero auth accounts',
    'duplicate balances',
    'staking denom drift away from `naet`',
    'mint denom drift away from `naet`',
    'fee denom drift away from `naet`',
    'token masters',
    'token wallets',
    'NFT collections/items',
    'SBT items',
    'duplicate contract addresses',
    'queued message sequences',
    'queue `next_sequence`',
    'Old Orbitalis public formats',
    '`ORB`, `norb`, `orb1`, and raw `0:` addresses',
    'explicit migration-only tooling',
    '`AE...` formats'
  )) {
  Assert-Contains -Text $genesisText -Pattern ([regex]::Escape($term)) -Message "genesis migration doc missing: $term"
}

foreach ($term in @(
    'Upgrade Dry-Run And Migration Checklist',
    'current consensus',
    'version `2`',
    'missing module versions',
    'impossible future versions',
    'export after the dry-run upgrade produces valid genesis',
    'Legacy Format Rule',
    '`ORB`, `norb`, `orb1`, and raw `0:`',
    'normalized to Aetra formats'
  )) {
  Assert-Contains -Text $upgradeText -Pattern ([regex]::Escape($term)) -Message "upgrade migration doc missing: $term"
}

foreach ($term in @(
    'import validation for duplicate contract addresses',
    'malformed contract state',
    'malformed queued messages',
    'duplicate queued sequences',
    '`next_sequence` drift',
    'duplicate contract address rejection',
    'malformed async queue state rejection'
  )) {
  Assert-Contains -Text $asyncText -Pattern ([regex]::Escape($term)) -Message "async doc missing Phase 13 import validation: $term"
}

Write-Host "genesis migration safety doc test passed"
