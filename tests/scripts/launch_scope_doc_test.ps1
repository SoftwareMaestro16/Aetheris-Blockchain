param(
  [string]$TestnetDoc = "docs\TESTNET.md",
  [string]$Readme = "README.md",
  [string]$OfficialStaking = "docs\official-liquid-staking.md",
  [string]$ValidatorOnboarding = "docs\validator-onboarding.md",
  [string]$ReleaseDir = "docs\release",
  [string]$PublicTestnetGlob = "docs\public-testnet*.md"
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

function Assert-NotContains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -match $Pattern) { throw $Message }
}

function Get-LaunchDocFiles {
  $files = @(
    (Resolve-RepoPath $TestnetDoc),
    (Resolve-RepoPath $Readme),
    (Resolve-RepoPath $OfficialStaking),
    (Resolve-RepoPath $ValidatorOnboarding)
  )

  $releasePath = Resolve-RepoPath $ReleaseDir
  if (Test-Path -LiteralPath $releasePath) {
    $files += Get-ChildItem -LiteralPath $releasePath -Filter "*.md" -File | Select-Object -ExpandProperty FullName
  }

  $publicPattern = Resolve-RepoPath $PublicTestnetGlob
  $files += Get-ChildItem -Path $publicPattern -File -ErrorAction SilentlyContinue | Select-Object -ExpandProperty FullName

  return $files | Sort-Object -Unique
}

function Get-RepoRelativePath {
  param([string]$Path)
  $full = [System.IO.Path]::GetFullPath($Path)
  if ($full.StartsWith($RepoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    return $full.Substring($RepoRoot.Length).TrimStart('\', '/')
  }
  return $full
}

$testnetText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $TestnetDoc)

foreach ($term in @(
    'Aetra Testnet Kernel',
    'Cosmos SDK application node running on CometBFT consensus',
    'AWCE-1 wallet compatibility layer',
    '`AE...` user-facing addresses',
    '`4:...` raw/internal/proof addresses',
    'native bank balance layer',
    'native account, auth, freeze, and storage-rent behavior',
    'already wired into the app',
    'pool-based staking through the official pool/index flow',
    'AVM contracts and AVM contract standards',
    'pool deposit',
    'Normal users do not choose validators directly'
  )) {
  Assert-Contains -Text $testnetText -Pattern ([regex]::Escape($term)) -Message "TESTNET kernel doc missing: $term"
}

$launchDocs = Get-LaunchDocFiles
if ($launchDocs.Count -eq 0) { throw "no launch docs found" }

foreach ($doc in $launchDocs) {
  $text = Get-Content -Raw -LiteralPath $doc
  $relative = Get-RepoRelativePath $doc

  Assert-NotContains -Text $text -Pattern '(?im)^\s*(?:build\\aetrad\.exe|aetrad)\s+tx\s+staking\s+delegate\b' -Message "$relative teaches direct validator delegation"

  Assert-NotContains -Text $text -Pattern '(?im)^\s*(?:build\\aetrad\.exe|aetrad)\s+query\s+staking\s+delegation\b' -Message "$relative teaches normal users validator delegation queries"
  Assert-NotContains -Text $text -Pattern '(?im)^\s*(?:build\\aetrad\.exe|aetrad)\s+query\s+staking\s+delegations\b' -Message "$relative teaches normal users validator delegation queries"

}

Assert-Contains -Text $testnetText -Pattern '(?i)official pool deposit' -Message "TESTNET user staking example must mention official pool deposit"

Write-Host "launch scope doc test passed"
