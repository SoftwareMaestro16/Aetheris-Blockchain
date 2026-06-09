param(
  [string]$Script = "scripts\localnet\stress-profile.ps1"
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

$scriptPath = Resolve-RepoPath $Script
if (-not (Test-Path -LiteralPath $scriptPath)) { throw "stress profile script missing: $scriptPath" }

$tokens = $null
$parseErrors = $null
[System.Management.Automation.Language.Parser]::ParseFile($scriptPath, [ref]$tokens, [ref]$parseErrors) | Out-Null
if ($parseErrors -and $parseErrors.Count -gt 0) {
  throw "stress profile script has PowerShell parse errors: $($parseErrors[0].Message)"
}

$text = Get-Content -Raw -LiteralPath $scriptPath

foreach ($term in @(
    'Set-LocalnetBankSendMessageAddresses',
    'CommandTimeoutSeconds',
    'Invoke-StressCommandText',
    'stress-profile-$RPCPort.lock',
    'another stress-profile.ps1 run is already active',
    'BroadcastTransport',
    'Send-StressRpcBroadcast',
    'broadcast_tx_async',
    'broadcast_tx_sync',
    'tx", "encode"',
    'Get-StressModuleBalance',
    'feecollector_treasury',
    '--account-number',
    '--sequence',
    '--broadcast-mode',
    'sync',
    'async',
    'Wait-StressTxs',
    'query", "block", "--type", "height"',
    'broadcast_tps',
    'committed_tps_wall',
    'finality_wall_seconds',
    'final_tx_block_time',
    'supply_delta_naet',
    'burned_delta_naet',
    'treasury_balance_delta_naet',
    'fee_collector_balance_delta_naet',
    'sampled_outstanding_rewards_delta_naet',
    'sampled_commission_delta_naet'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($term)) -Message "stress profile script missing: $term"
}

Write-Host "stress profile script test passed"
