param(
  [string]$Script = "scripts\invariants\check.ps1"
)

$ErrorActionPreference = "Stop"
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$ScriptPath = if ([System.IO.Path]::IsPathRooted($Script)) { $Script } else { Join-Path $RepoRoot $Script }
$text = Get-Content -Raw -LiteralPath $ScriptPath

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

Assert-Contains -Text $text -Pattern 'go test \$Package -run \$Run -count=1' -Message "invariant script must run targeted app invariant tests"
Assert-Contains -Text $text -Pattern '\./app' -Message "invariant script must default to app package"
Assert-Contains -Text $text -Pattern 'Invariant' -Message "invariant script must default to invariant tests"
Assert-Contains -Text $text -Pattern 'IncludeCLI' -Message "invariant script must optionally include CLI invariant tests"

Write-Host "invariant runner script test passed"
