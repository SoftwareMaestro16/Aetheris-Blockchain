function Get-LocalnetStakingParams {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "staking", "params", "--node", $node, "--output", "json")
  return $result.params
}

function Get-LocalnetStakingValidators {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "staking", "validators", "--node", $node, "--output", "json")
  return @($result.validators)
}

function Get-LocalnetBondedValidator {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  foreach ($validator in @(Get-LocalnetStakingValidators -Binary $Binary -RPCPort $RPCPort)) {
    $status = [string]$validator.status
    if ($status -eq "BOND_STATUS_BONDED" -or $status -eq "3") {
      return $validator
    }
  }
  throw "No bonded staking validator found on RPC $RPCPort"
}

function Get-LocalnetSlashingParams {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "slashing", "params", "--node", $node, "--output", "json")
  return $result.params
}

function Get-LocalnetSigningInfos {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "slashing", "signing-infos", "--node", $node, "--output", "json")
  if ($result.info) { return @($result.info) }
  if ($result.signing_infos) { return @($result.signing_infos) }
  return @()
}

function Get-LocalnetBankMetadata {
  param(
    [string]$Binary,
    [string]$Denom = "naet",
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "bank", "denom-metadata", $Denom, "--node", $node, "--output", "json")
  return $result.metadata
}

function Get-LocalnetBankSupplyOf {
  param(
    [string]$Binary,
    [string]$Denom = "naet",
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "bank", "total-supply-of", $Denom, "--node", $node, "--output", "json")
  if ($result.amount) { return $result.amount }
  return $result
}

function Get-LocalnetBankBalance {
  param(
    [string]$Binary,
    [string]$Address,
    [string]$Denom = "naet",
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "bank", "balance", $Address, $Denom, "--node", $node, "--output", "json")
  if ($result.balance) { return $result.balance }
  return $result
}

function Get-LocalnetKeyAddress {
  param(
    [string]$Binary,
    [string]$NodeHome,
    [string]$KeyName
  )

  $output = & $Binary keys show $KeyName -a --home $NodeHome --keyring-backend test 2>&1
  if ($LASTEXITCODE -ne 0) {
    throw "failed to read key $KeyName from $NodeHome`n$($output -join "`n")"
  }
  $address = (($output | Select-Object -Last 1).ToString().Trim())
  $converted = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("address", "convert", $address)
  if (-not $converted.user_friendly) {
    throw "failed to convert key $KeyName address $address to AE user-facing format"
  }
  return [string]$converted.user_friendly
}

function Convert-LocalnetCoinParts {
  param([string]$Coin)

  if ($Coin -notmatch '^([0-9]+)([a-zA-Z][a-zA-Z0-9/:._-]*)$') {
    throw "invalid coin amount: $Coin"
  }
  return @{
    Amount = [decimal]$Matches[1]
    Denom  = $Matches[2]
  }
}

function Wait-LocalnetBankBalanceIncrease {
  param(
    [string]$Binary,
    [string]$Address,
    [string]$Amount,
    [object]$BeforeBalance,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  $coin = Convert-LocalnetCoinParts -Coin $Amount
  $beforeAmount = 0
  if ($BeforeBalance -and $BeforeBalance.amount) {
    $beforeAmount = [decimal]$BeforeBalance.amount
  }
  $expectedAmount = $beforeAmount + $coin.Amount

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "bank balance $Address >= $expectedAmount$($coin.Denom)" -Condition {
    $current = Get-LocalnetBankBalance -Binary $Binary -Address $Address -Denom $coin.Denom -RPCPort $RPCPort
    $currentAmount = 0
    if ($current -and $current.amount) {
      $currentAmount = [decimal]$current.amount
    }
    if ($currentAmount -ge $expectedAmount) {
      return @{
        Address = $Address
        Denom   = $coin.Denom
        Before  = $beforeAmount.ToString()
        Current = $currentAmount.ToString()
      }
    }
    return $null
  }
}

function Wait-LocalnetDelegationBalance {
  param(
    [string]$Binary,
    [string]$DelegatorAddress,
    [string]$ValidatorAddress,
    [string]$Amount,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  $coin = Convert-LocalnetCoinParts -Coin $Amount
  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "delegation $DelegatorAddress -> $ValidatorAddress = $($coin.Amount)$($coin.Denom)" -Condition {
    $delegation = Get-LocalnetDelegation -Binary $Binary -DelegatorAddress $DelegatorAddress -ValidatorAddress $ValidatorAddress -RPCPort $RPCPort
    $balance = $delegation.delegation_response.balance
    if ($balance -and $balance.denom -eq $coin.Denom -and [decimal]$balance.amount -eq $coin.Amount) {
      return $delegation
    }
    return $null
  }
}

function Set-LocalnetBankSendMessageAddresses {
  param(
    [object]$UnsignedTx,
    [string]$FromAddress,
    [string]$ToAddress
  )

  $body = Get-LocalnetObjectProperty -InputObject $UnsignedTx -Name "body"
  $messages = Get-LocalnetObjectProperty -InputObject $body -Name "messages"
  $messageList = @($messages)
  if ($messageList.Count -ne 1) {
    throw "bank send unsigned tx must contain exactly one message"
  }
  $message = $messageList[0]
  if (-not $message.PSObject.Properties["from_address"] -or -not $message.PSObject.Properties["to_address"]) {
    throw "bank send unsigned tx missing from_address/to_address fields"
  }
  $message.from_address = $FromAddress
  $message.to_address = $ToAddress
  return $UnsignedTx
}

function Send-LocalnetDelegateTx {
  param(
    [string]$Binary,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [string]$ValidatorAddress,
    [string]$Amount = "5000000naet",
    [string]$Fees = "300000naet",
    [string]$ChainId = "aetra-local-1",
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $delegatorAddress = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $FromHome -KeyName $FromKey
  $tx = Invoke-LocalnetCliJson -Binary $Binary -Arguments @(
    "tx", "staking", "delegate", $ValidatorAddress, $Amount,
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--fees", $Fees,
    "--yes",
    "--broadcast-mode", "sync",
    "--node", $node,
    "--output", "json"
  )

  $broadcastCode = Get-LocalnetTxCode -Tx $tx
  if ($broadcastCode -ne 0) {
    throw "staking delegate broadcast failed with code $broadcastCode`: $(Get-LocalnetTxLog -Tx $tx)"
  }

  $txHash = $tx.txhash
  if (-not $txHash -and $tx.tx_response) {
    $txHash = $tx.tx_response.txhash
  }
  if (-not $txHash) {
    throw "staking delegate did not return txhash"
  }

  return Wait-LocalnetDelegationBalance `
    -Binary $Binary `
    -DelegatorAddress $delegatorAddress `
    -ValidatorAddress $ValidatorAddress `
    -Amount $Amount `
    -RPCPort $RPCPort `
    -TimeoutSeconds $TimeoutSeconds
}

function Get-LocalnetDelegation {
  param(
    [string]$Binary,
    [string]$DelegatorAddress,
    [string]$ValidatorAddress,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  return Invoke-LocalnetCliJson -Binary $Binary -Arguments @(
    "query", "staking", "delegation", $DelegatorAddress, $ValidatorAddress,
    "--node", $node,
    "--output", "json"
  )
}

function Send-LocalnetBankTx {
  param(
    [string]$Binary,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [string]$ToAddress,
    [string]$Amount = "1000naet",
    [string]$Fees = "300000naet",
    [string]$ChainId = "aetra-local-1",
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $coin = Convert-LocalnetCoinParts -Coin $Amount
  $balanceBefore = Get-LocalnetBankBalance -Binary $Binary -Address $ToAddress -Denom $coin.Denom -RPCPort $RPCPort
  $fromAddress = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $FromHome -KeyName $FromKey
  $unsigned = Invoke-LocalnetCliJson -Binary $Binary -Arguments @(
    "tx", "bank", "send", $FromKey, $ToAddress, $Amount,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--fees", $Fees,
    "--generate-only",
    "--output", "json"
  )
  $unsigned = Set-LocalnetBankSendMessageAddresses -UnsignedTx $unsigned -FromAddress $fromAddress -ToAddress $ToAddress

  $workDir = Join-Path $FromHome "tmp-localnet-tx"
  New-Item -ItemType Directory -Force -Path $workDir | Out-Null
  $unsignedPath = Join-Path $workDir "bank-send-unsigned.json"
  $signedPath = Join-Path $workDir "bank-send-signed.json"
  $utf8NoBom = New-Object System.Text.UTF8Encoding $false
  [System.IO.File]::WriteAllText($unsignedPath, ($unsigned | ConvertTo-Json -Depth 100), $utf8NoBom)

  $signArgs = @(
    "tx", "sign", $unsignedPath,
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--node", $node,
    "--output", "json",
    "--output-document", $signedPath
  )
  Invoke-ExternalChecked -FilePath $Binary -Arguments $signArgs -FailureMessage "aetrad tx sign failed" | Out-Null
  if (-not (Test-Path -LiteralPath $signedPath)) {
    throw "aetrad tx sign did not create signed tx file: $signedPath"
  }

  $tx = Invoke-LocalnetCliJson -Binary $Binary -Arguments @(
    "tx", "broadcast", $signedPath,
    "--broadcast-mode", "sync",
    "--node", $node,
    "--output", "json"
  )
  $broadcastCode = Get-LocalnetTxCode -Tx $tx
  if ($broadcastCode -ne 0) {
    throw "bank send broadcast failed with code $broadcastCode`: $(Get-LocalnetTxLog -Tx $tx)"
  }

  $txHash = $tx.txhash
  if (-not $txHash -and $tx.tx_response) {
    $txHash = $tx.tx_response.txhash
  }
  if (-not $txHash) {
    throw "bank send did not return txhash"
  }

  Wait-LocalnetBankBalanceIncrease `
    -Binary $Binary `
    -Address $ToAddress `
    -Amount $Amount `
    -BeforeBalance $balanceBefore `
    -RPCPort $RPCPort `
    -TimeoutSeconds $TimeoutSeconds | Out-Null

  return $tx
}
