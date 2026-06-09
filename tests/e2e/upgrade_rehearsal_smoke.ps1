param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$UpgradeDelay = 2,
  [int]$TimeoutSeconds = 120,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [bool]$EnableAPI = $true,
  [bool]$EnableGRPC = $true,
  [bool]$EnableRPC = $true,
  [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet-upgrade-rehearsal"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$rpcNode = "tcp://127.0.0.1:$($node0Ports.RPC)"

function Invoke-WithCommonLocalnetArgs {
  param(
    [string]$ScriptPath,
    [hashtable]$Extra = @{}
  )

  $args = @{
    OutputDir      = $OutputDir
    Binary         = $Binary
    ChainId        = $ChainId
    ValidatorCount = $ValidatorCount
    BaseP2PPort    = $BaseP2PPort
    BaseRPCPort    = $BaseRPCPort
    BaseRESTPort   = $BaseRESTPort
    BaseGRPCPort   = $BaseGRPCPort
    BasePprofPort  = $BasePprofPort
    PortStride     = $PortStride
    TimeoutCommit  = $TimeoutCommit
    LogLevel       = $LogLevel
    EnableAPI      = $EnableAPI
    EnableGRPC     = $EnableGRPC
    EnableRPC      = $EnableRPC
  }
  foreach ($key in $Extra.Keys) {
    $args[$key] = $Extra[$key]
  }
  & $ScriptPath @args
}

function Assert-True {
  param(
    [bool]$Condition,
    [string]$Message
  )

  if (-not $Condition) {
    throw $Message
  }
}

function Write-UpgradeInfo {
  param(
    [string]$NodeHome,
    [string]$Name,
    [int64]$Height,
    [string]$Info
  )

  $upgradeDir = Join-Path $NodeHome "data"
  New-Item -ItemType Directory -Force -Path $upgradeDir | Out-Null
  $upgradeInfoPath = Join-Path $upgradeDir "upgrade-info.json"
  $payload = [ordered]@{
    name   = $Name
    height = [int64]$Height
    info   = $Info
  }
  $payload | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath $upgradeInfoPath
  return $upgradeInfoPath
}

Push-Location $RepoRoot
try {
  $goCache = Join-Path $RepoRoot ".work\gocache"
  New-Item -ItemType Directory -Force -Path $goCache | Out-Null
  $env:GOCACHE = $goCache

  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  $initExtra = @{}
  if ($SkipBuild) {
    $initExtra.SkipBuild = $true
  }
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\init.ps1" -Extra $initExtra
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\validate-genesis.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\start.ps1" -Extra @{ NoInit = $true }

  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $preHeight = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $preStatus = Invoke-LocalnetRpc -RPCPort $node0Ports.RPC -Path "status" -TimeoutSeconds $TimeoutSeconds
  $preAppHash = [string]$preStatus.result.sync_info.latest_app_hash
  Assert-True (-not [string]::IsNullOrWhiteSpace($preAppHash)) "pre-upgrade app hash is empty"

  $upgradeHeight = [int64]$preHeight + $UpgradeDelay
  Assert-True ($upgradeHeight -gt $preHeight) "upgrade height must be in the future"

  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  $nodes = Get-LocalnetNodes -OutputDir $OutputDir
  Assert-True ($nodes.Count -ge 1) "no localnet node directories found for upgrade rehearsal"
  foreach ($node in $nodes) {
    Write-UpgradeInfo -NodeHome (Join-Path $node.FullName "aetrad") -Name "rehearsal-noop" -Height $upgradeHeight -Info "upgrade rehearsal no-op"
  }

  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\start.ps1" -Extra @{ NoInit = $true }
  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $postHeight = Wait-LocalnetHeight -TargetHeight ([int64]$upgradeHeight + 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $postStatus = Invoke-LocalnetRpc -RPCPort $node0Ports.RPC -Path "status" -TimeoutSeconds $TimeoutSeconds
  $postAppHash = [string]$postStatus.result.sync_info.latest_app_hash
  Assert-True (-not [string]::IsNullOrWhiteSpace($postAppHash)) "post-upgrade app hash is empty"
  Write-Host "upgrade rehearsal advanced from height $preHeight to $postHeight"

  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  & .\scripts\localnet\export-genesis.ps1 -OutputDir $OutputDir -Binary $Binary -ChainId $ChainId | Out-Null
  Write-Host "upgrade rehearsal export validated"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
