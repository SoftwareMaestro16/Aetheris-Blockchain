param(
  [string]$Doc = "docs\architecture\implementation-phases.md",
  [string]$Policy = "app\params\implementation_phases.go",
  [string]$Tests = "app\params\implementation_phases_test.go"
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
    'Implementation Phases',
    'Phase 0 - Baseline Audit',
    'inspect current Cosmos SDK and CometBFT versions',
    'document current app module graph',
    'identify existing modules overlapping with `aetra-staking-policy`, `aetra-validator-score`, and `aetra-economics`',
    'decide which modules are renamed, reused, or wrapped',
    'verify current `naet` staking denom',
    'verify fee collector, burn, treasury, emissions, mint authority wiring',
    'verify current localnet scripts and test coverage',
    'module inventory',
    'gap analysis',
    'risk list',
    'updated implementation checklist',
    'current full unit test run',
    'current integration test run',
    'current localnet smoke test',
    'current export/import test',
    'Phase 1 - Staking Policy and Validator Cap',
    'implement effective voting power cap',
    'implement overflow stake accounting',
    'implement commission floor/max/max-change policy',
    'add concentration metrics',
    'add queries for validator raw/effective/overflow stake',
    'add governance params with validation',
    'wire module into app lifecycle',
    'cap math unit tests',
    'validator set transition tests',
    'concentration query tests',
    'commission bounds tests',
    'integration tests with staking',
    'export/import tests',
    'invariant tests',
    'no validator can exceed configured effective power cap',
    'excess stake does not increase voting power',
    'params cannot be set outside safe bounds',
    'state remains deterministic after export/import'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "implementation phases doc missing: $term"
}

foreach ($term in @(
    'ImplementationPhaseBaselineAudit',
    'ImplementationPhaseStakingPolicyCap',
    'ImplementationPhaseItem',
    'ImplementationPhasePlan',
    'ImplementationPhaseReport',
    'DefaultImplementationPhasePlans',
    'ValidateImplementationPhasePlan',
    'BuildImplementationPhaseReport',
    'PhaseTaskInspectVersions',
    'PhaseTaskDocumentModuleGraph',
    'PhaseTaskIdentifyOverlappingModules',
    'PhaseTaskDecideRenameReuseWrap',
    'PhaseTaskVerifyNaetStakingDenom',
    'PhaseTaskVerifyEconomyWiring',
    'PhaseTaskVerifyLocalnetAndCoverage',
    'PhaseDeliverableModuleInventory',
    'PhaseDeliverableGapAnalysis',
    'PhaseDeliverableRiskList',
    'PhaseDeliverableImplementationChecklist',
    'PhaseTaskImplementEffectivePowerCap',
    'PhaseTaskImplementOverflowAccounting',
    'PhaseTaskImplementCommissionPolicy',
    'PhaseTaskAddConcentrationMetrics',
    'PhaseTaskAddStakeQueries',
    'PhaseTaskAddGovernanceParams',
    'PhaseTaskWireModuleLifecycle',
    'PhaseAcceptanceNoValidatorExceedsCap',
    'PhaseAcceptanceExcessNoVotingPower',
    'PhaseAcceptanceParamsSafeBounds',
    'PhaseAcceptanceDeterministicExportImport'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "implementation phases policy missing: $term"
}

foreach ($term in @(
    'TestDefaultImplementationPhasePlansCoverPhase0AndPhase1',
    'TestImplementationPhaseRejectsMissingEvidence',
    'TestImplementationPhaseRejectsMissingRequiredItem',
    'TestImplementationPhaseRejectsUnknownPhaseAndUnexpectedItem'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "implementation phases tests missing: $term"
}

Write-Host "implementation phases doc test passed"
