param(
  [string]$Doc = "docs\security\security-risks-controls.md",
  [string]$AuditPack = "docs\security\security-audit-pack.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$AuditPackPath = if ([System.IO.Path]::IsPathRooted($AuditPack)) { $AuditPack } else { Join-Path $RepoRoot $AuditPack }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$auditText = Get-Content -Raw -LiteralPath $AuditPackPath

foreach ($term in @(
    'Security Risks And Controls',
    'Infinite Supply Risks',
    'excessive inflation',
    'validator reward dilution',
    'weak token confidence',
    'governance abuse',
    'hard inflation cap',
    'target staking controller',
    'public mint/burn telemetry',
    'governance bounds',
    'export/import supply invariants',
    'Deflation Risks',
    'burn exceeds mint for too long',
    'validator income falls',
    'users hoard instead of transacting',
    'tx fees become politically hard to lower',
    'burn ratio floor/ceiling',
    'validator baseline mint rewards',
    'fee caps',
    'deflation guard',
    'Spam Risks',
    'low-cost tx floods',
    'memo spam',
    'async queue flooding',
    'contract deploy spam',
    'domain auction spam',
    'reputation rate limits',
    'per-account queue caps',
    'per-contract queue caps',
    'memo byte fees',
    'deploy deposits',
    'domain auction bid deposits',
    'bounded block processing',
    'scheduler fairness',
    'Staking Attacks',
    'stake concentration',
    'validator cartel',
    'long-range attack',
    'validator downtime',
    'double-sign',
    'unbonding period',
    'slashing',
    'tombstone for severe equivocation if enabled',
    'validator concentration alerts',
    'commission bounds',
    'delegation transparency',
    'snapshot/state-sync safety',
    'Economic Attacks',
    'fee market manipulation',
    'fake volume to trigger burn',
    'reputation farming',
    'domain squatting',
    'token metadata spoofing',
    'bridge asset spoofing',
    'fee multiplier smoothing',
    'bounded burn response',
    'deterministic reputation decay',
    'auction pricing and renewal',
    'native token metadata reservation',
    'bridged asset namespace isolation',
    'Audit Gate',
    'fee cap bypass or non-`naet` fee payment'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "security risks controls doc missing: $term"
}

foreach ($term in @(
    'Security Risks And Controls',
    'security-risks-controls.md'
  )) {
  Assert-Contains -Text $auditText -Pattern ([regex]::Escape($term)) -Message "security audit pack missing risks controls link: $term"
}

Write-Host "security risks controls doc test passed"
