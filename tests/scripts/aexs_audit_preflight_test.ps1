param(
  [string]$OutputDir = ".work\aexs-test"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Path))
}

function Assert-True {
  param([bool]$Condition, [string]$Message)
  if (-not $Condition) {
    throw $Message
  }
}

function Assert-Contains {
  param([object[]]$Values, [string]$Expected, [string]$Message)
  if ($Expected -notin $Values) {
    throw $Message
  }
}

$resolvedOutput = Resolve-RepoPath $OutputDir
$repoPrefix = $RepoRoot.TrimEnd('\', '/') + [System.IO.Path]::DirectorySeparatorChar
Assert-True ($resolvedOutput.StartsWith($repoPrefix, [System.StringComparison]::OrdinalIgnoreCase)) "AEXS test output must stay under repository"

if (Test-Path -LiteralPath $resolvedOutput) {
  Remove-Item -LiteralPath $resolvedOutput -Recurse -Force
}

Push-Location $RepoRoot
try {
  $jsonText = & .\scripts\security\aexs-audit.ps1 -OutputDir $OutputDir -Json
  $result = $jsonText | ConvertFrom-Json

  Assert-True ($result.campaign_id -match '^aexs-[0-9a-f]{12}-[0-9a-f]{16}$') "campaign id must be deterministic and commit-based"
  Assert-True ($result.output_dir.StartsWith($resolvedOutput, [System.StringComparison]::OrdinalIgnoreCase)) "runtime report must be under requested .work output"
  Assert-True ($result.source_task_file -eq "TO_AUDIT.md") "TO_AUDIT must be the task source"
  Assert-True ($result.source_pipeline_doc -eq "docs\security\aetheris-fuzzing-invariant-pipeline.md") "pipeline doc must be the primary source"
  Assert-True ($result.planned_coverage_percent -ge 95) "planned coverage must meet 95 percent threshold"
  Assert-True ($result.audit_passed -eq $false) "pre-campaign audit must not be marked passed"
  Assert-True ($result.production_safe -eq $false) "pre-campaign audit must not be production safe"
  Assert-True ($result.mandatory_invariant_pass_rate -eq 0) "pre-campaign invariant pass rate must be zero until execution evidence exists"
  Assert-True (@($result.modules_below_planned_threshold).Count -eq 0) "no module can be below planned coverage threshold"

  foreach ($module in @(
      "app",
      "x/fees",
      "x/tokenfactory",
      "x/dex",
      "x/aetherisvm",
      "x/execution",
      "x/vm",
      "x/messaging",
      "x/queue",
      "x/events",
      "x/actors",
      "x/scheduler",
      "x/storage",
      "x/identity",
      "x/reputation",
      "x/sharding/sim"
    )) {
    Assert-Contains -Values $result.target_modules -Expected $module -Message "AEXS target module missing: $module"
  }

  foreach ($name in @("summary.json", "coverage-matrix.json", "AUDIT_RESULT.md", "TO_AUDIT.md")) {
    Assert-True (Test-Path -LiteralPath (Join-Path $result.output_dir $name)) "AEXS output missing $name"
  }

  $coverage = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "coverage-matrix.json") | ConvertFrom-Json
  Assert-True (@($coverage).Count -ge 24) "coverage matrix must include all required module surfaces"
  Assert-True (@($coverage | Where-Object { $_.task_count -lt 5 }).Count -eq 0) "every module must have at least five tasks"
  Assert-True (@($coverage | Where-Object { $_.planned_coverage_percent -lt 95 }).Count -eq 0) "every module must meet planned coverage threshold"
  Assert-True (@($coverage | Where-Object { $_.safe -eq $true }).Count -eq 0) "no module may be marked safe by preflight alone"

  $enforceFailed = $false
  try {
    & .\scripts\security\aexs-audit.ps1 -OutputDir $OutputDir -EnforceSafe | Out-Null
  } catch {
    $enforceFailed = $true
  }
  Assert-True $enforceFailed "EnforceSafe must fail until executed fuzz/invariant evidence passes"
} finally {
  Pop-Location
}

Write-Host "AEXS audit preflight test passed"
