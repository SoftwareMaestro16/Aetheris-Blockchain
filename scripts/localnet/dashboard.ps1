param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$Denom = "naet",
  [int]$RPCPort = 26657,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [switch]$NoColor
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$RepoRoot = Get-LocalnetRepoRoot
$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"

if ([string]::IsNullOrWhiteSpace($Binary)) {
  $Binary = Join-Path $RepoRoot "build\aetrad.exe"
}
if (-not [System.IO.Path]::IsPathRooted($Binary)) {
  $Binary = Join-Path $RepoRoot $Binary
}
if (-not (Test-Path -LiteralPath $Binary)) {
  throw "aetrad binary not found: $Binary"
}

$node = "tcp://127.0.0.1:$RPCPort"
$genesisPath = Join-Path $OutputDir "node0\aetrad\config\genesis.json"
if (-not (Test-Path -LiteralPath $genesisPath)) {
  throw "genesis not found: $genesisPath"
}
$genesis = Get-Content -Raw -LiteralPath $genesisPath | ConvertFrom-Json

function Write-DashTitle {
  param([string]$Text)
  if ($NoColor) {
    Write-Host ""
    Write-Host "== $Text =="
    return
  }
  Write-Host ""
  Write-Host "== $Text ==" -ForegroundColor Cyan
}

function Write-DashRow {
  param(
    [string]$Name,
    [object]$Value,
    [ConsoleColor]$Color = [ConsoleColor]::Gray
  )
  $line = "{0,-30} {1}" -f $Name, $Value
  if ($NoColor) {
    Write-Host $line
  } else {
    Write-Host $line -ForegroundColor $Color
  }
}

function Invoke-AetraJson {
  param([string[]]$Arguments)

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $Binary @Arguments 2>&1
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }

  if ($exitCode -ne 0) {
    throw "aetrad command failed: $Binary $($Arguments -join ' ')`n$($output -join "`n")"
  }

  $text = $output -join "`n"
  $jsonStart = $text.IndexOf("{")
  if ($jsonStart -lt 0) {
    $jsonStart = $text.IndexOf("[")
  }
  if ($jsonStart -lt 0) {
    throw "aetrad command did not return JSON: $Binary $($Arguments -join ' ')`n$text"
  }
  return $text.Substring($jsonStart) | ConvertFrom-Json
}

function Invoke-AetraJsonMaybe {
  param([string[]]$Arguments)
  try {
    return Invoke-AetraJson -Arguments $Arguments
  } catch {
    return $null
  }
}

function Get-Amount {
  param([object]$Coin)
  if ($null -eq $Coin) { return [decimal]0 }
  if ($Coin.amount) { return [decimal]$Coin.amount }
  if ($Coin.Amount) { return [decimal]$Coin.Amount }
  return [decimal]0
}

function Format-Amount {
  param([decimal]$Amount)
  return "$Amount$Denom"
}

function Get-GenesisSupply {
  $total = [decimal]0
  foreach ($balance in @($genesis.app_state.bank.balances)) {
    foreach ($coin in @($balance.coins)) {
      if ([string]$coin.denom -eq $Denom) {
        $total += [decimal]$coin.amount
      }
    }
  }
  return $total
}

function Get-ModuleBalanceRow {
  param([string]$ModuleName)

  $account = Invoke-AetraJsonMaybe -Arguments @("query", "auth", "module-account", $ModuleName, "--node", $node, "--output", "json")
  if ($null -eq $account -or $null -eq $account.account) {
    return [PSCustomObject]@{
      Module  = $ModuleName
      Address = "not found"
      Balance = "-"
    }
  }

  $address = $account.account.value.address
  if (-not $address) {
    $address = $account.account.address
  }

  $balance = Invoke-AetraJsonMaybe -Arguments @("query", "bank", "balance", $address, $Denom, "--node", $node, "--output", "json")
  $amount = if ($balance -and $balance.balance) { "$($balance.balance.amount)$Denom" } else { "0$Denom" }

  return [PSCustomObject]@{
    Module  = $ModuleName
    Address = $address
    Balance = $amount
  }
}

try {
  $status = Invoke-AetraJson -Arguments @("status", "--node", $node)
} catch {
  Write-DashTitle "Localnet Dashboard"
  Write-DashRow "RPC" "offline/unreachable at $node" Red
  Write-DashRow "Hint" "start localnet first, or inspect exported genesis offline" Yellow
  $message = [string]$_.Exception.Message
  if ($message -match "(post failed: .*)") {
    $message = $Matches[1]
  } else {
    $message = ($message -split "`r?`n")[0]
  }
  Write-DashRow "Error" $message DarkGray
  exit 2
}

$sync = $status.sync_info
$nodeInfo = $status.node_info
$netInfo = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "net_info"
$latestBlock = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "block"
$validatorsRpc = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "validators?per_page=100"

$supply = Invoke-AetraJson -Arguments @("query", "bank", "total-supply-of", $Denom, "--node", $node, "--output", "json")
$genesisSupply = Get-GenesisSupply
$currentSupply = Get-Amount -Coin $supply.amount
$supplyDelta = $currentSupply - $genesisSupply

$authAccounts = Invoke-AetraJsonMaybe -Arguments @("query", "auth", "accounts", "--node", $node, "--output", "json")
$stakingPool = Invoke-AetraJsonMaybe -Arguments @("query", "staking", "pool", "--node", $node, "--output", "json")
$stakingParams = Invoke-AetraJsonMaybe -Arguments @("query", "staking", "params", "--node", $node, "--output", "json")
$validators = Invoke-AetraJsonMaybe -Arguments @("query", "staking", "validators", "--node", $node, "--output", "json")
$feesParams = Invoke-AetraJsonMaybe -Arguments @("query", "fees", "params", "--node", $node, "--output", "json")
$protocolPool = Invoke-AetraJsonMaybe -Arguments @("query", "protocolpool", "community-pool", "--node", $node, "--output", "json")
$slashingParams = Invoke-AetraJsonMaybe -Arguments @("query", "slashing", "params", "--node", $node, "--output", "json")
$signingInfos = Invoke-AetraJsonMaybe -Arguments @("query", "slashing", "signing-infos", "--node", $node, "--output", "json")
$evidence = Invoke-AetraJsonMaybe -Arguments @("query", "evidence", "list", "--node", $node, "--output", "json")

$moduleNames = @(
  "mint-authority",
  "burn",
  "fee_collector",
  "treasury",
  "storage-rent",
  "delegator-protection",
  "validator-insurance",
  "reporter-rewards",
  "protocolpool",
  "distribution",
  "gov"
)
$moduleRows = @($moduleNames | ForEach-Object { Get-ModuleBalanceRow -ModuleName $_ })

$validatorRows = @()
foreach ($validator in @($validators.validators)) {
  $commission = Invoke-AetraJsonMaybe -Arguments @("query", "distribution", "commission", $validator.operator_address, "--node", $node, "--output", "json")
  $outstanding = Invoke-AetraJsonMaybe -Arguments @("query", "distribution", "validator-outstanding-rewards", $validator.operator_address, "--node", $node, "--output", "json")
  $validatorRows += [PSCustomObject]@{
    Moniker     = $validator.description.moniker
    Status      = $validator.status
    Jailed      = $validator.jailed
    Tokens      = "$($validator.tokens)$Denom"
    Commission  = if ($commission.commission.commission) { ($commission.commission.commission -join ", ") } else { "-" }
    Outstanding = if ($outstanding.rewards.rewards) { ($outstanding.rewards.rewards -join ", ") } else { "-" }
    Operator    = $validator.operator_address
  }
}

$bonded = if ($stakingPool.pool.bonded_tokens) { [decimal]$stakingPool.pool.bonded_tokens } else { [decimal]0 }
$notBonded = if ($stakingPool.pool.not_bonded_tokens) { [decimal]$stakingPool.pool.not_bonded_tokens } else { [decimal]0 }
$stakingTotal = $bonded + $notBonded
$stakingRatio = if ($currentSupply -gt 0) { [math]::Round(([double]($stakingTotal / $currentSupply) * 100), 4) } else { 0 }

$stakingValidatorItems = @($validators.validators)
$bondedValidatorCount = @($stakingValidatorItems | Where-Object { $_.status -eq "BOND_STATUS_BONDED" -or $_.status -eq "3" }).Count
$jailedValidatorCount = @($stakingValidatorItems | Where-Object { $_.jailed -eq $true }).Count
$signingInfoItems = if ($signingInfos.info) { @($signingInfos.info) } elseif ($signingInfos.signing_infos) { @($signingInfos.signing_infos) } else { @() }
$evidenceItems = if ($evidence.evidence) { @($evidence.evidence) } else { @() }
$accountItems = if ($authAccounts.accounts) { @($authAccounts.accounts) } else { @() }

Write-DashTitle "Aetra Localnet Dashboard"
Write-DashRow "Output dir" $OutputDir DarkGray
Write-DashRow "Chain ID" $nodeInfo.network Green
Write-DashRow "Node" "$($nodeInfo.moniker) / $($nodeInfo.id)" Green
Write-DashRow "RPC" $node Green
Write-DashRow "Catching up" $sync.catching_up $(if ($sync.catching_up) { "Yellow" } else { "Green" })
Write-DashRow "Latest height" $sync.latest_block_height Green
Write-DashRow "Earliest height" $sync.earliest_block_height Gray
Write-DashRow "Latest block time" $sync.latest_block_time Gray
Write-DashRow "Latest app hash" $sync.latest_app_hash DarkGray
Write-DashRow "Peers" $netInfo.result.n_peers Gray
Write-DashRow "Consensus validators" @($validatorsRpc.result.validators).Count Gray
Write-DashRow "Latest block txs" @($latestBlock.result.block.data.txs).Count Gray

Write-DashTitle "Supply And Emission"
Write-DashRow "Denom" $Denom Green
Write-DashRow "Genesis supply" (Format-Amount $genesisSupply) Gray
Write-DashRow "Current supply" (Format-Amount $currentSupply) Green
Write-DashRow "Supply delta from genesis" (Format-Amount $supplyDelta) $(if ($supplyDelta -ge 0) { "Green" } else { "Yellow" })
Write-DashRow "Burn note" "historical burn count requires event/index scan; module balances below show current sinks" Yellow

Write-DashTitle "Fees And Protocol Economy"
if ($feesParams.params) {
  Write-DashRow "Allowed fee denoms" ($feesParams.params.allowed_fee_denoms -join ", ") Green
  Write-DashRow "Min fee" "$($feesParams.params.min_fee_amount)$Denom" Gray
  Write-DashRow "Base fee" "$($feesParams.params.base_fee_amount)$Denom" Gray
  Write-DashRow "Max fee" "$($feesParams.params.max_fee_amount)$Denom" Gray
  Write-DashRow "Validator rewards ratio" $feesParams.params.validator_rewards_ratio Gray
  Write-DashRow "Community pool ratio" $feesParams.params.community_pool_ratio Gray
  Write-DashRow "Max tx gas" $feesParams.params.max_tx_gas Gray
  Write-DashRow "Max block gas" $feesParams.params.max_block_gas Gray
  Write-DashRow "Max block txs" $feesParams.params.max_block_txs Gray
}
if ($protocolPool.pool) {
  Write-DashRow "Protocol community pool" (($protocolPool.pool | ForEach-Object { "$($_.amount)$($_.denom)" }) -join ", ") Green
}

Write-DashTitle "Staking / PoS"
Write-DashRow "Bonded tokens" (Format-Amount $bonded) Green
Write-DashRow "Not bonded tokens" (Format-Amount $notBonded) Yellow
Write-DashRow "Total staking pool" (Format-Amount $stakingTotal) Green
Write-DashRow "Staking ratio" "$stakingRatio%" Green
Write-DashRow "Validators total" $stakingValidatorItems.Count Green
Write-DashRow "Validators bonded" $bondedValidatorCount Green
Write-DashRow "Validators jailed" $jailedValidatorCount $(if ($jailedValidatorCount -gt 0) { "Red" } else { "Green" })
if ($stakingParams.params) {
  Write-DashRow "Bond denom" $stakingParams.params.bond_denom Gray
  Write-DashRow "Unbonding time" $stakingParams.params.unbonding_time Gray
  Write-DashRow "Max validators" $stakingParams.params.max_validators Gray
}
$validatorRows | Format-Table -AutoSize

Write-DashTitle "Slashing / Evidence"
if ($slashingParams.params) {
  Write-DashRow "Signed blocks window" $slashingParams.params.signed_blocks_window Gray
  Write-DashRow "Min signed per window" $slashingParams.params.min_signed_per_window Gray
  Write-DashRow "Downtime jail duration" $slashingParams.params.downtime_jail_duration Gray
  Write-DashRow "Slash double-sign" $slashingParams.params.slash_fraction_double_sign Gray
  Write-DashRow "Slash downtime" $slashingParams.params.slash_fraction_downtime Gray
}
Write-DashRow "Signing infos" $signingInfoItems.Count Green
Write-DashRow "Evidence records" $evidenceItems.Count $(if ($evidenceItems.Count -gt 0) { "Yellow" } else { "Green" })
$signingInfoItems |
  Select-Object address,index_offset,missed_blocks_counter,jailed_until,tombstoned |
  Format-Table -AutoSize

Write-DashTitle "Accounts And Module Balances"
Write-DashRow "Current auth accounts" $accountItems.Count Green
Write-DashRow "Genesis bank balances" @($genesis.app_state.bank.balances).Count Gray
$moduleRows | Format-Table -AutoSize

Write-DashTitle "Node Storage"
$nodes = Get-LocalnetNodes -OutputDir $OutputDir
$nodeRows = @()
for ($i = 0; $i -lt $nodes.Count; $i++) {
  $p = Get-LocalnetPortProfile -Index $i -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
  $size = (Get-ChildItem -LiteralPath $nodes[$i].FullName -Recurse -File | Measure-Object Length -Sum).Sum
  $nodeRows += [PSCustomObject]@{
    Node   = $nodes[$i].Name
    SizeMB = "{0:N2}" -f ($size / 1MB)
    P2P    = $p.P2P
    RPC    = $p.RPC
    REST   = $p.REST
    GRPC   = $p.GRPC
  }
}
$nodeRows | Format-Table -AutoSize
