param(
  [string]$Doc = "docs\architecture\observability-public-metrics.md",
  [string]$Metrics = "observability\metrics.go",
  [string]$Catalog = "observability\public_metrics.go",
  [string]$Tests = "observability\public_metrics_test.go"
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
$metricsText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Metrics)
$catalogText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Catalog)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)

foreach ($term in @(
    'Observability and Public Metrics',
    'block time',
    'finality latency',
    'missed blocks',
    'validator uptime',
    'validator concentration',
    'top-10/top-20/top-33 voting power',
    'inflation',
    'bonded ratio',
    'estimated APR',
    'burned fees',
    'treasury balance',
    'slashing events',
    'jail/unjail events',
    'contract execution gas',
    'failed tx reasons',
    'node sync status',
    'CLI queries',
    'gRPC queries',
    'REST queries where applicable',
    'Prometheus metrics',
    'explorer/indexer compatibility events',
    'public testnet dashboards'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "observability public metrics doc missing: $term"
}

foreach ($term in @(
    'MetricFinalityLatencySeconds',
    'MetricFailedTxReasons',
    'MetricEconomyBondedRatioBps',
    'MetricEconomyEstimatedAPRBps',
    'MetricEconomyBurnedFeesNaet',
    'MetricEconomyTreasuryBalanceNaet',
    'MetricSlashingEventsTotal',
    'MetricValidatorJailEventsTotal',
    'MetricValidatorUnjailEventsTotal',
    'MetricValidatorMissedBlocks',
    'MetricValidatorUptimeBps',
    'MetricValidatorConcentrationBps',
    'MetricContractExecutionGas',
    'MetricNodeSyncStatus'
  )) {
  Assert-Contains -Text $metricsText -Pattern ([regex]::Escape($term)) -Message "observability metrics missing: $term"
}

foreach ($term in @(
    'PublicMetricSpec',
    'PublicSurfaceSpec',
    'PublicMetricsReadinessReport',
    'DefaultPublicMetricSpecs',
    'DefaultPublicSurfaceSpecs',
    'ValidatePublicMetricsReadiness',
    'RequiredMetricBlockTime',
    'RequiredMetricFinalityLatency',
    'RequiredMetricMissedBlocks',
    'RequiredMetricValidatorUptime',
    'RequiredMetricValidatorConcentration',
    'RequiredMetricTopNVotingPower',
    'RequiredMetricInflation',
    'RequiredMetricBondedRatio',
    'RequiredMetricEstimatedAPR',
    'RequiredMetricBurnedFees',
    'RequiredMetricTreasuryBalance',
    'RequiredMetricSlashingEvents',
    'RequiredMetricJailUnjailEvents',
    'RequiredMetricContractExecutionGas',
    'RequiredMetricFailedTxReasons',
    'RequiredMetricNodeSyncStatus',
    'RequiredSurfaceCLIQueries',
    'RequiredSurfaceGRPCQueries',
    'RequiredSurfaceRESTQueries',
    'RequiredSurfacePrometheusMetrics',
    'RequiredSurfaceIndexerEvents',
    'RequiredSurfacePublicDashboards'
  )) {
  Assert-Contains -Text $catalogText -Pattern ([regex]::Escape($term)) -Message "public metrics catalog missing: $term"
}

foreach ($term in @(
    'TestDefaultPublicMetricsCoverRequiredSection14Metrics',
    'TestPublicMetricsRejectMissingRequiredMetricSurface',
    'TestPublicMetricsRejectPrometheusMetricNotInRegistry',
    'TestPublicMetricsRejectMissingRequiredSurface',
    'TestPublicMetricsRejectPrometheusOnlyExposure'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "public metrics tests missing: $term"
}

Write-Host "observability public metrics doc test passed"
