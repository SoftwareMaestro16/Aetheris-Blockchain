param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$RPCPort = 26657,
  [int]$Count = 1000,
  [decimal]$RatePerSecond = 25,
  [string]$FromKey = "node0",
  [int]$FromNodeIndex = 0,
  [string]$ToKey = "node1",
  [int]$ToNodeIndex = 1,
  [string]$Amount = "1naet",
  [string]$Fees = "300000naet",
  [ValidateSet("sync", "async")]
  [string]$BroadcastMode = "sync",
  [ValidateSet("rpc", "cli")]
  [string]$BroadcastTransport = "rpc",
  [int]$TimeoutSeconds = 600,
  [int]$CommandTimeoutSeconds = 60,
  [int]$PollIntervalMs = 1000,
  [int]$MaxValidatorRewardSamples = 25,
  [string]$WorkDir = "",
  [switch]$KeepTxFiles,
  [switch]$Json
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

if ($Count -lt 1) { throw "Count must be at least 1" }
if ($RatePerSecond -le 0) { throw "RatePerSecond must be positive" }
if ($RPCPort -lt 1 -or $RPCPort -gt 65535) { throw "RPCPort must be a valid TCP port" }
if ($TimeoutSeconds -lt 1) { throw "TimeoutSeconds must be positive" }
if ($CommandTimeoutSeconds -lt 1) { throw "CommandTimeoutSeconds must be positive" }
if ($PollIntervalMs -lt 100) { throw "PollIntervalMs must be at least 100" }
if ($ChainId -notmatch '(^|-)local($|-)' -and $ChainId -notmatch 'local') {
  throw "stress-profile is local/test-only; ChainId must contain 'local'"
}
if ($Fees -notmatch '^[0-9]+naet$') { throw "Fees must be a naet coin" }
if ($Amount -notmatch '^[1-9][0-9]*naet$') { throw "Amount must be a positive naet coin" }

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$WorkDir = Resolve-LocalnetPath -Path $WorkDir -DefaultRelativePath ".work\stress-profile"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"
Assert-LocalnetWorkspacePath -Path $WorkDir -Purpose "stress profile work directory"
if (-not (Test-Path -LiteralPath $Binary)) { throw "binary not found: $Binary" }

$node = "tcp://127.0.0.1:$RPCPort"
$rpcHttp = "http://127.0.0.1:$RPCPort"
$fromHome = Join-Path $OutputDir ("node{0}\aetrad" -f $FromNodeIndex)
$toHome = Join-Path $OutputDir ("node{0}\aetrad" -f $ToNodeIndex)
if (-not (Test-Path -LiteralPath $fromHome)) { throw "from node home not found: $fromHome" }
if (-not (Test-Path -LiteralPath $toHome)) { throw "to node home not found: $toHome" }

New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null
$stressLockPath = Join-Path $WorkDir "stress-profile-$RPCPort.lock"
if (Test-Path -LiteralPath $stressLockPath) {
  $existingPid = [int]0
  $existingText = Get-Content -Raw -LiteralPath $stressLockPath -ErrorAction SilentlyContinue
  [void][int]::TryParse(([string]$existingText).Trim(), [ref]$existingPid)
  if ($existingPid -gt 0 -and (Get-Process -Id $existingPid -ErrorAction SilentlyContinue)) {
    throw "another stress-profile.ps1 run is already active for RPC port $RPCPort in process $existingPid; wait for it or stop that PowerShell process first"
  }
  Remove-Item -LiteralPath $stressLockPath -Force -ErrorAction SilentlyContinue
}
Set-Content -LiteralPath $stressLockPath -Value ([string]$PID) -Encoding ASCII
$stressLockReleased = $false
trap {
  if ($stressLockPath -and -not $stressLockReleased) {
    try {
      Remove-Item -LiteralPath $stressLockPath -Force -ErrorAction SilentlyContinue
      $stressLockReleased = $true
    } catch {
    }
  }
  throw $_
}

function Invoke-StressCommandText {
  param(
    [string[]]$Arguments,
    [string]$FailureMessage = "aetrad command failed"
  )

  New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null
  $argString = (($Arguments | ForEach-Object {
        $arg = [string]$_
        if ($arg -notmatch '[\s"]') { return $arg }
        return '"' + (($arg -replace '\\(?=")', '\\') -replace '"', '\"') + '"'
      }) -join ' ')
  $commandId = [guid]::NewGuid().ToString("N")
  $stdoutPath = Join-Path $WorkDir "stress-cmd-$commandId.out"
  $stderrPath = Join-Path $WorkDir "stress-cmd-$commandId.err"

  $process = Start-Process `
    -FilePath $Binary `
    -ArgumentList $argString `
    -RedirectStandardOutput $stdoutPath `
    -RedirectStandardError $stderrPath `
    -WindowStyle Hidden `
    -PassThru

  if (-not $process.WaitForExit($CommandTimeoutSeconds * 1000)) {
    try {
      $process.Kill($true)
    } catch {
      try { $process.Kill() } catch {}
    }
    throw "$FailureMessage timed out after $CommandTimeoutSeconds seconds: $Binary $($Arguments -join ' ')"
  }
  $process.WaitForExit()
  $process.Refresh()

  $stdout = if (Test-Path -LiteralPath $stdoutPath) { Get-Content -Raw -LiteralPath $stdoutPath } else { "" }
  $stderr = if (Test-Path -LiteralPath $stderrPath) { Get-Content -Raw -LiteralPath $stderrPath } else { "" }
  Remove-Item -LiteralPath $stdoutPath, $stderrPath -Force -ErrorAction SilentlyContinue
  $text = (($stdout, $stderr) | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }) -join "`n"
  if ($null -ne $process.ExitCode -and $process.ExitCode -ne 0) {
    throw "$FailureMessage`: $Binary $($Arguments -join ' ')`n$text"
  }
  return $text
}

function Invoke-StressCliJson {
  param([string[]]$Arguments)
  $args = $Arguments + @("--node", $node, "--output", "json")
  $text = Invoke-StressCommandText -Arguments $args -FailureMessage "aetrad JSON command failed"
  $jsonStart = $text.IndexOf("{")
  if ($jsonStart -lt 0) {
    $jsonStart = $text.IndexOf("[")
  }
  if ($jsonStart -lt 0) {
    throw "aetrad command did not return JSON: $Binary $($args -join ' ')`n$text"
  }
  return $text.Substring($jsonStart) | ConvertFrom-Json
}

function Invoke-StressCliJsonNoNode {
  param([string[]]$Arguments)
  $args = $Arguments + @("--output", "json")
  $text = Invoke-StressCommandText -Arguments $args -FailureMessage "aetrad JSON command failed"
  $jsonStart = $text.IndexOf("{")
  if ($jsonStart -lt 0) {
    $jsonStart = $text.IndexOf("[")
  }
  if ($jsonStart -lt 0) {
    throw "aetrad command did not return JSON: $Binary $($args -join ' ')`n$text"
  }
  return $text.Substring($jsonStart) | ConvertFrom-Json
}

function Invoke-StressExternal {
  param(
    [string[]]$Arguments,
    [string]$FailureMessage
  )
  $text = Invoke-StressCommandText -Arguments $Arguments -FailureMessage $FailureMessage
  if ([string]::IsNullOrWhiteSpace($text)) { return @() }
  return $text -split "`r?`n"
}

function Convert-StressCoinParts {
  param([string]$Coin)
  $parts = Convert-LocalnetCoinParts -Coin $Coin
  return @{
    Amount = [decimal]$parts.Amount
    Denom = [string]$parts.Denom
  }
}

function Convert-StressDecimalString {
  param([object]$Value)
  if ($null -eq $Value) { return [decimal]0 }
  $text = [string]$Value
  if ([string]::IsNullOrWhiteSpace($text)) { return [decimal]0 }
  if ($text -match '^([0-9]+(?:\.[0-9]+)?)[A-Za-z][A-Za-z0-9/:._-]*$') {
    $text = $Matches[1]
  }
  $text = $text.Replace(',', '.')
  return [decimal]::Parse($text, [System.Globalization.CultureInfo]::InvariantCulture)
}

function Format-StressDecimal {
  param([decimal]$Value)
  return $Value.ToString([System.Globalization.CultureInfo]::InvariantCulture)
}

function Get-StressCoinAmount {
  param(
    [object]$Coin,
    [string]$Denom = "naet"
  )
  if ($null -eq $Coin) { return [decimal]0 }
  if ($Coin -is [array]) {
    $sum = [decimal]0
    foreach ($item in $Coin) { $sum += Get-StressCoinAmount -Coin $item -Denom $Denom }
    return $sum
  }
  if ($Coin -is [string]) {
    if ($Coin -match "^([0-9]+(?:\.[0-9]+)?)$Denom$") { return [decimal]$Matches[1] }
    return [decimal]0
  }
  if ($Coin.denom -eq $Denom -and $Coin.amount) {
    return Convert-StressDecimalString $Coin.amount
  }
  return [decimal]0
}

function Get-StressStatus {
  $status = Invoke-StressCliJson -Arguments @("status")
  return @{
    Height = [int64]$status.sync_info.latest_block_height
    Time = [datetime]$status.sync_info.latest_block_time
    CatchingUp = [bool]$status.sync_info.catching_up
  }
}

function Get-StressSupply {
  param([string]$Denom = "naet")
  $res = Invoke-StressCliJson -Arguments @("query", "bank", "total-supply-of", $Denom)
  if ($res.amount) { return Get-StressCoinAmount -Coin $res.amount -Denom $Denom }
  return Get-StressCoinAmount -Coin $res -Denom $Denom
}

function Get-StressBalance {
  param(
    [string]$Address,
    [string]$Denom = "naet"
  )
  $res = Invoke-StressCliJson -Arguments @("query", "bank", "balance", $Address, $Denom)
  if ($res.balance) { return Get-StressCoinAmount -Coin $res.balance -Denom $Denom }
  return Get-StressCoinAmount -Coin $res -Denom $Denom
}

function Get-StressBurned {
  param([string]$Denom = "naet")
  try {
    $res = Invoke-StressCliJson -Arguments @("query", "burn", "burned-by-denom")
    $entries = @()
    if ($res.burned_by_denom) { $entries += @($res.burned_by_denom) }
    if ($res.burnedByDenom) { $entries += @($res.burnedByDenom) }
    foreach ($entry in $entries) {
      if ($entry.denom -eq $Denom -or $entry.Denom -eq $Denom) {
        if ($entry.amount) { return Convert-StressDecimalString $entry.amount }
        if ($entry.Amount) { return Convert-StressDecimalString $entry.Amount }
      }
    }
  } catch {
  }
  return [decimal]0
}

function Get-StressFeeCollectorBalance {
 param([string]$Denom = "naet")
  $moduleBalance = Get-StressModuleBalance -ModuleName "feecollector" -Denom $Denom
  if ($moduleBalance -gt 0) { return $moduleBalance }
  try {
    $res = Invoke-StressCliJson -Arguments @("query", "feecollector", "fee-balances")
    if ($res.balances) {
      $sum = [decimal]0
      foreach ($property in $res.balances.PSObject.Properties) {
        $sum += Get-StressCoinAmount -Coin $property.Value -Denom $Denom
      }
      return $sum
    }
  } catch {
  }
  return [decimal]0
}

function Get-StressTreasuryBalance {
  param([string]$Denom = "naet")
  $moduleBalance = Get-StressModuleBalance -ModuleName "feecollector_treasury" -Denom $Denom
  if ($moduleBalance -gt 0) { return $moduleBalance }
  try {
    $treasury = Invoke-StressCliJson -Arguments @("query", "treasury", "treasury-balance")
    if ($treasury.balance) { return Get-StressCoinAmount -Coin $treasury.balance -Denom $Denom }
    if ($treasury.balances) { return Get-StressCoinAmount -Coin $treasury.balances -Denom $Denom }
    if ($treasury.module_account) {
      return Get-StressBalance -Address ([string]$treasury.module_account) -Denom $Denom
    }
  } catch {
  }
  return [decimal]0
}

function Get-StressModuleBalance {
  param(
    [string]$ModuleName,
    [string]$Denom = "naet"
  )
  try {
    $res = Invoke-StressCliJson -Arguments @("query", "fees", "module-balances")
    foreach ($balance in @($res.balances)) {
      if ([string]$balance.module_name -ne $ModuleName) {
        continue
      }
      return Get-StressCoinAmount -Coin @($balance.balance) -Denom $Denom
    }
  } catch {
  }
  return [decimal]0
}

function Get-StressValidatorRewardSample {
  param([int]$Limit = 25)
  $validators = @((Invoke-StressCliJson -Arguments @("query", "staking", "validators")).validators)
  $bonded = @($validators | Where-Object { [string]$_.status -eq "BOND_STATUS_BONDED" -or [string]$_.status -eq "3" })
  $samples = @()
  $outstandingTotal = [decimal]0
  $commissionTotal = [decimal]0
  foreach ($validator in ($bonded | Select-Object -First $Limit)) {
    $operator = [string]$validator.operator_address
    $outstanding = [decimal]0
    $commission = [decimal]0
    try {
      $rewards = Invoke-StressCliJson -Arguments @("query", "distribution", "validator-outstanding-rewards", $operator)
      if ($rewards.rewards.rewards) {
        $outstanding = Get-StressCoinAmount -Coin @($rewards.rewards.rewards) -Denom "naet"
      }
    } catch {
    }
    try {
      $comm = Invoke-StressCliJson -Arguments @("query", "distribution", "commission", $operator)
      if ($comm.commission.commission) {
        $commission = Get-StressCoinAmount -Coin @($comm.commission.commission) -Denom "naet"
      }
    } catch {
    }
    $outstandingTotal += $outstanding
    $commissionTotal += $commission
    $samples += [pscustomobject]@{
      operator = $operator
      moniker = [string]$validator.description.moniker
      status = [string]$validator.status
      voting_power_tokens = [string]$validator.tokens
      outstanding_rewards_naet = Format-StressDecimal $outstanding
      commission_naet = Format-StressDecimal $commission
    }
  }
  return @{
    ValidatorCount = @($validators).Count
    BondedCount = @($bonded).Count
    SampledCount = @($samples).Count
    OutstandingTotal = $outstandingTotal
    CommissionTotal = $commissionTotal
    Samples = @($samples)
  }
}

function Get-StressEconomicsSnapshot {
  param(
    [string]$FromAddress,
    [string]$ToAddress
  )
  $status = Get-StressStatus
  $validatorRewards = Get-StressValidatorRewardSample -Limit $MaxValidatorRewardSamples
  return [ordered]@{
    height = $status.Height
    block_time = $status.Time.ToUniversalTime().ToString("o")
    supply_naet = Format-StressDecimal (Get-StressSupply -Denom "naet")
    sender_balance_naet = Format-StressDecimal (Get-StressBalance -Address $FromAddress -Denom "naet")
    recipient_balance_naet = Format-StressDecimal (Get-StressBalance -Address $ToAddress -Denom "naet")
    burned_naet = Format-StressDecimal (Get-StressBurned -Denom "naet")
    treasury_balance_naet = Format-StressDecimal (Get-StressTreasuryBalance -Denom "naet")
    fee_collector_balance_naet = Format-StressDecimal (Get-StressFeeCollectorBalance -Denom "naet")
    validators = [ordered]@{
      total = $validatorRewards.ValidatorCount
      bonded = $validatorRewards.BondedCount
      sampled = $validatorRewards.SampledCount
      sampled_outstanding_rewards_naet = Format-StressDecimal $validatorRewards.OutstandingTotal
      sampled_commission_naet = Format-StressDecimal $validatorRewards.CommissionTotal
      samples = @($validatorRewards.Samples)
    }
  }
}

function Get-StressAccountNumbers {
  param([string]$Address)
  $account = Invoke-StressCliJson -Arguments @("query", "auth", "account", $Address)
  $value = $account.account.value
  $accountNumber = [uint64]0
  if ($value.account_number) { $accountNumber = [uint64]$value.account_number }
  $sequence = [uint64]0
  if ($value.sequence) { $sequence = [uint64]$value.sequence }
  return @{
    AccountNumber = $accountNumber
    Sequence = $sequence
  }
}

function Write-StressTxFile {
  param(
    [object]$Tx,
    [string]$Path
  )
  $utf8NoBom = New-Object System.Text.UTF8Encoding $false
  [System.IO.File]::WriteAllText($Path, ($Tx | ConvertTo-Json -Depth 100), $utf8NoBom)
}

function New-StressSignedBankTx {
  param(
    [int]$Index,
    [uint64]$AccountNumber,
    [uint64]$Sequence,
    [string]$FromAddress,
    [string]$ToAddress,
    [string]$TxDir
  )
  $unsigned = Invoke-StressCliJsonNoNode -Arguments @(
    "tx", "bank", "send", $FromKey, $ToAddress, $Amount,
    "--home", $fromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--fees", $Fees,
    "--generate-only"
  )
  $unsigned = Set-LocalnetBankSendMessageAddresses -UnsignedTx $unsigned -FromAddress $FromAddress -ToAddress $ToAddress
  $unsignedPath = Join-Path $TxDir ("tx-{0:D6}-unsigned.json" -f $Index)
  $signedPath = Join-Path $TxDir ("tx-{0:D6}-signed.json" -f $Index)
  Write-StressTxFile -Tx $unsigned -Path $unsignedPath
  Invoke-StressExternal -Arguments @(
    "tx", "sign", $unsignedPath,
    "--from", $FromKey,
    "--home", $fromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--account-number", ([string]$AccountNumber),
    "--sequence", ([string]$Sequence),
    "--offline",
    "--output", "json",
    "--output-document", $signedPath
  ) -FailureMessage "aetrad tx sign failed" | Out-Null
  if (-not (Test-Path -LiteralPath $signedPath)) { throw "signed tx file was not created: $signedPath" }
  return $signedPath
}

function Send-StressBroadcast {
  param([string]$Payload)
  $sw = [System.Diagnostics.Stopwatch]::StartNew()
  try {
    if ($BroadcastTransport -eq "rpc") {
      $tx = Send-StressRpcBroadcast -TxHex $Payload
      $sw.Stop()
      $code = if ($null -ne $tx.result.code) { [int]$tx.result.code } else { 0 }
      $hash = [string]$tx.result.hash
      return [ordered]@{
        ok = ($code -eq 0 -and -not [string]::IsNullOrWhiteSpace($hash))
        txhash = $hash
        code = $code
        latency_ms = [int64]$sw.ElapsedMilliseconds
        error = if ($code -eq 0) { "" } else { [string]$tx.result.log }
      }
    }
    $tx = Invoke-StressCliJson -Arguments @(
      "tx", "broadcast", $Payload,
      "--broadcast-mode", $BroadcastMode
    )
    $sw.Stop()
    $code = Get-LocalnetTxCode -Tx $tx
    $hash = Get-LocalnetTxHash -Tx $tx
    return [ordered]@{
      ok = ($code -eq 0 -and -not [string]::IsNullOrWhiteSpace($hash))
      txhash = $hash
      code = $code
      latency_ms = [int64]$sw.ElapsedMilliseconds
      error = if ($code -eq 0) { "" } else { Get-LocalnetTxLog -Tx $tx }
    }
  } catch {
    $sw.Stop()
    return [ordered]@{
      ok = $false
      txhash = ""
      code = -1
      latency_ms = [int64]$sw.ElapsedMilliseconds
      error = ($_.Exception.Message -replace '\s+', ' ')
    }
  }
}

function Convert-StressBase64ToHex {
  param([string]$Text)
  $bytes = [Convert]::FromBase64String($Text.Trim())
  return [System.BitConverter]::ToString($bytes).Replace("-", "").ToLowerInvariant()
}

function Encode-StressTxHex {
  param([string]$SignedPath)
  $text = Invoke-StressCommandText -Arguments @("tx", "encode", $SignedPath) -FailureMessage "aetrad tx encode failed"
  $encoded = (($text -split "`r?`n") | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | Select-Object -Last 1).Trim()
  return Convert-StressBase64ToHex -Text $encoded
}

function Send-StressRpcBroadcast {
  param([string]$TxHex)
  if ($BroadcastMode -eq "sync") {
    $endpoint = "broadcast_tx_sync"
  } else {
    $endpoint = "broadcast_tx_async"
  }
  $uri = "$rpcHttp/$endpoint`?tx=0x$TxHex"
  return Invoke-RestMethod -Method Get -Uri $uri -TimeoutSec $CommandTimeoutSeconds
}

function Get-StressTx {
  param([string]$TxHash)
  return Invoke-StressCliJson -Arguments @("query", "tx", $TxHash)
}

function Wait-StressTxs {
  param(
    [object[]]$Broadcasts,
    [int]$TimeoutSeconds
  )
  $pending = @{}
  foreach ($broadcast in $Broadcasts) {
    if ($broadcast.ok -and $broadcast.txhash) {
      $pending[$broadcast.txhash] = $broadcast
    }
  }
  $committed = @{}
  $failed = @{}
  $deadline = [datetime]::UtcNow.AddSeconds($TimeoutSeconds)
  while ($pending.Count -gt 0 -and [datetime]::UtcNow -lt $deadline) {
    foreach ($hash in @($pending.Keys)) {
      try {
        $tx = Get-StressTx -TxHash $hash
        $response = if ($tx.tx_response) { $tx.tx_response } else { $tx }
        $code = 0
        if ($null -ne $response.code -and -not [string]::IsNullOrWhiteSpace([string]$response.code)) {
          $code = [int]$response.code
        }
        $height = 0
        if ($null -ne $response.height -and -not [string]::IsNullOrWhiteSpace([string]$response.height)) {
          $height = [int64]$response.height
        }
        $record = [ordered]@{
          txhash = $hash
          height = $height
          code = $code
          raw_log = [string]$response.raw_log
          timestamp = [string]$response.timestamp
        }
        if ($code -eq 0) {
          $committed[$hash] = $record
        } else {
          $failed[$hash] = $record
        }
        $pending.Remove($hash)
      } catch {
      }
    }
    if ($pending.Count -gt 0) {
      Start-Sleep -Milliseconds $PollIntervalMs
    }
  }
  return @{
    Committed = @($committed.Values)
    Failed = @($failed.Values)
    Missing = @($pending.Keys)
  }
}

function Get-StressBlockTime {
  param([int64]$Height)
  if ($Height -le 0) { return $null }
  try {
    $block = Invoke-StressCliJson -Arguments @("query", "block", "--type", "height", ([string]$Height))
    if ($block.header.time) { return [datetime]$block.header.time }
    if ($block.block.header.time) { return [datetime]$block.block.header.time }
  } catch {
  }
  return $null
}

$status = Get-StressStatus
if ($status.CatchingUp) { throw "localnet RPC is catching up" }
if ((Invoke-StressCliJson -Arguments @("status")).node_info.network -ne $ChainId) {
  throw "connected chain-id mismatch"
}

New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null
$runId = (Get-Date).ToUniversalTime().ToString("yyyyMMddTHHmmssZ")
$txDir = Join-Path $WorkDir "tx-$runId"
New-Item -ItemType Directory -Force -Path $txDir | Out-Null

$fromAddress = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $fromHome -KeyName $FromKey
$toAddress = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $toHome -KeyName $ToKey
$coin = Convert-StressCoinParts -Coin $Amount
$fee = Convert-StressCoinParts -Coin $Fees
$account = Get-StressAccountNumbers -Address $fromAddress
$before = Get-StressEconomicsSnapshot -FromAddress $fromAddress -ToAddress $toAddress

$signedPaths = New-Object System.Collections.Generic.List[string]
for ($i = 0; $i -lt $Count; $i++) {
  $signedPaths.Add((New-StressSignedBankTx `
        -Index $i `
        -AccountNumber $account.AccountNumber `
        -Sequence ($account.Sequence + [uint64]$i) `
        -FromAddress $fromAddress `
        -ToAddress $toAddress `
        -TxDir $txDir))
}
$broadcastPayloads = New-Object System.Collections.Generic.List[string]
foreach ($signedPath in $signedPaths) {
  if ($BroadcastTransport -eq "rpc") {
    $broadcastPayloads.Add((Encode-StressTxHex -SignedPath $signedPath))
  } else {
    $broadcastPayloads.Add($signedPath)
  }
}

$broadcasts = @()
$intervalMs = [int][Math]::Ceiling(1000 / [double]$RatePerSecond)
$broadcastStart = [datetime]::UtcNow
for ($i = 0; $i -lt $broadcastPayloads.Count; $i++) {
  $iterationStart = [datetime]::UtcNow
  $result = Send-StressBroadcast -Payload $broadcastPayloads[$i]
  $result["index"] = $i
  $broadcasts += [pscustomobject]$result
  $elapsedMs = [int]([datetime]::UtcNow - $iterationStart).TotalMilliseconds
  $sleepMs = $intervalMs - $elapsedMs
  if ($sleepMs -gt 0 -and $i -lt ($broadcastPayloads.Count - 1)) {
    Start-Sleep -Milliseconds $sleepMs
  }
}
$broadcastEnd = [datetime]::UtcNow

$wait = Wait-StressTxs -Broadcasts $broadcasts -TimeoutSeconds $TimeoutSeconds
$after = Get-StressEconomicsSnapshot -FromAddress $fromAddress -ToAddress $toAddress

$committed = @($wait.Committed | Sort-Object @{ Expression = { [int64]$_.height }; Ascending = $true }, txhash)
$failedCommitted = @($wait.Failed)
$missing = @($wait.Missing)
$accepted = @($broadcasts | Where-Object { $_.ok })
$broadcastFailures = @($broadcasts | Where-Object { -not $_.ok })
$firstHeight = if ($committed.Count -gt 0) { [int64]$committed[0].height } else { 0 }
$lastHeight = if ($committed.Count -gt 0) { [int64]$committed[$committed.Count - 1].height } else { 0 }
$firstBlockTime = Get-StressBlockTime -Height $firstHeight
$lastBlockTime = Get-StressBlockTime -Height $lastHeight
$broadcastSeconds = [Math]::Round(($broadcastEnd - $broadcastStart).TotalSeconds, 4)
$finalitySecondsWall = [Math]::Round(([datetime]::UtcNow - $broadcastStart).TotalSeconds, 4)
$blockSpanSeconds = if ($firstBlockTime -and $lastBlockTime) { [Math]::Round(($lastBlockTime - $firstBlockTime).TotalSeconds, 4) } else { $null }

$beforeSupply = [decimal]$before.supply_naet
$afterSupply = [decimal]$after.supply_naet
$beforeSender = [decimal]$before.sender_balance_naet
$afterSender = [decimal]$after.sender_balance_naet
$beforeRecipient = [decimal]$before.recipient_balance_naet
$afterRecipient = [decimal]$after.recipient_balance_naet
$beforeBurn = [decimal]$before.burned_naet
$afterBurn = [decimal]$after.burned_naet
$beforeTreasury = [decimal]$before.treasury_balance_naet
$afterTreasury = [decimal]$after.treasury_balance_naet
$beforeFeeCollector = [decimal]$before.fee_collector_balance_naet
$afterFeeCollector = [decimal]$after.fee_collector_balance_naet
$beforeRewards = [decimal]$before.validators.sampled_outstanding_rewards_naet
$afterRewards = [decimal]$after.validators.sampled_outstanding_rewards_naet
$beforeCommission = [decimal]$before.validators.sampled_commission_naet
$afterCommission = [decimal]$after.validators.sampled_commission_naet

$expectedRecipientDelta = [decimal]$committed.Count * $coin.Amount
$expectedSenderDebit = [decimal]$committed.Count * ($coin.Amount + $fee.Amount)

$report = [ordered]@{
  profile = "local-bank-stress"
  chain_id = $ChainId
  rpc = "127.0.0.1:$RPCPort"
  from = [ordered]@{
    key = $FromKey
    node_index = $FromNodeIndex
    address = $fromAddress
    starting_sequence = $account.Sequence
    account_number = $account.AccountNumber
  }
  to = [ordered]@{
    key = $ToKey
    node_index = $ToNodeIndex
    address = $toAddress
  }
  tx = [ordered]@{
    requested = $Count
    amount = $Amount
    fees = $Fees
    broadcast_mode = $BroadcastMode
    broadcast_transport = $BroadcastTransport
    target_rate_tps = [double]$RatePerSecond
    broadcasts_accepted = $accepted.Count
    broadcast_failures = $broadcastFailures.Count
    committed_success = $committed.Count
    committed_failed = $failedCommitted.Count
    missing_after_timeout = $missing.Count
    failure_rate = if ($Count -gt 0) { [Math]::Round(($broadcastFailures.Count + $failedCommitted.Count + $missing.Count) / [double]$Count, 6) } else { $null }
  }
  performance = [ordered]@{
    broadcast_seconds = $broadcastSeconds
    broadcast_tps = if ($broadcastSeconds -gt 0) { [Math]::Round($accepted.Count / $broadcastSeconds, 4) } else { $null }
    finality_wall_seconds = $finalitySecondsWall
    committed_tps_wall = if ($finalitySecondsWall -gt 0) { [Math]::Round($committed.Count / $finalitySecondsWall, 4) } else { $null }
    first_tx_height = $firstHeight
    last_tx_height = $lastHeight
    tx_height_span = if ($firstHeight -gt 0 -and $lastHeight -ge $firstHeight) { $lastHeight - $firstHeight + 1 } else { 0 }
    first_tx_block_time = if ($firstBlockTime) { $firstBlockTime.ToUniversalTime().ToString("o") } else { $null }
    final_tx_block_time = if ($lastBlockTime) { $lastBlockTime.ToUniversalTime().ToString("o") } else { $null }
    block_span_seconds = $blockSpanSeconds
    committed_tps_by_block_time = if ($blockSpanSeconds -and $blockSpanSeconds -gt 0) { [Math]::Round($committed.Count / $blockSpanSeconds, 4) } else { $null }
  }
  validators = [ordered]@{
    before_total = $before.validators.total
    before_bonded = $before.validators.bonded
    after_total = $after.validators.total
    after_bonded = $after.validators.bonded
    sampled = $after.validators.sampled
    sampled_outstanding_rewards_delta_naet = Format-StressDecimal ($afterRewards - $beforeRewards)
    sampled_commission_delta_naet = Format-StressDecimal ($afterCommission - $beforeCommission)
    sample_after = @($after.validators.samples)
  }
  economics = [ordered]@{
    supply_before_naet = Format-StressDecimal $beforeSupply
    supply_after_naet = Format-StressDecimal $afterSupply
    supply_delta_naet = Format-StressDecimal ($afterSupply - $beforeSupply)
    sender_balance_delta_naet = Format-StressDecimal ($afterSender - $beforeSender)
    recipient_balance_delta_naet = Format-StressDecimal ($afterRecipient - $beforeRecipient)
    expected_sender_debit_naet = Format-StressDecimal $expectedSenderDebit
    expected_recipient_credit_naet = Format-StressDecimal $expectedRecipientDelta
    burned_delta_naet = Format-StressDecimal ($afterBurn - $beforeBurn)
    treasury_balance_delta_naet = Format-StressDecimal ($afterTreasury - $beforeTreasury)
    fee_collector_balance_delta_naet = Format-StressDecimal ($afterFeeCollector - $beforeFeeCollector)
    total_fee_paid_naet = Format-StressDecimal ([decimal]$committed.Count * $fee.Amount)
    note = "Bank transfer stress proves fee charging and validator finality. Burn/treasury/reward deltas are reported when the active modules expose changing on-chain accounting for this tx path."
  }
  before = $before
  after = $after
  failures = [ordered]@{
    broadcast_samples = @($broadcastFailures | Select-Object -First 5 index, code, error)
    committed_failure_samples = @($failedCommitted | Select-Object -First 5 txhash, height, code, raw_log)
    missing_samples = @($missing | Select-Object -First 5)
  }
  artifacts = [ordered]@{
    tx_dir = $txDir
    kept_tx_files = [bool]$KeepTxFiles
  }
}

if (-not $KeepTxFiles) {
  Remove-Item -LiteralPath $txDir -Recurse -Force
  $report.artifacts.tx_dir = ""
}

if ($Json) {
  $report | ConvertTo-Json -Depth 20
} else {
  Write-Host "profile: $($report.profile)"
  Write-Host "tx: requested=$Count accepted=$($report.tx.broadcasts_accepted) committed=$($report.tx.committed_success) failures=$($report.tx.broadcast_failures + $report.tx.committed_failed + $report.tx.missing_after_timeout)"
  Write-Host "performance: broadcast_tps=$($report.performance.broadcast_tps) committed_tps_wall=$($report.performance.committed_tps_wall) finality_wall_seconds=$($report.performance.finality_wall_seconds)"
  Write-Host "heights: $($report.performance.first_tx_height)->$($report.performance.last_tx_height) final_tx_block_time=$($report.performance.final_tx_block_time)"
  Write-Host "validators: bonded $($report.validators.before_bonded)->$($report.validators.after_bonded) sampled_rewards_delta=$($report.validators.sampled_outstanding_rewards_delta_naet)naet sampled_commission_delta=$($report.validators.sampled_commission_delta_naet)naet"
  Write-Host "economics: supply_delta=$($report.economics.supply_delta_naet)naet burned_delta=$($report.economics.burned_delta_naet)naet treasury_delta=$($report.economics.treasury_balance_delta_naet)naet fee_collector_delta=$($report.economics.fee_collector_balance_delta_naet)naet"
}

if (-not $stressLockReleased) {
  Remove-Item -LiteralPath $stressLockPath -Force -ErrorAction SilentlyContinue
  $stressLockReleased = $true
}
