param(
  [string]$OutputDir = "",
  [int64]$TargetHeight = 1,
  [int]$TimeoutSeconds = 60,
  [int]$BaseRPCPort = 26657,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [switch]$Json
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"
$ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$height = Wait-LocalnetHeight -TargetHeight $TargetHeight -RPCPort $ports.RPC -TimeoutSeconds $TimeoutSeconds

$result = [ordered]@{
  output_dir    = $OutputDir
  rpc_port      = $ports.RPC
  target_height = $TargetHeight
  height        = [int64]$height
}

if ($Json) {
  $result | ConvertTo-Json -Depth 4
} else {
  Write-Host "localnet reached height $height on RPC $($ports.RPC)"
}
