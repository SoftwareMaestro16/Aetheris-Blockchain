function Invoke-LocalnetRpc {
  param(
    [int]$RPCPort,
    [string]$Path,
    [int]$TimeoutSeconds = 2
  )

  return Invoke-RestMethod -Uri "http://127.0.0.1:$RPCPort/$Path" -TimeoutSec $TimeoutSeconds
}

function Wait-LocalnetCondition {
  param(
    [scriptblock]$Condition,
    [int]$TimeoutSeconds,
    [string]$Description,
    [int]$PollMilliseconds = 500
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  $lastError = $null
  while ((Get-Date) -lt $deadline) {
    try {
      $result = & $Condition
      if ($result) {
        return $result
      }
    } catch {
      $lastError = $_.Exception.Message
    }
    Start-Sleep -Milliseconds $PollMilliseconds
  }

  if ($lastError) {
    throw "Timed out waiting for $Description; last error: $lastError"
  }
  throw "Timed out waiting for $Description"
}

function Wait-LocalnetRpc {
  param(
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "RPC port $RPCPort" -Condition {
    $status = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "status"
    if ($status.result.node_info.network) { return $status }
    return $null
  }
}

function Get-LocalnetHeight {
  param([int]$RPCPort = 26657)

  $status = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "status"
  return [int64]$status.result.sync_info.latest_block_height
}

function Wait-LocalnetHeight {
  param(
    [int64]$TargetHeight,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "height $TargetHeight on RPC $RPCPort" -Condition {
    $height = Get-LocalnetHeight -RPCPort $RPCPort
    if ($height -ge $TargetHeight) { return $height }
    return $null
  }
}

function Wait-LocalnetValidators {
  param(
    [int]$ExpectedCount,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "$ExpectedCount validators on RPC $RPCPort" -Condition {
    $validators = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "validators?per_page=100"
    $count = @($validators.result.validators).Count
    if ($count -ne $ExpectedCount) { return $null }
    foreach ($validator in @($validators.result.validators)) {
      if ([int64]$validator.voting_power -le 0) { return $null }
    }
    return $validators
  }
}

function Get-LocalnetTotalVotingPower {
  param([int]$RPCPort = 26657)

  $validators = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "validators?per_page=100"
  $total = [int64]0
  foreach ($validator in @($validators.result.validators)) {
    $total += [int64]$validator.voting_power
  }
  return $total
}

function Wait-LocalnetTotalVotingPowerGreater {
  param(
    [int64]$PreviousPower,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "validator voting power greater than $PreviousPower on RPC $RPCPort" -Condition {
    $power = Get-LocalnetTotalVotingPower -RPCPort $RPCPort
    if ($power -gt $PreviousPower) { return $power }
    return $null
  }
}

function Wait-LocalnetPeers {
  param(
    [int]$ExpectedMinPeers,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "at least $ExpectedMinPeers peers on RPC $RPCPort" -Condition {
    $netInfo = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "net_info"
    $peers = [int]$netInfo.result.n_peers
    if ($peers -ge $ExpectedMinPeers) { return $netInfo }
    return $null
  }
}

function Wait-LocalnetRest {
  param(
    [int]$RESTPort = 1317,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "REST endpoint $RESTPort" -Condition {
    $latest = Invoke-RestMethod -Uri "http://127.0.0.1:$RESTPort/cosmos/base/tendermint/v1beta1/blocks/latest" -TimeoutSec 2
    if ($latest.block.header.height) { return $latest }
    return $null
  }
}

function Wait-LocalnetGrpc {
  param(
    [int]$GRPCPort = 9090,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "gRPC TCP endpoint $GRPCPort" -Condition {
    return (Test-LocalnetTcpPortOpen -Port $GRPCPort -TimeoutMilliseconds 500)
  }
}
