param(
  [string]$PeersPath = "docs\testnet\peers.example.json",
  [string]$SeedsPath = "docs\testnet\seeds.example.txt"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Path))
}

function Assert-ValidNodeId {
  param(
    [string]$NodeId,
    [string]$Source
  )

  if ([string]::IsNullOrWhiteSpace($NodeId)) {
    throw "$Source node id is required"
  }
  if ($NodeId -notmatch '^[0-9a-fA-F]{40}$') {
    throw "$Source node id must be 40 hex characters: $NodeId"
  }
}

function Assert-ValidEndpoint {
  param(
    [string]$Endpoint,
    [string]$Source
  )

  if ([string]::IsNullOrWhiteSpace($Endpoint)) {
    throw "$Source endpoint is required"
  }
  if ($Endpoint -match '\s') {
    throw "$Source endpoint must not contain whitespace: $Endpoint"
  }

  try {
    $uri = [System.Uri]::new("tcp://$Endpoint")
  } catch {
    throw "$Source endpoint is invalid: $Endpoint"
  }

  if ([string]::IsNullOrWhiteSpace($uri.Host)) {
    throw "$Source endpoint host is required: $Endpoint"
  }
  if ($uri.Port -lt 1 -or $uri.Port -gt 65535) {
    throw "$Source endpoint port is out of range: $Endpoint"
  }
  if ($uri.UserInfo -or $uri.Query -or $uri.Fragment) {
    throw "$Source endpoint must be a bare host:port value: $Endpoint"
  }
}

function Assert-PeerEntry {
  param(
    [object]$Entry,
    [int]$Index,
    [string]$Source
  )

  if ($null -eq $Entry) {
    throw "$Source entry $Index is null"
  }
  if (-not $Entry.PSObject.Properties["node_id"] -or -not $Entry.PSObject.Properties["endpoint"]) {
    throw "$Source entry $Index must contain node_id and endpoint"
  }

  Assert-ValidNodeId -NodeId ([string]$Entry.node_id) -Source "$Source entry $Index"
  Assert-ValidEndpoint -Endpoint ([string]$Entry.endpoint) -Source "$Source entry $Index"
}

function Assert-SeedLine {
  param(
    [string]$Line,
    [int]$Index
  )

  if ($Line -notmatch '^\s*([0-9a-fA-F]{40})@(.+?)\s*$') {
    throw "seed line $Index must use node_id@host:port: $Line"
  }

  Assert-ValidNodeId -NodeId $Matches[1] -Source "seed line $Index"
  Assert-ValidEndpoint -Endpoint $Matches[2].Trim() -Source "seed line $Index"
}

$PeersPath = Resolve-RepoPath $PeersPath
$SeedsPath = Resolve-RepoPath $SeedsPath

if (-not (Test-Path -LiteralPath $PeersPath)) {
  throw "peer list JSON not found: $PeersPath"
}
if (-not (Test-Path -LiteralPath $SeedsPath)) {
  throw "seed list text not found: $SeedsPath"
}

$peerJson = Get-Content -Raw -LiteralPath $PeersPath | ConvertFrom-Json
if (-not $peerJson.PSObject.Properties["persistent_peers"]) {
  throw "peer list JSON must contain persistent_peers"
}

$peerEntries = @($peerJson.persistent_peers)
for ($i = 0; $i -lt $peerEntries.Count; $i++) {
  Assert-PeerEntry -Entry $peerEntries[$i] -Index ($i + 1) -Source "peer list JSON"
}

$seedLines = @(Get-Content -LiteralPath $SeedsPath | ForEach-Object { $_.Trim() } | Where-Object { -not [string]::IsNullOrWhiteSpace($_) -and -not $_.StartsWith("#") })
for ($i = 0; $i -lt $seedLines.Count; $i++) {
  Assert-SeedLine -Line $seedLines[$i] -Index ($i + 1)
}

[pscustomobject]@{
  peers = $peerEntries.Count
  seeds = $seedLines.Count
} | ConvertTo-Json -Depth 3
