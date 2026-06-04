param(
  [string]$OutputDir = "",
  [string]$BundleDir = "",
  [int]$ValidatorCount = 0,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [bool]$EnableAPI = $true,
  [bool]$EnableGRPC = $true,
  [bool]$EnableRPC = $true,
  [int]$TimeoutSeconds = 10
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$RepoRoot = Get-LocalnetRepoRoot
$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"

if ([string]::IsNullOrWhiteSpace($BundleDir)) {
  $stamp = Get-Date -Format "yyyyMMdd-HHmmss"
  $BundleDir = Join-Path $RepoRoot ".work\diagnostics\localnet-$stamp"
} elseif (-not [System.IO.Path]::IsPathRooted($BundleDir)) {
  $BundleDir = Join-Path $RepoRoot $BundleDir
}
$BundleDir = [System.IO.Path]::GetFullPath($BundleDir)
Assert-LocalnetWorkspacePath -Path $BundleDir -Purpose "diagnostic bundle directory"

New-Item -ItemType Directory -Force -Path $BundleDir | Out-Null

$nodes = Get-LocalnetNodes -OutputDir $OutputDir
if ($ValidatorCount -gt 0 -and $nodes.Count -ne $ValidatorCount) {
  throw "Expected $ValidatorCount localnet nodes, found $($nodes.Count)"
}

$meta = [ordered]@{
  created_at_utc = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
  output_dir     = $OutputDir
  node_count     = $nodes.Count
  note           = "Excluded keyring data, priv_validator_key.json, priv_validator_state.json, and node_key.json."
}
$meta | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath (Join-Path $BundleDir "manifest.json")

$logDir = Join-Path $OutputDir "logs"
if (Test-Path -LiteralPath $logDir) {
  Copy-Item -LiteralPath $logDir -Destination (Join-Path $BundleDir "logs") -Recurse -Force
}

foreach ($node in $nodes) {
  $nodeName = $node.Name
  $nodeHome = Join-Path $node.FullName "orbitalisd"
  $safeConfigDir = Join-Path $BundleDir "$nodeName\config"
  New-Item -ItemType Directory -Force -Path $safeConfigDir | Out-Null

  foreach ($file in @("app.toml", "config.toml", "genesis.json")) {
    $source = Join-Path $nodeHome "config\$file"
    if (Test-Path -LiteralPath $source) {
      Copy-Item -LiteralPath $source -Destination (Join-Path $safeConfigDir $file) -Force
    }
  }
}

if ($EnableRPC) {
  $p = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
  $rpcDir = Join-Path $BundleDir "rpc"
  New-Item -ItemType Directory -Force -Path $rpcDir | Out-Null

  foreach ($item in @(
      @{ Name = "status.json"; Path = "status" },
      @{ Name = "net_info.json"; Path = "net_info" },
      @{ Name = "validators.json"; Path = "validators?per_page=100" }
    )) {
    try {
      Invoke-LocalnetRpc -RPCPort $p.RPC -Path $item.Path -TimeoutSeconds $TimeoutSeconds |
        ConvertTo-Json -Depth 20 |
        Set-Content -LiteralPath (Join-Path $rpcDir $item.Name)
    } catch {
      $_.Exception.Message | Set-Content -LiteralPath (Join-Path $rpcDir "$($item.Name).error.txt")
    }
  }

  try {
    & (Join-Path $PSScriptRoot "health.ps1") `
      -OutputDir $OutputDir `
      -ValidatorCount $ValidatorCount `
      -BaseP2PPort $BaseP2PPort `
      -BaseRPCPort $BaseRPCPort `
      -BaseRESTPort $BaseRESTPort `
      -BaseGRPCPort $BaseGRPCPort `
      -BasePprofPort $BasePprofPort `
      -PortStride $PortStride `
      -EnableAPI $EnableAPI `
      -EnableGRPC $EnableGRPC `
      -EnableRPC $EnableRPC `
      -TimeoutSeconds $TimeoutSeconds `
      -Json | Set-Content -LiteralPath (Join-Path $BundleDir "health.json")
  } catch {
    $_.Exception.Message | Set-Content -LiteralPath (Join-Path $BundleDir "health.error.txt")
  }
}

Write-Host "Diagnostic bundle written to $BundleDir"
