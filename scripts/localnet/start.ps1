param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 0,
  [string]$ChainId = "orbitalis-local-1",
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
  [switch]$NoInit,
  [switch]$Wait,
  [switch]$CleanLogs,
  [int]$TimeoutSeconds = 60
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\orbitalisd.exe"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"

if (!(Test-Path $Binary) -or !(Test-Path $OutputDir)) {
  if ($NoInit) {
    throw "Binary or output directory is missing and -NoInit was specified: binary=$Binary output=$OutputDir"
  }
  $initArgs = @{
    OutputDir      = $OutputDir
    Binary         = $Binary
    ChainId        = $ChainId
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
  if ($ValidatorCount -gt 0) { $initArgs.ValidatorCount = $ValidatorCount }
  & (Join-Path $PSScriptRoot "init.ps1") @initArgs
}

$nodes = Get-LocalnetNodes -OutputDir $OutputDir

if ($ValidatorCount -gt 0 -and $nodes.Count -ne $ValidatorCount) {
  throw "Expected $ValidatorCount validator nodes under $OutputDir, found $($nodes.Count)"
}
if ($nodes.Count -lt 1) {
  throw "No validator node directories found under $OutputDir"
}
$actualValidatorCount = $nodes.Count

$pidDir = Join-Path $OutputDir "pids"
$logDir = Join-Path $OutputDir "logs"
New-Item -ItemType Directory -Force -Path $pidDir, $logDir | Out-Null
if ($CleanLogs) {
  Get-ChildItem -LiteralPath $logDir -Filter "*.log" -ErrorAction SilentlyContinue | Remove-Item -Force
}

Get-ChildItem -LiteralPath $pidDir -Filter *.pid -ErrorAction SilentlyContinue | ForEach-Object {
  $pidValue = [int](Get-Content -Raw -LiteralPath $_.FullName)
  $proc = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
  if ($proc) {
    throw "Localnet already appears to be running: $($_.FullName) pid=$pidValue"
  }
  Remove-Item -LiteralPath $_.FullName -Force
}

Assert-LocalnetPortsAvailable `
  -ValidatorCount $actualValidatorCount `
  -BaseP2PPort $BaseP2PPort `
  -BaseRPCPort $BaseRPCPort `
  -BaseRESTPort $BaseRESTPort `
  -BaseGRPCPort $BaseGRPCPort `
  -BasePprofPort $BasePprofPort `
  -PortStride $PortStride `
  -EnableAPI $EnableAPI `
  -EnableGRPC $EnableGRPC `
  -EnableRPC $EnableRPC

Repair-LocalnetProcessPathEnvironment

foreach ($node in $nodes) {
  $nodeName = $node.Name
  $nodeHome = Join-Path $node.FullName "orbitalisd"
  $stdout = Join-Path $logDir "$nodeName.out.log"
  $stderr = Join-Path $logDir "$nodeName.err.log"
  $proc = Start-Process -FilePath $Binary `
    -ArgumentList @("start", "--home", $nodeHome, "--log_level", $LogLevel) `
    -RedirectStandardOutput $stdout `
    -RedirectStandardError $stderr `
    -WindowStyle Hidden `
    -PassThru
  Set-Content -LiteralPath (Join-Path $pidDir "$nodeName.pid") -Value $proc.Id
  Write-Host "Started $nodeName pid=$($proc.Id)"
}

Write-Host "Logs: $logDir"

if ($Wait) {
  $p = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
  if ($EnableRPC) {
    Wait-LocalnetRpc -RPCPort $p.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Wait-LocalnetHeight -TargetHeight 1 -RPCPort $p.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  }
  if ($EnableAPI) {
    Wait-LocalnetRest -RESTPort $p.REST -TimeoutSeconds $TimeoutSeconds | Out-Null
  }
  if ($EnableGRPC) {
    Wait-LocalnetGrpc -GRPCPort $p.GRPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  }
}
