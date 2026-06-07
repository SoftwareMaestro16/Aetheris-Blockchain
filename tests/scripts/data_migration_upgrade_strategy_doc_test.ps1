param(
  [string]$Doc = "docs\architecture\data-migration-upgrade-strategy.md",
  [string]$Catalog = "app\params\data_migration_upgrade_strategy.go",
  [string]$Tests = "app\params\data_migration_upgrade_strategy_test.go"
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
$catalogText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Catalog)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)

foreach ($term in @(
    'Data Migration and Upgrade Strategy',
    'Aetra is expected to evolve. Upgrade safety is part of the architecture.',
    'module consensus versions are explicit',
    'migrations are registered before public testnet',
    'deterministic state transformations',
    'validates old state before writing new state',
    'validates new state after writing',
    'export/import after upgrade must produce valid genesis',
    'app hash after restart must remain stable',
    'dry-run upgrade procedure must be documented',
    'rollback limits must be documented',
    'unsafe downgrade behavior must be rejected',
    'upgrade handlers must emit stable events',
    'tests must cover migration success and failure paths',
    '31.1 Upgrade Requirements',
    'Every new module or state-breaking change must include',
    'store key decision',
    'genesis import/export',
    'migration handler',
    'version map update',
    'upgrade test',
    'rollback notes where possible',
    'operator instructions',
    'The store key decision must state whether the module introduces a new store key',
    'Migration handlers must be registered in the upgrade path',
    'Rollback notes are required even when rollback is limited',
    'DefaultAetraUpgradeStrategyEvidence',
    '31.2 Migration Tests',
    'Required tests:',
    'old genesis imports into new binary',
    'migration initializes params',
    'migration preserves balances',
    'migration preserves staking state',
    'migration preserves slashing state',
    'migration preserves contract state if applicable',
    'app hash after migration is deterministic',
    'Manual notes are not enough for balances, staking, slashing, contract state, or app hash determinism',
    'current consensus version',
    'migration handlers from every supported previous version',
    'version map sanity checks',
    'genesis validation for migrated state',
    'export/import compatibility tests',
    'deterministic replay tests',
    'invariant tests after migration',
    'upgrade name',
    'target height',
    'binary version',
    'expected module version map before upgrade',
    'expected module version map after upgrade',
    'operator runbook',
    'post-upgrade smoke tests',
    'public rollback boundary',
    'missing module version rejected',
    'future module version rejected',
    'app hash stability after migration',
    'event emission for upgrade handler execution',
    'BuildAetraUpgradeStrategyReport'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "data migration upgrade strategy doc missing: $term"
}

foreach ($term in @(
    'AetraUpgradeRequirementStoreKeyDecision',
    'AetraUpgradeRequirementGenesisImportExport',
    'AetraUpgradeRequirementMigrationHandler',
    'AetraUpgradeRequirementVersionMapUpdate',
    'AetraUpgradeRequirementUpgradeTest',
    'AetraUpgradeRequirementRollbackNotes',
    'AetraUpgradeRequirementOperatorInstructions',
    'AetraMigrationTestOldGenesisImports',
    'AetraMigrationTestInitializesParams',
    'AetraMigrationTestPreservesBalances',
    'AetraMigrationTestPreservesStakingState',
    'AetraMigrationTestPreservesSlashingState',
    'AetraMigrationTestPreservesContractState',
    'AetraMigrationTestDeterministicAppHash',
    'AetraUpgradeStrategyEvidence',
    'AetraUpgradeStrategyReport',
    'DefaultAetraUpgradeStrategyEvidence',
    'ValidateAetraUpgradeStrategy',
    'BuildAetraUpgradeStrategyReport',
    'RequiredAetraUpgradeRequirements',
    'RequiredAetraMigrationTests'
  )) {
  Assert-Contains -Text $catalogText -Pattern ([regex]::Escape($term)) -Message "data migration upgrade strategy catalog missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraUpgradeStrategyCoversSection31',
    'TestAetraUpgradeStrategyRejectsMissingRequirementsAndTests'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "data migration upgrade strategy tests missing: $term"
}

Write-Host "data migration upgrade strategy doc test passed"
