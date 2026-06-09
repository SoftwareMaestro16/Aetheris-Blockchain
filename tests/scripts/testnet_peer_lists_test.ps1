param(
  [string]$ValidatorScript = "scripts\testnet\validate-peer-lists.ps1"
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

function Assert-True {
  param([bool]$Condition, [string]$Message)
  if (-not $Condition) { throw $Message }
}

function Assert-Fails {
  param(
    [scriptblock]$Script,
    [string]$Pattern
  )

  try {
    & $Script | Out-Null
  } catch {
    if ($_.Exception.Message -notmatch $Pattern) {
      throw "expected failure matching '$Pattern', got: $($_.Exception.Message)"
    }
    return
  }
  throw "expected command to fail: $Pattern"
}

$validatorScriptPath = Resolve-RepoPath $ValidatorScript
& $validatorScriptPath | Out-Null

$workRoot = Resolve-RepoPath ".work\peer-list-test"
if (Test-Path -LiteralPath $workRoot) {
  Remove-Item -LiteralPath $workRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $workRoot | Out-Null

$goodSeeds = Join-Path $workRoot "seeds.good.txt"
@'
0123456789abcdef0123456789abcdef01234567@seed-1.aetra.example:26656
'@ | Set-Content -LiteralPath $goodSeeds

$badNodeIdJson = Join-Path $workRoot "peers.bad-node-id.json"
@'
{
  "persistent_peers": [
    {
      "node_id": "bad-node-id",
      "endpoint": "seed-1.aetra.example:26656"
    }
  ]
}
'@ | Set-Content -LiteralPath $badNodeIdJson

$badEndpointJson = Join-Path $workRoot "peers.bad-endpoint.json"
@'
{
  "persistent_peers": [
    {
      "node_id": "0123456789abcdef0123456789abcdef01234567",
      "endpoint": "seed-1.aetra.example"
    }
  ]
}
'@ | Set-Content -LiteralPath $badEndpointJson

$badSeedLine = Join-Path $workRoot "seeds.bad-line.txt"
@'
0123456789abcdef0123456789abcdef01234567@seed-1.aetra.example
'@ | Set-Content -LiteralPath $badSeedLine

& $validatorScriptPath `
  -PeersPath "docs\testnet\peers.example.json" `
  -SeedsPath "docs\testnet\seeds.example.txt" | Out-Null

Assert-Fails {
  & $validatorScriptPath -PeersPath $badNodeIdJson -SeedsPath $goodSeeds
} "node id"

Assert-Fails {
  & $validatorScriptPath -PeersPath $badEndpointJson -SeedsPath $goodSeeds
} "endpoint"

Assert-Fails {
  & $validatorScriptPath -PeersPath "docs\testnet\peers.example.json" -SeedsPath $badSeedLine
} "seed line"

Write-Host "testnet peer list validation test passed"
