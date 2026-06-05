param(
  [string]$Doc = "docs\architecture\reputation-state.md",
  [string]$Boundaries = "docs\module-boundaries.md",
  [string]$Economy = "docs\architecture\economy-interop-architecture.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }
$EconomyPath = if ([System.IO.Path]::IsPathRooted($Economy)) { $Economy } else { Join-Path $RepoRoot $Economy }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath
$economyText = Get-Content -Raw -LiteralPath $EconomyPath

foreach ($term in @(
    'ReputationRecord',
    'score: uint8 // 0..100',
    'age_score',
    'staking_score',
    'tx_success_score',
    'volume_score',
    'domain_score',
    'contract_score',
    'spam_penalty',
    'failed_tx_penalty',
    'slash_penalty',
    '0-20   restricted',
    '20-50  new',
    '50-80  normal',
    '80-95  trusted',
    '95-100 elite',
    'score = clamp(score, 0, 100)',
    'Domain ownership can add bounded reputation',
    'cannot be directly',
    'deterministic on-chain events',
    'score -= inactivity_decay_rate * inactive_epochs',
    'New accounts have progressive limits',
    'Contracts also have reputation',
    'must not bypass `naet` fee validation',
    'low score means lower tx rate limit',
    'low score means lower async queue quota',
    'low score means higher memo/storage byte cost',
    'low score means stricter contract deploy limits',
    'high score can improve queue priority within deterministic bounds',
    'priority cannot bypass fees, signatures, or validation',
    'validators must compute identical priority ordering',
    'token creation may require score threshold or deposit',
    'contract deployment may require score threshold or deposit',
    'DEX pool creation may require score threshold or deposit',
    'domain auction spam can be rate-limited by score',
    'high-score users still must pay required protocol fees',
    'deposits can satisfy access gates but do not increase score directly',
    'failed contract executions add deterministic penalties',
    'successful contract executions can increase bounded contract reputation'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "reputation doc missing: $term"
}

foreach ($term in @(
    '`x/reputation`',
    'pure validation and scoring helpers only',
    'deterministic on-chain events',
    'no direct reputation purchase',
    'progressive limits',
    'Low score lowers tx rate limit and async queue quota',
    'High score may improve deterministic queue priority',
    'Token creation, contract deployment, and DEX pool creation',
    'Domain auction spam can be rate-limited by score',
    'Contract reputation updates'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing reputation: $term"
}

foreach ($term in @(
    'future reputation and scheduler modules',
    'reputation scores are deterministic',
    'no direct reputation purchase',
    'low reputation raises memo/storage byte cost',
    'high reputation may improve deterministic queue priority'
  )) {
  Assert-Contains -Text $economyText -Pattern ([regex]::Escape($term)) -Message "economy architecture missing reputation: $term"
}

Write-Host "reputation state doc test passed"
