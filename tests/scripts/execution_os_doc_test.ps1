param(
  [string]$Doc = "docs\architecture\execution-os.md",
  [string]$Boundaries = "docs\module-boundaries.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath

foreach ($term in @(
    'Execution OS',
    'x/execution',
    'transaction orchestration',
    'execution pipeline',
    'module dispatch',
    'async entrypoint',
    'deterministic ordering',
    'event collection',
    'error handling',
    'CheckTx:',
    'validate signatures',
    'validate fees',
    'validate memo',
    'DeliverTx/FinalizeBlock:',
    'ExecutionEnvelope',
    'optional memo metadata',
    'resolver lookup',
    'reputation limits',
    'dynamic fee estimator',
    'deterministic execution trace',
    'x/vm',
    'AVM runtime routing',
    'gated CosmWasm runtime routing',
    'counter contract',
    'external call',
    'internal call',
    'bounced call',
    'query/getter',
    'bytecode/module format',
    'deterministic host functions',
    'CosmWasm remains disabled by default',
    'x/messaging',
    'async calls between contracts',
    'internal message envelope',
    'bounce/refund behavior',
    'outgoing message validation',
    'value_naet',
    'created_lt',
    'deterministic message id',
    'refund/no-double-spend behavior',
    'x/queue',
    'delayed execution',
    'scheduled tasks',
    'retry and failure handling',
    'queue limits',
    'queue observability',
    'scheduled_height',
    'reputation_class',
    'source_logical_time',
    'low reputation cannot starve forever',
    'max per-account queued messages',
    'max per-contract queued messages',
    'x/events',
    'event-driven system',
    'protocol events',
    'indexer events',
    'contract events',
    'memo events',
    'domain events',
    'reputation events',
    '`EventTransfer`',
    '`EventMemoAttached`',
    '`EventDomainAuctionStarted`',
    '`EventDomainResolved`',
    '`EventContractMessageQueued`',
    '`EventContractMessageProcessed`',
    '`EventReputationUpdated`',
    '`EventFeeDistributed`',
    'canonical attribute',
    'x/actors',
    'each contract behaves as an actor',
    'actor state isolation',
    'actor mailbox',
    'actor message processing',
    'actor lifecycle',
    'state_root',
    'logical_time',
    'mailbox_stats',
    'one actor state transition per delivered message',
    'actor cannot mutate another actor state directly',
    'all cross-actor effects go through messages',
    'exported state includes actor state and mailbox',
    'x/scheduler',
    'parallel execution planning',
    'conflict detection',
    'deterministic batching',
    'safe concurrent state access',
    'sequential deterministic execution',
    'optimistic parallel execution',
    'DAG scheduler',
    'read/write set tracking',
    'fallback to sequential on conflict',
    'x/storage',
    'KV state engine',
    'versioning',
    'snapshots',
    'state sync',
    'contract storage',
    'bounded iteration',
    'contract namespace',
    'storage key format',
    'max state size',
    'storage rent/deposit',
    'export/import exact state',
    'snapshot/state-sync tests'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "execution OS doc missing: $term"
}

foreach ($term in @(
    '`x/execution`',
    '`ExecutionEnvelope`',
    'deterministic execution trace',
    '`x/vm`',
    'AVM and gated CosmWasm runtime facade',
    '`x/messaging`',
    'async contract messaging facade',
    '`x/queue`',
    'deterministic delayed execution',
    '`x/events`',
    'deterministic event schema',
    '`x/actors`',
    'contract actor model',
    '`x/scheduler`',
    'deterministic execution planning',
    'read/write set',
    '`x/storage`',
    'KV state engine',
    'bounded iteration'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing execution OS: $term"
}

Write-Host "execution OS doc test passed"
