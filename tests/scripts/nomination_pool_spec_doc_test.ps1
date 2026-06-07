param(
  [string]$Doc = "docs\architecture\nomination-pool-spec.md",
  [string]$Policy = "app\params\nomination_pool_spec.go",
  [string]$Tests = "app\params\nomination_pool_spec_test.go"
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
    'Nomination Pool Detailed Specification',
    'Nomination pools are important for accessibility',
    'centralization risks',
    'The implementation gate is `app/params/nomination_pool_spec.go`',
    '26. Nomination Pool Detailed Specification',
    '26.1 Pool Model',
    'Pool:',
    'PoolId',
    'OperatorAddress',
    'ValidatorAddress',
    'TotalBonded',
    'TotalShares',
    'CommissionBps',
    'Status',
    'CreatedHeight',
    'UnbondingEntries',
    'PoolDelegation:',
    'DelegatorAddress',
    'PrincipalEstimate',
    'RewardsAccrued',
    'AetraNominationPoolModuleName',
    'BuildAetraNominationPoolModelReport',
    'ValidateAetraNominationPoolModel',
    'accessibility',
    'deterministic accounting',
    'centralization risks',
    'Current Implementation Mapping',
    'PoolID',
    'PoolOperator',
    'ValidatorTarget',
    'TotalBondedStake',
    'PoolCommissionBps',
    'UnbondingQueue',
    'DelegatorShare',
    'PendingRewards',
    'TotalShares',
    'sum of all delegator shares',
    'deposits mint shares using deterministic integer math',
    'withdrawals burn shares and create unbonding entries',
    'slash losses must reduce pool bonded value without corrupting share supply',
    'export/import must preserve sorted pools, delegations, and unbonding entries',
    'no mandatory KYC should be embedded into consensus pool admission'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "nomination pool spec doc missing: $term"
}

foreach ($term in @(
    'AetraNominationPoolModuleName',
    'AetraNominationPoolModelEvidence',
    'AetraNominationPoolModelReport',
    'DefaultAetraNominationPoolModelEvidence',
    'ValidateAetraNominationPoolModel',
    'BuildAetraNominationPoolModelReport',
    'AetraNominationPoolFieldPoolID',
    'AetraNominationPoolFieldOperatorAddress',
    'AetraNominationPoolFieldValidatorAddress',
    'AetraNominationPoolFieldTotalBonded',
    'AetraNominationPoolFieldTotalShares',
    'AetraNominationPoolFieldCommissionBps',
    'AetraNominationPoolFieldStatus',
    'AetraNominationPoolFieldCreatedHeight',
    'AetraNominationPoolFieldUnbondingEntries',
    'AetraNominationPoolFieldDelegatorAddress',
    'AetraNominationPoolFieldDelegationPoolID',
    'AetraNominationPoolFieldShares',
    'AetraNominationPoolFieldPrincipalEstimate',
    'AetraNominationPoolFieldRewardsAccrued',
    'AetraNominationPoolRiskAccessibility',
    'AetraNominationPoolRiskAccounting',
    'AetraNominationPoolRiskCentralization',
    'AetraNominationPoolImplementationMap'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "nomination pool spec policy missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraNominationPoolModelCoversSection261',
    'TestAetraNominationPoolModelRejectsMissingRequiredFieldsAndRisks',
    'TestAetraNominationPoolModelRejectsDuplicateUnexpectedAndWrongModule',
    'module_name_required',
    'CreatedHeight',
    'PrincipalEstimate',
    'OperatorKycStatus',
    'LocalUiEstimate'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "nomination pool spec tests missing: $term"
}

Write-Host "nomination pool spec doc test passed"
