param(
  [string]$OutputDir = "",
  [int]$ExpectedCount = 0,
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

if ($ExpectedCount -gt 0) {
  $validators = Wait-LocalnetValidators -ExpectedCount $ExpectedCount -RPCPort $ports.RPC -TimeoutSeconds $TimeoutSeconds
} else {
  $validators = Invoke-LocalnetRpc -RPCPort $ports.RPC -Path "validators?per_page=100" -TimeoutSeconds $TimeoutSeconds
}

$validatorList = @($validators.result.validators)
$result = [ordered]@{
  output_dir      = $OutputDir
  rpc_port        = $ports.RPC
  expected_count  = $ExpectedCount
  validator_count = $validatorList.Count
  validators      = @($validatorList | ForEach-Object {
      [ordered]@{
        address      = $_.address
        voting_power = $_.voting_power
        proposer     = $_.proposer_priority
      }
    })
}

if ($Json) {
  $result | ConvertTo-Json -Depth 8
} else {
  Write-Host "validator set contains $($validatorList.Count) validators on RPC $($ports.RPC)"
}
