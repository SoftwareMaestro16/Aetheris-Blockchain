param(
  [string]$Doc = "docs\architecture\initial-release-non-goals.md",
  [string]$Policy = "app\params\initial_release_non_goals.go",
  [string]$Tests = "app\params\initial_release_non_goals_test.go"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) { return $Path }
  return Join-Path $RepoRoot $Path
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Doc)
$policyText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Policy)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)

foreach ($term in @(
    'Initial Release Non-Goals',
    'Initial release should not attempt',
    'PoH',
    'Solana-level TPS',
    '1-second blocks',
    'mandatory KYC validator admission',
    'EVM at genesis unless separately approved',
    'subjective slashing',
    'unlimited validator set',
    'unbounded contract execution',
    'high inflation APR marketing',
    'hard scope boundaries',
    'consensus on CometBFT BFT without PoH',
    'bounded contract execution',
    'APR and inflation messaging conservative and evidence-based'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "initial release non-goals doc missing: $term"
}

foreach ($term in @(
    'InitialReleaseScopePolicy',
    'InitialReleaseNonGoalsReport',
    'DefaultInitialReleaseScopePolicy',
    'ValidateInitialReleaseScope',
    'BuildInitialReleaseNonGoalsReport',
    'InitialReleaseNonGoalPoH',
    'InitialReleaseNonGoalSolanaLevelTPS',
    'InitialReleaseNonGoalOneSecondBlocks',
    'InitialReleaseNonGoalMandatoryKYC',
    'InitialReleaseNonGoalEVMAtGenesis',
    'InitialReleaseNonGoalSubjectiveSlashing',
    'InitialReleaseNonGoalUnlimitedValidatorSet',
    'InitialReleaseNonGoalUnboundedContractExecution',
    'InitialReleaseNonGoalHighInflationAPRMarketing'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "initial release non-goals policy missing: $term"
}

foreach ($term in @(
    'TestDefaultInitialReleaseScopeRespectsNonGoals',
    'TestInitialReleaseScopeRejectsPerformanceAndValidatorScopeCreep',
    'TestInitialReleaseScopeRejectsExecutionSecurityAndMarketingScopeCreep'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "initial release non-goals tests missing: $term"
}

Write-Host "initial release non-goals doc test passed"
