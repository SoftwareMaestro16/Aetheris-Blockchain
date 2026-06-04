param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "orbitalis-local-1",
  [int]$ValidatorCount = 3,
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
  [bool]$EnableRPC = $true
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\orbitalisd.exe"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"
if ($ValidatorCount -lt 1) { throw "ValidatorCount must be at least 1" }

if (!(Test-Path $Binary) -or !(Test-Path $OutputDir)) {
  & (Join-Path $PSScriptRoot "init.ps1") `
    -OutputDir $OutputDir `
    -Binary $Binary `
    -ValidatorCount $ValidatorCount `
    -ChainId $ChainId `
    -BaseP2PPort $BaseP2PPort `
    -BaseRPCPort $BaseRPCPort `
    -BaseRESTPort $BaseRESTPort `
    -BaseGRPCPort $BaseGRPCPort `
    -BasePprofPort $BasePprofPort `
    -PortStride $PortStride `
    -TimeoutCommit $TimeoutCommit `
    -LogLevel $LogLevel `
    -EnableAPI $EnableAPI `
    -EnableGRPC $EnableGRPC `
    -EnableRPC $EnableRPC
}

$nodes = Get-LocalnetNodes -OutputDir $OutputDir

if ($nodes.Count -ne $ValidatorCount) {
  throw "Expected $ValidatorCount validator nodes under $OutputDir, found $($nodes.Count)"
}

$firstHash = $null
$secretPattern = '(?i)\b(mnemonic|private[_-]?key|priv_validator|secret|seed|wallet)\b'

foreach ($node in $nodes) {
  $nodeHome = Join-Path $node.FullName "orbitalisd"
  $genesisPath = Join-Path $nodeHome "config\genesis.json"

  & $Binary genesis validate-genesis $genesisPath --home $nodeHome
  if ($LASTEXITCODE -ne 0) {
    throw "genesis validation failed for $genesisPath"
  }

  $raw = Get-Content -Raw -LiteralPath $genesisPath
  if ($raw -match $secretPattern) {
    throw "genesis for $($node.Name) contains secret-like material"
  }

  $hash = (Get-FileHash -LiteralPath $genesisPath -Algorithm SHA256).Hash
  if ($null -eq $firstHash) {
    $firstHash = $hash
  } elseif ($hash -ne $firstHash) {
    throw "genesis hash mismatch for $($node.Name): expected $firstHash, got $hash"
  }

  $doc = $raw | ConvertFrom-Json
  if ($doc.chain_id -ne $ChainId) {
    throw "unexpected chain-id for $($node.Name): $($doc.chain_id)"
  }

  $appState = $doc.app_state
  if ($null -eq $appState) {
    throw "missing app_state for $($node.Name)"
  }

  $bankMetadata = @($appState.bank.denom_metadata | Where-Object { $_.base -eq "norb" })
  if ($bankMetadata.Count -ne 1 -or $bankMetadata[0].display -ne "ORB") {
    throw "native token metadata for norb/ORB is missing or invalid"
  }

  if ($appState.staking.params.bond_denom -ne "norb") {
    throw "staking bond denom is not norb"
  }

  if ($appState.mint.params.mint_denom -ne "norb") {
    throw "mint denom is not norb"
  }

  $feeDenoms = @($appState.fees.params.allowed_fee_denoms)
  if ($feeDenoms.Count -ne 1 -or $feeDenoms[0] -ne "norb") {
    throw "fees module does not restrict fees to norb"
  }

  if (@($appState.tokenfactory.denoms).Count -ne 0) {
    throw "tokenfactory genesis is expected to start with no factory denoms"
  }

  if ([int64]$appState.dex.next_pool_id -ne 1 -or @($appState.dex.pools).Count -ne 0) {
    throw "dex genesis is expected to start with next_pool_id=1 and no pools"
  }

  $genTxs = @($appState.genutil.gen_txs)
  if ($genTxs.Count -ne $ValidatorCount) {
    throw "expected $ValidatorCount gentxs, found $($genTxs.Count)"
  }
}

Write-Host "Validated $ValidatorCount-node genesis for $ChainId at $OutputDir"
Write-Host "genesis sha256: $firstHash"
