param(
  [string]$Doc = "docs\architecture\aetra-economics-spec.md",
  [string]$Policy = "app\params\aetra_economics_spec.go",
  [string]$Tests = "app\params\aetra_economics_spec_test.go"
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
    'x/aetra-economics Module Specification',
    'Purpose: low/moderate inflation, fee burn, treasury allocation, reward smoothing, and transparent APR model.',
    'economic-control module of Aetra',
    'implement low/moderate inflation',
    'implement fee burn',
    'implement treasury allocation',
    'implement reward smoothing',
    'expose a transparent APR model',
    'The implementation gate is `app/params/aetra_economics_spec.go`',
    '`AetraEconomicsModuleName` must be `x/aetra-economics`',
    'wrong or missing module identity must fail validation',
    '23.1 Responsibilities',
    'calculate dynamic inflation',
    'track bonded ratio',
    'estimate staking APR',
    'split fees',
    'burn configured fee share',
    'send configured share to distribution/rewards',
    'send configured share to treasury',
    'smooth reward changes',
    'expose economic metrics',
    'protect supply invariants',
    'BuildAetraEconomicsResponsibilitiesReport',
    '23.2 State',
    'Params:',
    'InflationMinBps',
    'InflationMaxBps',
    'InflationChangeRateBps',
    'TargetBondedRatioBps',
    'BurnFeeShareBps',
    'RewardFeeShareBps',
    'TreasuryFeeShareBps',
    'RewardSmoothingEpochs',
    'AprTargetMinBps',
    'AprTargetMaxBps',
    'EpochEconomics:',
    'EpochNumber',
    'StartHeight',
    'EndHeight',
    'BondedRatioBps',
    'InflationBps',
    'EstimatedAprBps',
    'FeesCollected',
    'FeesBurned',
    'FeesToRewards',
    'FeesToTreasury',
    'MintedRewards',
    'SupplyStats:',
    'TotalMinted',
    'TotalBurned',
    'NetIssuance',
    'BuildAetraEconomicsStateSpecReport'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "aetra economics spec doc missing: $term"
}

foreach ($term in @(
    'AetraEconomicsModuleName',
    'AetraEconomicsSpecEvidence',
    'AetraEconomicsSpecReport',
    'DefaultAetraEconomicsSpecEvidence',
    'ValidateAetraEconomicsSpec',
    'BuildAetraEconomicsSpecReport',
    'AetraEconomicsPurposeLowModerateInflation',
    'AetraEconomicsPurposeFeeBurn',
    'AetraEconomicsPurposeTreasuryAllocation',
    'AetraEconomicsPurposeRewardSmoothing',
    'AetraEconomicsPurposeTransparentAPRModel',
    'AetraEconomicsResponsibilitiesEvidence',
    'AetraEconomicsResponsibilitiesReport',
    'DefaultAetraEconomicsResponsibilitiesEvidence',
    'ValidateAetraEconomicsResponsibilities',
    'BuildAetraEconomicsResponsibilitiesReport',
    'AetraEconomicsResponsibilityDynamicInflation',
    'AetraEconomicsResponsibilityBondedRatio',
    'AetraEconomicsResponsibilityStakingAPR',
    'AetraEconomicsResponsibilityFeeSplit',
    'AetraEconomicsResponsibilityBurnFeeShare',
    'AetraEconomicsResponsibilityRewardsShare',
    'AetraEconomicsResponsibilityTreasuryShare',
    'AetraEconomicsResponsibilityRewardSmoothing',
    'AetraEconomicsResponsibilityEconomicMetrics',
    'AetraEconomicsResponsibilitySupplyInvariants',
    'AetraEconomicsStateSpecEvidence',
    'AetraEconomicsStateSpecReport',
    'DefaultAetraEconomicsStateSpecEvidence',
    'ValidateAetraEconomicsStateSpec',
    'BuildAetraEconomicsStateSpecReport',
    'AetraEconomicsStateParams',
    'AetraEconomicsStateEpochEconomics',
    'AetraEconomicsStateSupplyStats',
    'AetraEconomicsStateParamInflationMinBps',
    'AetraEconomicsStateParamInflationMaxBps',
    'AetraEconomicsStateParamInflationChangeRateBps',
    'AetraEconomicsStateParamTargetBondedRatioBps',
    'AetraEconomicsStateParamBurnFeeShareBps',
    'AetraEconomicsStateParamRewardFeeShareBps',
    'AetraEconomicsStateParamTreasuryFeeShareBps',
    'AetraEconomicsStateParamRewardSmoothingEpochs',
    'AetraEconomicsStateParamAprTargetMinBps',
    'AetraEconomicsStateParamAprTargetMaxBps',
    'AetraEconomicsStateEpochNumber',
    'AetraEconomicsStateEpochStartHeight',
    'AetraEconomicsStateEpochEndHeight',
    'AetraEconomicsStateEpochBondedRatioBps',
    'AetraEconomicsStateEpochInflationBps',
    'AetraEconomicsStateEpochEstimatedAprBps',
    'AetraEconomicsStateEpochFeesCollected',
    'AetraEconomicsStateEpochFeesBurned',
    'AetraEconomicsStateEpochFeesToRewards',
    'AetraEconomicsStateEpochFeesToTreasury',
    'AetraEconomicsStateEpochMintedRewards',
    'AetraEconomicsStateSupplyTotalMinted',
    'AetraEconomicsStateSupplyTotalBurned',
    'AetraEconomicsStateSupplyNetIssuance'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "aetra economics spec policy missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraEconomicsSpecCoversModulePurpose',
    'TestAetraEconomicsSpecRejectsMissingPurposeComponents',
    'TestAetraEconomicsSpecRejectsWrongModuleIdentity',
    'TestDefaultAetraEconomicsResponsibilitiesCoverSection231',
    'TestAetraEconomicsResponsibilitiesRejectMissingRequiredItems',
    'TestDefaultAetraEconomicsStateSpecCoversSection232',
    'TestAetraEconomicsStateSpecRejectsMissingFields',
    'TestAetraEconomicsStateSpecRejectsDuplicateUnexpectedAndWrongModule'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "aetra economics spec tests missing: $term"
}

Write-Host "aetra economics spec doc test passed"
