param(
  [string]$Doc = "docs\security\slashing-system.md",
  [string]$AuditPack = "docs\security\security-audit-pack.md",
  [string]$PosDoc = "docs\security\pos-staking-correctness.md",
  [string]$Policy = "app\params\slashing_policy.go",
  [string]$PolicyTests = "app\params\slashing_policy_test.go"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$AuditPath = if ([System.IO.Path]::IsPathRooted($AuditPack)) { $AuditPack } else { Join-Path $RepoRoot $AuditPack }
$PosPath = if ([System.IO.Path]::IsPathRooted($PosDoc)) { $PosDoc } else { Join-Path $RepoRoot $PosDoc }
$PolicyPath = if ([System.IO.Path]::IsPathRooted($Policy)) { $Policy } else { Join-Path $RepoRoot $Policy }
$PolicyTestsPath = if ([System.IO.Path]::IsPathRooted($PolicyTests)) { $PolicyTests } else { Join-Path $RepoRoot $PolicyTests }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$auditText = Get-Content -Raw -LiteralPath $AuditPath
$posText = Get-Content -Raw -LiteralPath $PosPath
$policyText = Get-Content -Raw -LiteralPath $PolicyPath
$policyTestsText = Get-Content -Raw -LiteralPath $PolicyTestsPath

foreach ($term in @(
    'Aetra Slashing System',
    'core consensus security primitive',
    'validator honesty',
    'chain finality safety',
    'prevention of forks',
    'prevention of double-signing attacks',
    'Without slashing',
    'Staking And Validator Model',
    'validator set structure',
    'staking collateral is AET in base denom `naet`',
    'minimum validator self-delegation',
    'delegators share slashing exposure pro rata',
    'unbonding stake remains slashable',
    'Slashable Conditions',
    'Double signing',
    'Equivocation in consensus votes',
    'Conflicting block proposals',
    'Prolonged downtime',
    'Invalid votes beyond tolerance threshold',
    'Censorship proof',
    'Protocol signing rule violation',
    'Evidence And Proof System',
    'cryptographically verifiable',
    'no off-chain trust assumptions',
    'SlashingEvidence',
    'nodes independently verify',
    'Slashing Execution Flow',
    'evidence id',
    'processed evidence ids',
    'duplicate evidence',
    'Conflicting evidence resolution',
    'Economic Security Model',
    'expected_attack_cost',
    'expected_attack_reward',
    'Penalty distribution',
    'critical_slash',
    'medium_slash',
    'Validator Lifecycle Impact',
    'candidate',
    'active validator set',
    'tombstoned',
    'unbonding stake locked',
    'redelegating stake remains slashable',
    'Governance Constraints',
    'Governance can adjust only bounded parameters',
    'Governance cannot',
    'reverse valid slashing events',
    'override cryptographic proofs',
    'selectively punish validators',
    'Security Model And Attack Resistance',
    'Validator collusion',
    'Long-range attacks',
    'Bribery attacks',
    'Fake slashing evidence',
    'Evidence spam and griefing',
    'Cryptographic Assumptions Summary',
    'Slashable Event Table',
    '25. Slashing Implementation Details',
    '25.1 Standard Slashing Integration',
    'Cosmos SDK `x/slashing` and CometBFT evidence',
    'double-sign',
    'liveness/downtime',
    'jail/unjail',
    '`x/slashing` remains the source of truth',
    'custom logic may wrap or extend standard behavior only where necessary',
    'custom logic must not fork core slashing behavior',
    'SlashingAccountabilityPolicy',
    'UsesCosmosSlashingAndEvidence',
    'BaseFaultsUseCometBFTEvidence',
    'StandardDoubleSignIntegrated',
    'StandardLivenessDowntimeIntegrated',
    'StandardTombstoneIntegrated',
    'StandardJailUnjailIntegrated',
    'CustomLogicWrapsStandardOnly',
    'CoreSlashingForkForbidden',
    '25.2 Progressive Downtime Design',
    'DowntimeOffense:',
    'ValidatorConsAddr',
    'OffenseCount',
    'FirstOffenseTime',
    'LastOffenseTime',
    'LastSlashFraction',
    'CurrentJailDuration',
    'offense count decays after long clean period',
    'repeated downtime increases penalty',
    'maximum penalty is capped',
    'delegators inherit validator downtime risk',
    'validator can query own downtime status',
    'unjail does not erase slash history immediately',
    'downtime_offense_clean_decay_period = 30 days',
    'downtime_offense_status_query       = QueryDowntimeOffenseStatus',
    '25.3 Evidence And Unbonding',
    'evidence submitted while validator bonded',
    'evidence submitted while validator unbonding',
    'evidence submitted after unbonding but before evidence expiration',
    'evidence submitted after expiration',
    'slashing delegators who were bonded at infraction height',
    'tombstone cap behavior',
    'bonded validators are slashable immediately',
    'unbonding validators remain slashable',
    'expired evidence must be rejected before stake mutation',
    'delegator slash accounting must use stake/shares at infraction height',
    'duplicate evidence must not multiply critical penalties',
    '25.4 Invalid Proposal Policy',
    'reject objectively invalid proposals',
    'do not slash unless there is signed, reproducible evidence',
    'do not depend on local wall clock',
    'do not depend on external APIs',
    'do not make `ProcessProposal` fragile',
    'invalid tx proposal rejected',
    'oversized proposal rejected',
    'malformed special tx rejected',
    'valid proposal accepted',
    'same proposal accepted/rejected identically by all validators in test harness',
    'ProcessProposal` must use only deterministic proposal bytes',
    'local system time, remote RPC, indexers, price feeds, HTTP APIs',
    'oversized proposals must be rejected before expensive decoding or VM work',
    'Exact Penalty Structure',
    'double_sign_slash_fraction        = 0.05',
    'downtime_slash_fraction           = 0.0001',
    'Rounding dust goes to burn',
    'Slashing in Aetra is a deterministic, protocol-level enforcement mechanism that guarantees economic security of consensus through enforceable stake penalties.'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "slashing system doc missing: $term"
}

foreach ($term in @(
    'Aetra Slashing System',
    'slashing-system.md'
  )) {
  Assert-Contains -Text $auditText -Pattern ([regex]::Escape($term)) -Message "security audit pack missing slashing doc link: $term"
  Assert-Contains -Text $posText -Pattern ([regex]::Escape($term)) -Message "PoS doc missing slashing doc link: $term"
}

foreach ($term in @(
    'EvidenceWhileValidatorBondedTest',
    'EvidenceWhileValidatorUnbondingTest',
    'EvidenceAfterUnbondingBeforeExpiryTest',
    'EvidenceAfterExpirationRejectedTest',
    'DelegatorInfractionHeightSlashTest',
    'TombstoneCapBehaviorTest',
    'InvalidTxProposalRejectedTest',
    'OversizedProposalRejectedTest',
    'MalformedSpecialTxRejectedTest',
    'ValidProposalAcceptedTest',
    'AllValidatorsProposalDeterminismTest',
    'InvalidProposalWallClockForbidden',
    'InvalidProposalExternalAPIsForbidden',
    'ProcessProposalFragilityForbidden',
    'InvalidProposalMaxBytesDefault'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "slashing policy missing: $term"
}

foreach ($term in @(
    'TestSlashingAccountabilityPolicyRequiresEvidenceAndUnbondingCoverage',
    'TestSlashingAccountabilityPolicyRequiresInvalidProposalCoverage',
    'validator bonded',
    'validator unbonding',
    'before evidence expiration',
    'after expiration',
    'delegators bonded at infraction height',
    'tombstone cap',
    'invalid tx proposal',
    'oversized proposal',
    'malformed special tx',
    'valid proposal',
    'across validators',
    'local wall clock',
    'external APIs',
    'not be fragile'
  )) {
  Assert-Contains -Text $policyTestsText -Pattern ([regex]::Escape($term)) -Message "slashing policy tests missing: $term"
}

Write-Host "slashing system doc test passed"
