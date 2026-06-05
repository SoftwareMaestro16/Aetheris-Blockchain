param(
  [string]$Doc = "docs\test-production-gates.md",
  [string]$PublicGates = "docs\public-testnet-production-gates.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$PublicGatesPath = if ([System.IO.Path]::IsPathRooted($PublicGates)) { $PublicGates } else { Join-Path $RepoRoot $PublicGates }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$publicText = Get-Content -Raw -LiteralPath $PublicGatesPath

foreach ($term in @(
    'Test And Production Gates',
    'Unit Tests',
    'fee formulas',
    'inflation formulas',
    'burn accounting',
    'staking ratio controller',
    'reputation scoring',
    'domain pricing',
    'resolver validation',
    'memo validation',
    'token/NFT/SBT storage rules',
    'Keeper Tests',
    '`x/fee`',
    '`x/token`',
    '`x/identity`',
    '`x/resolver`',
    '`x/reputation`',
    '`x/execution`',
    '`x/messaging`',
    '`x/queue`',
    '`x/events`',
    '`x/actors`',
    '`x/storage`',
    'Integration Tests',
    'bank transfer with memo',
    'resolver-based payment',
    'domain auction to ownership',
    'token creation and transfer',
    'NFT mint and transfer',
    'SBT mint and transfer rejection',
    'async contract call',
    'queue bounce/refund',
    'reputation rate limit',
    'dynamic fee under load',
    'E2E Smoke',
    '3-validator localnet',
    '5-validator localnet',
    'staking lifecycle',
    'fee distribution',
    'domain lifecycle',
    'AVM counter contract',
    'AFT token transfer',
    'ANFT mint/transfer',
    'ASBT mint/prove/revoke',
    'memo indexing',
    'restart persistence',
    'snapshot/state-sync',
    'Security Gates',
    '`go test ./...`',
    '`go vet ./...`',
    '`buf lint`',
    'deterministic execution gate',
    'state export/import gate',
    'govulncheck',
    'gosec',
    'gitleaks',
    'CodeQL',
    'dependency review',
    'independent audit before production claim',
    'Production Gate',
    'long-running public testnet has no untriaged consensus or fund-safety issues',
    'validator set can upgrade safely',
    'staking, fees, DEX, AVM, domains, reputation, memo, and contract standards',
    'state export/import is deterministic',
    'emergency governance and halt/restart process tested',
    'Immediate Build Order',
    'Finish base-chain rename, address policy, and `naet` cleanup',
    'Implement production fee formulas in `x/fee`',
    'Implement adaptive mint/burn controller in `x/token`',
    'Build deterministic async queue without AVM',
    'Build minimal AVM with a counter contract',
    'Add scheduler parallelism only after deterministic sequential async',
    'execution is stable',
    'Add compute/storage/market modules after baseline abuse controls exist',
    'Start partitioning/sharding simulator and spec only after async queue and',
    'AVM are audited',
    'Final Economic Architecture Summary',
    'staking participation -> adaptive inflation -> validator/delegator rewards',
    'network activity      -> burn             -> supply pressure reduction',
    'network load          -> soft fees/queues -> congestion control',
    'account behavior      -> reputation       -> anti-spam and priority',
    '`AET` has uncapped',
    'but bounded PoS supply',
    '`naet` remains the only protocol fee asset'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "test production gates doc missing: $term"
}

foreach ($term in @(
    'Test And Production Gates',
    'test-production-gates.md'
  )) {
  Assert-Contains -Text $publicText -Pattern ([regex]::Escape($term)) -Message "public gates missing Track 10 link: $term"
}

Write-Host "test production gates doc test passed"
