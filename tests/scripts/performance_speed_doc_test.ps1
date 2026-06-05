param(
  [string]$Doc = "docs\performance-speed.md",
  [string]$StartScript = "scripts\localnet\start.ps1",
  [string]$AdversarialWorkflow = ".github\workflows\adversarial-e2e.yml",
  [string]$PrototypeWorkflow = ".github\workflows\prototype-release.yml"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Path))
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Doc)
$startText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $StartScript)
$adversarialText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $AdversarialWorkflow)
$prototypeText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $PrototypeWorkflow)

foreach ($term in @(
    "Performance And Speed",
    "measurement-first",
    "must not bypass signer, fee, denom, zero-address, authority, or genesis validation",
    "BenchmarkParseRawAddress",
    "BenchmarkParseUserFriendlyAddress",
    "BenchmarkValidateFeeCoinsAllowedNaet",
    "BenchmarkDexSwapExactAmountIn",
    "BenchmarkTokenfactoryCreateDenom",
    "BenchmarkQueueProcessBlock",
    "BenchmarkContractStateExportImport",
    "BenchmarkTokenMasterMint",
    "BenchmarkTokenWalletTransfer",
    "BenchmarkNFTMint",
    "BenchmarkNFTTransfer",
    "BenchmarkSBTProofAndRevoke",
    "BenchmarkWalletSignedSend",
    "startup-timing.json",
    "CI Timing Summary",
    "ConvertFrom-Json",
    "protobuf tx",
    "DefaultQueryDenoms",
    "MaxQueryDenoms",
    "DefaultQueryPools",
    "MaxQueryPools",
    "Future token, NFT, contract, and async indexes must add default and max query limits"
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "performance doc missing term: $term"
}

foreach ($term in @(
    "startup-timing.json",
    "Startup timing:",
    "total_ms",
    "process_launch_ms",
    "health_wait_ms",
    "System.Diagnostics.Stopwatch"
  )) {
  Assert-Contains -Text $startText -Pattern ([regex]::Escape($term)) -Message "localnet start script missing timing term: $term"
}

foreach ($term in @(
    "Measure-Command",
    "CI Timing Summary",
    "adversarial localnet smoke",
    "scaled localnet smoke",
    "GITHUB_STEP_SUMMARY"
  )) {
  Assert-Contains -Text $adversarialText -Pattern ([regex]::Escape($term)) -Message "adversarial workflow missing timing term: $term"
}

foreach ($term in @(
    "Measure-Command",
    "CI Timing Summary",
    "prototype acceptance smoke",
    "GITHUB_STEP_SUMMARY"
  )) {
  Assert-Contains -Text $prototypeText -Pattern ([regex]::Escape($term)) -Message "prototype workflow missing timing term: $term"
}

Write-Host "performance speed doc test passed"
