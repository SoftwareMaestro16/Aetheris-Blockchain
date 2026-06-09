param(
  [string]$OutputRoot = ".work\localnet-validate-genesis-test"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$OutputRoot = if ([System.IO.Path]::IsPathRooted($OutputRoot)) {
  [System.IO.Path]::GetFullPath($OutputRoot)
} else {
  [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $OutputRoot))
}

function Assert-TestPathInsideRepo {
  param([string]$Path)
  $repo = $RepoRoot.TrimEnd('\', '/')
  $full = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  if ($full -eq $repo -or -not $full.StartsWith($repo + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing test path outside repo: $full"
  }
}

function Assert-True {
  param([bool]$Condition, [string]$Message)
  if (-not $Condition) { throw $Message }
}

function Assert-Fails {
  param([scriptblock]$Script, [string]$Pattern)
  try {
    & $Script | Out-Null
  } catch {
    if ($_.Exception.Message -notmatch $Pattern) {
      throw "Expected failure matching '$Pattern', got: $($_.Exception.Message)"
    }
    return
  }
  throw "Expected command to fail: $Pattern"
}

function New-FakeAetrad {
  param([string]$Path)
  @'
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Args)
if ($Args.Count -ge 3 -and $Args[0] -eq "genesis" -and $Args[1] -eq "validate-genesis") {
  $raw = Get-Content -Raw -LiteralPath $Args[2]
  if ($raw -match '(?i)bad-validator') { Write-Error "invalid module params"; exit 1 }
  exit 0
}
Write-Error "unexpected fake aetrad args: $($Args -join ' ')"
exit 1
'@ | Set-Content -LiteralPath $Path
}

function New-TestLocalnet {
  param([string]$Path, [string]$ChainId = "aetra-local-1", [string]$Mutation = "")
  New-Item -ItemType Directory -Force -Path $Path | Out-Null
  for ($i = 0; $i -lt 3; $i++) {
    $nodeHome = Join-Path $Path "node$i\aetrad"
    New-Item -ItemType Directory -Force -Path (Join-Path $nodeHome "config") | Out-Null
    $rpcPort = 26657 + ($i * 100)
    $restPort = 1317 + $i
    $grpcPort = 9090 + $i
    @"
moniker = "node$i"
laddr = "tcp://0.0.0.0:$rpcPort"
"@ | Set-Content -LiteralPath (Join-Path $nodeHome "config\config.toml")
    @"
minimum-gas-prices = "0naet"
address = "tcp://0.0.0.0:$restPort"
address = "127.0.0.1:$grpcPort"
"@ | Set-Content -LiteralPath (Join-Path $nodeHome "config\app.toml")
    $nativeDenom = if ($Mutation -eq "wrong-denom") { "uaet" } else { "naet" }
    $extra = if ($Mutation -eq "secret" -and $i -eq 1) { ',"mnemonic":"never store this"' } elseif ($Mutation -eq "bad-validator" -and $i -eq 0) { ',"bad-validator":true' } else { "" }
    $variant = if ($Mutation -eq "hash-mismatch" -and $i -eq 2) { ',"extra":"mismatch"' } else { "" }
    $genesis = @"
{
  "chain_id": "$ChainId",
  "app_state": {
    "load": {"Params": {"Enabled": false}},
    "routing": {"Params": {"Enabled": false}},
    "zones": {"Params": {"Enabled": false}, "State": {"ActiveZones": []}},
    "mesh": {"Params": {"Enabled": false}, "State": {"Destinations": []}},
    "bank": {
      "denom_metadata": [{"base":"$nativeDenom","display":"AET"}],
      "balances": [
        {"address":"AE111","coins":[{"denom":"testtoken","amount":"1000000000"},{"denom":"naet","amount":"500000000"}]},
        {"address":"AE222","coins":[{"denom":"testtoken","amount":"1000000000"},{"denom":"naet","amount":"500000000"}]},
        {"address":"AE333","coins":[{"denom":"testtoken","amount":"1000000000"},{"denom":"naet","amount":"500000000"}]}
      ]
    },
    "staking": {"params": {"bond_denom": "naet"}},
    "mint": {"params": {"mint_denom": "naet"}},
    "fees": {"params": {"allowed_fee_denoms": ["naet"]}},
    "genutil": {"gen_txs": [
      {"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgCreateValidator","value":{"denom":"naet","amount":"100000000"}}]}},
      {"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgCreateValidator","value":{"denom":"naet","amount":"100000000"}}]}},
      {"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgCreateValidator","value":{"denom":"naet","amount":"100000000"}}]}}
    ]}
  }$extra$variant
}
"@
    $genesis | Set-Content -LiteralPath (Join-Path $nodeHome "config\genesis.json")
  }
}

Assert-TestPathInsideRepo -Path $OutputRoot
if (Test-Path -LiteralPath $OutputRoot) {
  Remove-Item -LiteralPath $OutputRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $OutputRoot | Out-Null
$fake = Join-Path $OutputRoot "aetrad.ps1"
New-FakeAetrad -Path $fake

$good = Join-Path $OutputRoot "good"
New-TestLocalnet -Path $good
& .\scripts\localnet\validate-genesis.ps1 -OutputDir $good -Binary $fake -ValidatorCount 3 -ChainId "aetra-local-1" | Out-Null

$mismatch = Join-Path $OutputRoot "mismatch"
New-TestLocalnet -Path $mismatch -Mutation "hash-mismatch"
Assert-Fails { & .\scripts\localnet\validate-genesis.ps1 -OutputDir $mismatch -Binary $fake -ValidatorCount 3 -ChainId "aetra-local-1" } "genesis hash mismatch"

$secret = Join-Path $OutputRoot "secret"
New-TestLocalnet -Path $secret -Mutation "secret"
Assert-Fails { & .\scripts\localnet\validate-genesis.ps1 -OutputDir $secret -Binary $fake -ValidatorCount 3 -ChainId "aetra-local-1" } "secret-like"

$wrongDenom = Join-Path $OutputRoot "wrong-denom"
New-TestLocalnet -Path $wrongDenom -Mutation "wrong-denom"
Assert-Fails { & .\scripts\localnet\validate-genesis.ps1 -OutputDir $wrongDenom -Binary $fake -ValidatorCount 3 -ChainId "aetra-local-1" } "native token metadata"

Assert-Fails { & .\scripts\localnet\validate-genesis.ps1 -OutputDir $good -Binary $fake -ValidatorCount 11 -ChainId "aetra-local-1" } "at most 10"

Write-Host "localnet validate genesis script test passed"
