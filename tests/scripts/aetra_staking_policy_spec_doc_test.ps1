param(
  [string]$Doc = "docs\architecture\aetra-staking-policy-spec.md",
  [string]$Policy = "app\params\aetra_staking_policy_spec.go",
  [string]$Tests = "app\params\aetra_staking_policy_spec_test.go"
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
    'x/aetra-staking-policy Module Specification',
    'Purpose: control effective voting power, delegation overflow, commission policy, and anti-concentration incentives.',
    'This module is the central anti-centralization module of Aetra.',
    'calculate raw validator stake',
    'calculate effective validator stake',
    'calculate overflow stake',
    'enforce or expose effective voting power cap',
    'calculate reward multiplier for overflow stake',
    'expose delegation concentration warnings',
    'enforce commission floor',
    'enforce max commission',
    'enforce max commission change rate',
    'expose top-N concentration metrics',
    'validate governance param changes',
    'emit events for cap/overflow/commission policy changes',
    'remain deterministic and export/import safe',
    'staking policy = stake math + cap enforcement/exposure + overflow accounting + commission policy + concentration metrics + governance params + events + export/import safety + tests + docs',
    'The implementation gate is `app/params/aetra_staking_policy_spec.go`',
    '`AetraStakingPolicyModuleName` must be `x/aetra-staking-policy`',
    'wrong or missing module identity is rejected'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "aetra staking policy spec doc missing: $term"
}

foreach ($term in @(
    'AetraStakingPolicyModuleName',
    'AetraStakingPolicySpecEvidence',
    'AetraStakingPolicySpecReport',
    'DefaultAetraStakingPolicySpecEvidence',
    'ValidateAetraStakingPolicySpec',
    'BuildAetraStakingPolicySpecReport',
    'AetraStakingPolicyPurposeEffectivePowerOverflowCommissionAntiConcentration',
    'AetraStakingPolicyCentralAntiCentralizationModule',
    'AetraStakingPolicyResponsibilityRawStake',
    'AetraStakingPolicyResponsibilityEffectiveStake',
    'AetraStakingPolicyResponsibilityOverflowStake',
    'AetraStakingPolicyResponsibilityEffectiveVotingPowerCap',
    'AetraStakingPolicyResponsibilityOverflowRewardMultiplier',
    'AetraStakingPolicyResponsibilityDelegationConcentrationWarning',
    'AetraStakingPolicyResponsibilityCommissionFloor',
    'AetraStakingPolicyResponsibilityMaxCommission',
    'AetraStakingPolicyResponsibilityMaxCommissionChangeRate',
    'AetraStakingPolicyResponsibilityTopNConcentrationMetrics',
    'AetraStakingPolicyResponsibilityGovernanceParamValidation',
    'AetraStakingPolicyResponsibilityPolicyChangeEvents',
    'AetraStakingPolicyResponsibilityDeterministicExportImport'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "aetra staking policy spec policy missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraStakingPolicySpecCoversResponsibilities',
    'TestAetraStakingPolicySpecRejectsMissingStakeAndPowerResponsibilities',
    'TestAetraStakingPolicySpecRejectsMissingCommissionConcentrationAndSafetyResponsibilities',
    'TestAetraStakingPolicySpecRejectsMissingPurposeAndCentralModuleRole',
    'TestAetraStakingPolicySpecRejectsWrongModuleIdentity'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "aetra staking policy spec tests missing: $term"
}

Write-Host "aetra staking policy spec doc test passed"
