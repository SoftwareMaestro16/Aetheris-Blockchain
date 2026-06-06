param(
  [string]$AuditPack = "docs\security\security-audit-pack.md",
  [string]$Workflow = ".github\workflows\security.yml",
  [string]$ManualChecklist = "docs\security\manual-audit-checklist.md",
  [string]$CosmosChecklist = "docs\security\cosmos-security-checklist.md",
  [string]$PrototypeGate = "docs\security\prototype-audit-gate.md",
  [string]$TriagePolicy = "docs\security\security-triage-policy.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Path))
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$auditText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $AuditPack)
$workflowText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Workflow)
$manualText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $ManualChecklist)
$cosmosText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $CosmosChecklist)
$prototypeText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $PrototypeGate)
$triageText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $TriagePolicy)

foreach ($term in @(
    "Security Audit Pack",
    "Public testnet cannot proceed with untriaged high/critical",
    "fund-safety",
    "consensus-safety",
    "secret-leak",
    "Security Risks And Controls",
    "Aetra Slashing System",
    "govulncheck",
    "gosec high severity",
    "gitleaks secrets",
    "dependency review",
    "CodeQL"
  )) {
  Assert-Contains -Text $auditText -Pattern ([regex]::Escape($term)) -Message "audit pack missing required term: $term"
}

foreach ($check in @(
    "non-determinism",
    "incorrect signers",
    "ABCI panics",
    "unsafe rounding",
    "unbounded iteration",
    "malformed genesis",
    "replay paths",
    "invalid authority paths",
    "wallet replay",
    "wrong wallet_id",
    "extension takeover",
    "token supply divergence",
    "NFT unauthorized transfer",
    "SBT transfer bypass",
    "async queue DoS",
    "bounce/refund double-spend",
    "metadata spoofing",
    "admin takeover"
  )) {
  Assert-Contains -Text $auditText -Pattern ([regex]::Escape($check)) -Message "audit pack missing check: $check"
}

foreach ($workflowTerm in @(
    "name: govulncheck",
    "name: gosec high severity",
    "name: gitleaks secrets",
    "name: dependency review",
    "name: CodeQL",
    "fail-on-severity: high",
    ".github/security/govulncheck-triage.txt"
  )) {
  Assert-Contains -Text $workflowText -Pattern ([regex]::Escape($workflowTerm)) -Message "security workflow missing gate: $workflowTerm"
}

foreach ($term in @(
    "Contract Standards Review",
    'wrong `wallet_id`',
    "extension",
    "token supply divergence",
    "ANFT-66",
    "ASBT-67",
    "queue DoS",
    "bounce/refund",
    "metadata spoofing",
    "admin takeover",
    "naet"
  )) {
  Assert-Contains -Text $manualText -Pattern ([regex]::Escape($term)) -Message "manual checklist missing contract/security term: $term"
}

foreach ($term in @(
    "Contract Standards And Async",
    "wallet replay",
    'wrong `wallet_id`',
    "extension takeover",
    "token supply divergence",
    "NFT/SBT transfer bypass",
    "async queue DoS",
    "bounce/refund double-spend",
    "metadata spoofing",
    "admin takeover"
  )) {
  Assert-Contains -Text $cosmosText -Pattern ([regex]::Escape($term)) -Message "Cosmos checklist missing Phase 14 coverage: $term"
}

foreach ($term in @(
    "Public testnet cannot proceed",
    "untriaged High/Critical",
    "fund-safety",
    "consensus-safety",
    "secret-leak",
    "Contract Manual Checklist",
    "Wallet replay",
    "Extension takeover",
    "Token accounting",
    "NFT/SBT transfer",
    "Async queue",
    "Bounce/refund",
    "Metadata spoofing"
  )) {
  Assert-Contains -Text $prototypeText -Pattern ([regex]::Escape($term)) -Message "prototype audit gate missing Phase 14 coverage: $term"
}

foreach ($term in @(
    "Aetra",
    "Public Testnet Gate",
    'untriaged `Critical` or `High`',
    "fund-safety",
    "consensus-safety",
    "secret-leak",
    "govulncheck",
    "gosec",
    "CodeQL",
    "gitleaks",
    "Dependency Review"
  )) {
  Assert-Contains -Text $triageText -Pattern ([regex]::Escape($term)) -Message "triage policy missing gate term: $term"
}

Write-Host "security audit pack doc test passed"
