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
$expectedTestAssetAmount = "1000000000"
$expectedNativeAmount = "500000000"
$expectedSelfDelegation = "100000000"

foreach ($node in $nodes) {
  $nodeHome = Join-Path $node.FullName "orbitalisd"
  $genesisPath = Join-Path $nodeHome "config\genesis.json"
  $configToml = Join-Path $nodeHome "config\config.toml"
  $appToml = Join-Path $nodeHome "config\app.toml"
  $nodeIndex = [int]($node.Name -replace '^\D+', '')
  $ports = Get-LocalnetPortProfile `
    -Index $nodeIndex `
    -BaseP2PPort $BaseP2PPort `
    -BaseRPCPort $BaseRPCPort `
    -BaseRESTPort $BaseRESTPort `
    -BaseGRPCPort $BaseGRPCPort `
    -BasePprofPort $BasePprofPort `
    -PortStride $PortStride

  & $Binary genesis validate-genesis $genesisPath --home $nodeHome
  if ($LASTEXITCODE -ne 0) {
    throw "genesis validation failed for $genesisPath"
  }

  $configRaw = Get-Content -Raw -LiteralPath $configToml
  $appRaw = Get-Content -Raw -LiteralPath $appToml
  if ($configRaw -notmatch "(?m)^moniker = `"$([regex]::Escape($node.Name))`"$") {
    throw "config moniker for $($node.Name) does not match node directory"
  }
  if ($EnableRPC -and $configRaw -notmatch [regex]::Escape("tcp://0.0.0.0:$($ports.RPC)")) {
    throw "RPC port for $($node.Name) does not match profile port $($ports.RPC)"
  }
  if ($EnableAPI -and $appRaw -notmatch [regex]::Escape("tcp://0.0.0.0:$($ports.REST)")) {
    throw "REST port for $($node.Name) does not match profile port $($ports.REST)"
  }
  if ($EnableGRPC -and $appRaw -notmatch [regex]::Escape("127.0.0.1:$($ports.GRPC)")) {
    throw "gRPC port for $($node.Name) does not match profile port $($ports.GRPC)"
  }
  if ($appRaw -notmatch '(?m)^minimum-gas-prices = "0norb"$') {
    throw "minimum-gas-prices for $($node.Name) must be 0norb"
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

  $genTxs = @($appState.genutil.gen_txs)
  if ($genTxs.Count -ne $ValidatorCount) {
    throw "expected $ValidatorCount gentxs, found $($genTxs.Count)"
  }

  $balances = @($appState.bank.balances)
  if ($balances.Count -ne $ValidatorCount) {
    throw "expected $ValidatorCount initial bank balances, found $($balances.Count)"
  }
  foreach ($balance in $balances) {
    $coins = @($balance.coins)
    $testAsset = @($coins | Where-Object { $_.denom -eq "testtoken" })
    $native = @($coins | Where-Object { $_.denom -eq "norb" })
    if ($testAsset.Count -ne 1 -or $testAsset[0].amount -ne $expectedTestAssetAmount) {
      throw "initial account $($balance.address) must have ${expectedTestAssetAmount}testtoken"
    }
    if ($native.Count -ne 1 -or $native[0].amount -ne $expectedNativeAmount) {
      throw "initial account $($balance.address) must have ${expectedNativeAmount}norb"
    }
  }

  foreach ($genTx in $genTxs) {
    $genTxRaw = $genTx | ConvertTo-Json -Depth 100 -Compress
    if ($genTxRaw -notmatch '"@type":"/cosmos.staking.v1beta1.MsgCreateValidator"') {
      throw "gentx does not contain MsgCreateValidator"
    }
    if ($genTxRaw -notmatch '"denom":"norb"' -or $genTxRaw -notmatch "`"amount`":`"$expectedSelfDelegation`"") {
      throw "gentx self-delegation must be ${expectedSelfDelegation}norb"
    }
  }
}

Write-Host "Validated $ValidatorCount-node genesis for $ChainId at $OutputDir"
Write-Host "genesis sha256: $firstHash"
