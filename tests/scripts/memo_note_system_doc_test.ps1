param(
  [string]$Doc = "docs\architecture\memo-note-system.md",
  [string]$Boundaries = "docs\module-boundaries.md",
  [string]$ProtocolFees = "docs\protocol-fees.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }
$ProtocolFeesPath = if ([System.IO.Path]::IsPathRooted($ProtocolFees)) { $ProtocolFees } else { Join-Path $RepoRoot $ProtocolFees }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath
$feesText = Get-Content -Raw -LiteralPath $ProtocolFeesPath

foreach ($term in @(
    'Memo / Note System',
    'TxMetadata',
    'memo: string optional',
    'memo_hash: bytes optional',
    'memo_visible: bool',
    'native bank transfer',
    'resolver/domain payment',
    'token transfer',
    'NFT transfer',
    'SBT proof/revoke',
    'contract call',
    'domain auction bid',
    'domain renewal',
    'DEX swap/liquidity action',
    'UTF-8 only',
    '`200` characters',
    'hard max memo chars: `500`',
    'default max memo bytes: `1024`',
    'immutable after block inclusion',
    'stored as transaction metadata, not execution input',
    'cannot change state transition result',
    'require valid_utf8(memo)',
    'require char_count(memo) <= max_memo_chars',
    'require byte_len(memo) <= max_memo_bytes',
    'require no prohibited control chars',
    '32-byte SHA-256',
    'memo_fee',
    'reputation_multiplier(sender)',
    'congestion_multiplier(load)',
    'score >= 80: 0.75',
    'score >= 50: 1.00',
    'score >= 20: 1.50',
    'score < 20:  3.00',
    'memo fee paid only in `naet`',
    'memo fee can be zero for empty memo',
    'memo size contributes to tx byte cost',
    'memo cannot become cheap spam storage',
    '`x/memo/types`',
    'Storage And Indexing',
    'tx hash',
    'sender',
    'receiver if known',
    'asset type',
    'related domain if any',
    'memo or memo hash depending on storage policy',
    'block height',
    'timestamp',
    'by tx hash',
    'by sender',
    'by receiver',
    'by domain',
    'by contract',
    'by asset',
    'by event type',
    'full memo stored on-chain',
    'store memo hash on-chain',
    'store bounded memo on-chain only if below configured byte limit',
    'consensus does not depend on search index',
    'EventMemoAttached',
    'memo_hash',
    'Low reputation memo cost is higher'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "memo doc missing: $term"
}

foreach ($term in @(
    '`x/memo`',
    'optional human-readable transaction metadata',
    'UTF-8 only',
    'hard protocol bound',
    'does not affect execution state transitions',
    'memo fees are paid only in `naet`',
    'Memo projection may index by tx hash',
    'Full memo on-chain and hash-only on-chain storage policies',
    'Consensus does not depend on search index results',
    '`EventMemoAttached` is deterministic'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing memo: $term"
}

foreach ($term in @(
    'memo fee',
    'paid only in `naet`',
    'can be zero for an empty memo',
    'memo bytes contribute to tx byte cost'
  )) {
  Assert-Contains -Text $feesText -Pattern ([regex]::Escape($term)) -Message "protocol fees missing memo: $term"
}

Write-Host "memo note system doc test passed"
