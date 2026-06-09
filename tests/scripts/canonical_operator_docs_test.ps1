param(
  [string]$Validator = "docs\VALIDATOR.md",
  [string]$Testnet = "docs\TESTNET.md",
  [string]$Cosmovisor = "docs\COSMOVISOR.md",
  [string]$Health = "docs\HEALTH.md",
  [string]$AVM = "docs\AVM.md",
  [string]$UserStaking = "docs\official-liquid-staking.md",
  [string]$UserStakingModel = "docs\native-account-staking-reputation.md"
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

$validatorText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Validator)
$testnetText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Testnet)
$cosmovisorText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Cosmovisor)
$healthText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Health)
$avmText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $AVM)
$userStakingText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $UserStaking)
$userStakingModelText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $UserStakingModel)

foreach ($term in @(
    'Hardware',
    'OS',
    'Build Or Download Binary',
    'Version Verification',
    'Chain ID',
    'Genesis Validation',
    'Keyring',
    'Validator Key Safety',
    'State Sync',
    'Snapshots',
    'Create Validator',
    'Monitor',
    'Restart',
    'Upgrade',
    'Incident Response'
  )) {
  Assert-Contains -Text $validatorText -Pattern ([regex]::Escape($term)) -Message "validator doc missing section: $term"
}

foreach ($term in @(
    'Chain ID',
    'AWCE-1 Wallet Compatibility Summary',
    'Genesis URL And Checksum',
    'Seed Nodes And Persistent Peers',
    'RPC Endpoints',
    'Faucet Path',
    'Minimum Fees',
    'Expected Block Time',
    'Launch Profile',
    'Known Non-Goals'
  )) {
  Assert-Contains -Text $testnetText -Pattern ([regex]::Escape($term)) -Message "testnet doc missing section: $term"
}

foreach ($term in @(
    'Install',
    'Directory Layout',
    'Current Binary',
    'Upgrades Directory',
    'Environment',
    'Upgrade Handler Naming',
    'Rollback Policy'
  )) {
  Assert-Contains -Text $cosmovisorText -Pattern ([regex]::Escape($term)) -Message "cosmovisor doc missing section: $term"
}

foreach ($term in @(
    'Health Check',
    'Logs',
    'Diagnostics',
    'Troubleshooting Order',
    'Minimum Metrics'
  )) {
  Assert-Contains -Text $healthText -Pattern ([regex]::Escape($term)) -Message "health doc missing section: $term"
}

foreach ($term in @(
    'Canonical Bytecode And Module Format',
    'Verifier',
    'Instruction Set And Gas Schedule',
    'Typed Values',
    'Deterministic Stack Execution',
    'Chunk And ChunkMap Persistent State',
    'Deploy, Execute, And Get Methods',
    'Receipts, Events, And Proofs',
    'Host Function Allowlist',
    'Examples In CI',
    'Non-Goals'
  )) {
  Assert-Contains -Text $avmText -Pattern ([regex]::Escape($term)) -Message "AVM doc missing section: $term"
}

foreach ($term in @(
    '10 AET',
    'MinPoolDeposit'
  )) {
  Assert-Contains -Text $userStakingText -Pattern ([regex]::Escape($term)) -Message "official liquid staking doc missing pool minimum term: $term"
  Assert-Contains -Text $userStakingModelText -Pattern ([regex]::Escape($term)) -Message "native staking model missing pool minimum term: $term"
}

foreach ($path in @($validatorText, $testnetText, $cosmovisorText, $healthText, $avmText)) {
  Assert-NotContains -Text $path -Pattern 'tx staking delegate|MsgDelegate|direct delegation example' -Message "canonical operator doc contains a normal direct-delegation example"
}

Write-Host "canonical operator docs test passed"
