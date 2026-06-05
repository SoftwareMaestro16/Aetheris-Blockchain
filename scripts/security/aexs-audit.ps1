param(
  [string]$OutputDir = ".work\aexs",
  [string]$TaskFile = "TO_AUDIT.md",
  [string]$PipelineDoc = "docs\security\aetheris-fuzzing-invariant-pipeline.md",
  [switch]$Json,
  [switch]$EnforceSafe
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

function Get-AexsRepoRoot {
  return [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
}

function Resolve-AexsPath {
  param([string]$Path)
  $repoRoot = Get-AexsRepoRoot
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $repoRoot $Path))
}

function Assert-AexsWorkspacePath {
  param([string]$Path, [string]$Purpose)
  $repoRoot = (Get-AexsRepoRoot).TrimEnd('\', '/')
  $fullPath = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  $prefix = $repoRoot + [System.IO.Path]::DirectorySeparatorChar
  if ($fullPath.Equals($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use repository root as $Purpose`: $fullPath"
  }
  if (-not $fullPath.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use path outside repository as $Purpose`: $fullPath"
  }
}

function Get-AexsRelativePath {
  param([string]$BasePath, [string]$TargetPath)
  $base = [System.IO.Path]::GetFullPath($BasePath).TrimEnd('\', '/')
  $target = [System.IO.Path]::GetFullPath($TargetPath)
  $prefix = $base + [System.IO.Path]::DirectorySeparatorChar
  if ($target.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    return $target.Substring($prefix.Length)
  }
  return $target
}

function Invoke-AexsTextCommand {
  param([string]$FilePath, [string[]]$Arguments)
  try {
    $output = & $FilePath @Arguments 2>$null
    if ($LASTEXITCODE -ne 0) {
      return ""
    }
    return ($output -join "`n").Trim()
  } catch {
    return ""
  }
}

function Get-AexsGoVersion {
  $repoRoot = Get-AexsRepoRoot
  $bundled = Join-Path $repoRoot ".work\tools\go1.25.11\go\bin\go.exe"
  if (Test-Path -LiteralPath $bundled) {
    return Invoke-AexsTextCommand -FilePath $bundled -Arguments @("version")
  }
  $go = Get-Command go -ErrorAction SilentlyContinue
  if ($null -ne $go) {
    return Invoke-AexsTextCommand -FilePath $go.Source -Arguments @("version")
  }
  return "go version unavailable"
}

function Get-AexsSha256Hex {
  param([string]$Text)
  $sha = [System.Security.Cryptography.SHA256]::Create()
  try {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($Text)
    $hash = $sha.ComputeHash($bytes)
    return ([System.BitConverter]::ToString($hash)).Replace("-", "").ToLowerInvariant()
  } finally {
    $sha.Dispose()
  }
}

function Get-AexsTaskCount {
  param([string]$Text, [string]$Prefix)
  return ([regex]::Matches($Text, "(?m)^- \[ \]\s+$([regex]::Escape($Prefix))-\d{2}\b")).Count
}

function Get-AexsMatrixRowText {
  param([string]$Text, [string]$Label)
  if ($Text -match "(?m)^\|\s*$([regex]::Escape($Label))\s*\|.*$") {
    return $Matches[0]
  }
  return ""
}

function Test-AexsMatrixRow {
  param([string]$Text, [string]$Label)
  return (Get-AexsMatrixRowText -Text $Text -Label $Label) -ne ""
}

function Get-AexsMatrixCells {
  param([string]$Row)
  if ([string]::IsNullOrWhiteSpace($Row)) {
    return @()
  }
  return @($Row.Trim().Trim('|').Split('|') | ForEach-Object { $_.Trim() })
}

function Test-AexsTextAny {
  param([string]$Text, [string[]]$Terms)
  foreach ($term in $Terms) {
    if ($Text -match [regex]::Escape($term)) {
      return $true
    }
  }
  return $false
}

function Get-AexsEvidence {
  param([object]$Module)
  $repoRoot = Get-AexsRepoRoot
  $files = @()
  foreach ($root in $Module.EvidenceRoots) {
    $fullRoot = Resolve-AexsPath -Path $root
    if (-not (Test-Path -LiteralPath $fullRoot)) {
      continue
    }
    if (Test-Path -LiteralPath $fullRoot -PathType Leaf) {
      $files += Get-Item -LiteralPath $fullRoot
      continue
    }
    $files += Get-ChildItem -LiteralPath $fullRoot -Recurse -File -Include *.go,*.md,*.ps1
  }

  $matchedFiles = @()
  $invariantFiles = @()
  $fuzzFiles = @()
  $adversarialFiles = @()
  $determinismFiles = @()
  foreach ($file in $files) {
    $text = Get-Content -Raw -LiteralPath $file.FullName
    $matched = $false
    foreach ($term in $Module.EvidenceTerms) {
      if ($text -match [regex]::Escape($term)) {
        $matched = $true
        break
      }
    }
    if (-not $matched) {
      continue
    }
    $relative = Get-AexsRelativePath -BasePath $repoRoot -TargetPath $file.FullName
    $matchedFiles += $relative
    if ($text -match '(?i)invariant|invariants|state integrity|supply consistency|app hash|deterministic replay') {
      $invariantFiles += $relative
    }
    if ($text -match '(?i)\bFuzz[A-Za-z0-9_]*|fuzz') {
      $fuzzFiles += $relative
    }
    if ($text -match '(?i)adversarial|attack|exploit|malformed|unauthorized|replay') {
      $adversarialFiles += $relative
    }
    if ($text -match '(?i)deterministic|determinism|same input|same tx|same genesis') {
      $determinismFiles += $relative
    }
  }

  return [ordered]@{
    files              = @($matchedFiles | Sort-Object -Unique)
    invariant_files    = @($invariantFiles | Sort-Object -Unique)
    fuzz_files         = @($fuzzFiles | Sort-Object -Unique)
    adversarial_files  = @($adversarialFiles | Sort-Object -Unique)
    determinism_files  = @($determinismFiles | Sort-Object -Unique)
  }
}

$repoRoot = Get-AexsRepoRoot
$taskPath = Resolve-AexsPath -Path $TaskFile
$pipelinePath = Resolve-AexsPath -Path $PipelineDoc
$outputRoot = Resolve-AexsPath -Path $OutputDir
Assert-AexsWorkspacePath -Path $outputRoot -Purpose "AEXS output directory"

if (-not (Test-Path -LiteralPath $taskPath)) {
  throw "AEXS task file not found: $taskPath"
}
if (-not (Test-Path -LiteralPath $pipelinePath)) {
  throw "AEXS pipeline source not found: $pipelinePath"
}

$taskText = Get-Content -Raw -LiteralPath $taskPath
$pipelineText = Get-Content -Raw -LiteralPath $pipelinePath
$commit = Invoke-AexsTextCommand -FilePath "git" -Arguments @("rev-parse", "--short=12", "HEAD")
if ([string]::IsNullOrWhiteSpace($commit)) {
  $commit = "no-git-commit"
}
$branch = Invoke-AexsTextCommand -FilePath "git" -Arguments @("branch", "--show-current")
if ([string]::IsNullOrWhiteSpace($branch)) {
  $branch = "detached-or-unknown"
}
$dirtyStatus = Invoke-AexsTextCommand -FilePath "git" -Arguments @("status", "--short")
$sourceHash = (Get-AexsSha256Hex -Text ($taskText + "`n---PIPELINE---`n" + $pipelineText)).Substring(0, 16)
$campaignId = "aexs-$commit-$sourceHash"
$campaignDir = Join-Path $outputRoot $campaignId
New-Item -ItemType Directory -Force -Path $campaignDir | Out-Null

$moduleCatalog = @(
  [ordered]@{ Module = "x/auth"; Label = '`x/auth`'; Prefix = "AUTH"; Value = $true; EvidenceRoots = @("app", "tests\integration", "tests\adversarial", "docs\transaction-lifecycle-matrix.md"); EvidenceTerms = @("auth", "signature", "sequence", "signer") },
  [ordered]@{ Module = "x/bank"; Label = '`x/bank`'; Prefix = "BANK"; Value = $true; EvidenceRoots = @("app", "tests\integration", "tests\adversarial", "docs\transaction-lifecycle-matrix.md"); EvidenceTerms = @("bank", "balance", "send", "supply") },
  [ordered]@{ Module = "x/staking"; Label = '`x/staking`'; Prefix = "STAKE"; Value = $true; EvidenceRoots = @("app", "tests\integration", "tests\e2e", "docs\security\pos-staking-correctness.md"); EvidenceTerms = @("staking", "delegate", "validator", "unbond") },
  [ordered]@{ Module = "x/slashing"; Label = '`x/slashing`'; Prefix = "SLASH"; Value = $true; EvidenceRoots = @("app", "docs\security\slashing-system.md", "docs\security\pos-staking-correctness.md"); EvidenceTerms = @("slashing", "slash", "evidence", "tombstone") },
  [ordered]@{ Module = "x/gov"; Label = '`x/gov`'; Prefix = "GOV"; Value = $true; EvidenceRoots = @("app", "docs", "tests\integration"); EvidenceTerms = @("governance", "proposal", "vote", "authority") },
  [ordered]@{ Module = "x/distribution"; Label = '`x/distribution`'; Prefix = "DIST"; Value = $true; EvidenceRoots = @("app", "docs\security\pos-staking-correctness.md", "tests\integration"); EvidenceTerms = @("distribution", "reward", "commission", "community pool") },
  [ordered]@{ Module = "app"; Label = '`app` / BaseApp'; Prefix = "APP"; Value = $true; EvidenceRoots = @("app", "tests\integration", "docs\genesis-migrations.md", "docs\state-export-import.md"); EvidenceTerms = @("BaseApp", "app hash", "genesis", "export", "determinism") },
  [ordered]@{ Module = "x/fees"; Label = '`x/fees`'; Prefix = "FEES"; Value = $true; EvidenceRoots = @("x\fees", "tests\adversarial", "tests\integration", "docs\fees-ante-policy.md"); EvidenceTerms = @("fees", "fee", "naet", "ante") },
  [ordered]@{ Module = "x/tokenfactory"; Label = '`x/tokenfactory`'; Prefix = "TF"; Value = $true; EvidenceRoots = @("x\tokenfactory", "tests\adversarial", "tests\e2e", "docs\security\module-bank-movement-audit.md"); EvidenceTerms = @("tokenfactory", "mint", "burn", "admin") },
  [ordered]@{ Module = "x/dex"; Label = '`x/dex`'; Prefix = "DEX"; Value = $true; EvidenceRoots = @("x\dex", "tests\adversarial", "tests\e2e", "docs\architecture\dex-direction.md"); EvidenceTerms = @("dex", "pool", "swap", "liquidity", "reserve") },
  [ordered]@{ Module = "x/identity"; Label = '`x/identity`'; Prefix = "ID"; Value = $true; EvidenceRoots = @("x\identity", "tests\adversarial", "docs\architecture\aetheris-modular-execution-os.md"); EvidenceTerms = @("identity", ".aet", "domain", "resolver") },
  [ordered]@{ Module = "x/reputation"; Label = '`x/reputation`'; Prefix = "REP"; Value = $true; EvidenceRoots = @("x\reputation", "docs\module-boundaries.md", "docs\test-production-gates.md"); EvidenceTerms = @("reputation", "score", "rate limit", "priority") },
  [ordered]@{ Module = "x/execution"; Label = '`x/execution`'; Prefix = "EXEC"; Value = $true; EvidenceRoots = @("x\execution", "docs\architecture\execution-os.md", "docs\module-boundaries.md"); EvidenceTerms = @("execution", "dispatch", "route", "receipt") },
  [ordered]@{ Module = "x/vm"; Label = '`x/vm` / AVM'; Prefix = "VM"; Value = $true; EvidenceRoots = @("x\vm", "x\aetherisvm", "docs\architecture\avm.md", "docs\architecture\vm-direction.md"); EvidenceTerms = @("vm", "AVM", "bytecode", "gas") },
  [ordered]@{ Module = "x/aetherisvm"; Label = '`x/vm` / AVM'; Prefix = "VM"; Value = $true; EvidenceRoots = @("x\aetherisvm", "docs\architecture\avm.md", "docs\architecture\vm-direction.md"); EvidenceTerms = @("AVM", "async", "contract", "gas") },
  [ordered]@{ Module = "x/messaging"; Label = '`x/messaging`'; Prefix = "MSG"; Value = $true; EvidenceRoots = @("x\messaging", "x\mesh", "tests\adversarial", "docs\architecture\execution-os.md"); EvidenceTerms = @("messaging", "message", "receipt", "proof") },
  [ordered]@{ Module = "x/queue"; Label = '`x/queue`'; Prefix = "QUEUE"; Value = $true; EvidenceRoots = @("x\queue", "x\aetherisvm\async", "docs\architecture\async-smart-contract-execution.md"); EvidenceTerms = @("queue", "bounce", "refund", "delayed") },
  [ordered]@{ Module = "x/events"; Label = '`x/events`'; Prefix = "EVENTS"; Value = $false; EvidenceRoots = @("x\events", "docs\event-contract.md", "tests\scripts\event_contract_doc_test.ps1"); EvidenceTerms = @("events", "event", "receipt", "attributes") },
  [ordered]@{ Module = "x/actors"; Label = '`x/actors`'; Prefix = "ACTOR"; Value = $true; EvidenceRoots = @("x\actors", "docs\module-boundaries.md"); EvidenceTerms = @("actor", "mailbox", "logical time") },
  [ordered]@{ Module = "x/scheduler"; Label = '`x/scheduler`'; Prefix = "SCHED"; Value = $true; EvidenceRoots = @("x\scheduler", "x\schedulerv2", "docs\module-boundaries.md"); EvidenceTerms = @("scheduler", "schedule", "task", "priority") },
  [ordered]@{ Module = "x/storage"; Label = '`x/storage`'; Prefix = "STORE"; Value = $true; EvidenceRoots = @("x\storage", "docs\module-boundaries.md", "docs\architecture\avm.md"); EvidenceTerms = @("storage", "snapshot", "export", "state root") },
  [ordered]@{ Module = "x/memo"; Label = '`x/memo`'; Prefix = "MEMO"; Value = $true; EvidenceRoots = @("x\memo", "docs\mempool-checktx-negative-flow.md", "docs\transaction-lifecycle-matrix.md"); EvidenceTerms = @("memo", "UTF-8", "metadata") },
  [ordered]@{ Module = "x/indexer"; Label = '`x/indexer`'; Prefix = "INDEX"; Value = $false; EvidenceRoots = @("x\indexer", "app\indexer", "docs\event-contract.md", "docs\query-surface.md"); EvidenceTerms = @("index", "indexer", "query", "event") },
  [ordered]@{ Module = "x/sharding/sim"; Label = '`x/sharding/sim` and load/routing'; Prefix = "SHARD"; Value = $true; EvidenceRoots = @("x\sharding", "x\load", "x\routing", "tests\adversarial", "docs\architecture\sharding-rd.md"); EvidenceTerms = @("sharding", "LOAD_SCORE", "route", "shard") }
)

$requiredSourceTerms = @(
  "docs/security/aetheris-fuzzing-invariant-pipeline.md",
  ".work/aexs/",
  "100%",
  "95%",
  "defensive analysis",
  "adversarial simulation",
  "mandatory invariants",
  "Coverage Matrix"
)

$sourceFailures = @()
foreach ($term in $requiredSourceTerms) {
  if (-not (Test-AexsTextAny -Text $taskText -Terms @($term)) -and -not (Test-AexsTextAny -Text $pipelineText -Terms @($term))) {
    $sourceFailures += "missing source term: $term"
  }
}

$moduleRows = @()
foreach ($module in $moduleCatalog) {
  $taskCount = Get-AexsTaskCount -Text $taskText -Prefix $module.Prefix
  $matrixRow = Get-AexsMatrixRowText -Text $taskText -Label $module.Label
  $matrixCells = Get-AexsMatrixCells -Row $matrixRow
  $hasMatrixRow = $matrixRow -ne ""
  $hasAttackSurface = $hasMatrixRow -and $matrixCells.Count -ge 5 -and -not [string]::IsNullOrWhiteSpace($matrixCells[3])
  $hasInvariantPlan = $hasMatrixRow -and $matrixCells.Count -ge 5 -and -not [string]::IsNullOrWhiteSpace($matrixCells[4])
  $hasValueTask = (-not $module.Value) -or ($taskText -match "$([regex]::Escape($module.Prefix))-05")
  $plannedChecks = @(($taskCount -ge 5), $hasMatrixRow, $hasAttackSurface, $hasInvariantPlan, $hasValueTask)
  $passedPlanned = @($plannedChecks | Where-Object { $_ -eq $true }).Count
  $plannedCoverage = [math]::Round(($passedPlanned / $plannedChecks.Count) * 100, 2)
  $evidence = Get-AexsEvidence -Module $module
  $hasInvariantEvidence = @($evidence.invariant_files).Count -gt 0
  $hasAdversarialEvidence = @($evidence.adversarial_files).Count -gt 0
  $hasFuzzEvidence = @($evidence.fuzz_files).Count -gt 0
  $safe = $false
  $reasons = @()
  if ($taskCount -lt 5) { $reasons += "fewer than five atomic audit tasks" }
  if (-not $hasMatrixRow) { $reasons += "missing mandatory coverage matrix row" }
  if (-not $hasInvariantPlan) { $reasons += "missing planned invariant mapping" }
  if (-not $hasInvariantEvidence) { $reasons += "no invariant evidence found" }
  if (-not $hasAdversarialEvidence) { $reasons += "no adversarial evidence found" }
  if (-not $hasFuzzEvidence) { $reasons += "no fuzz evidence found" }
  if ($plannedCoverage -lt 95) { $reasons += "planned coverage below 95 percent" }

  $moduleRows += [ordered]@{
    module                    = $module.Module
    task_prefix               = $module.Prefix
    task_count                = $taskCount
    planned_coverage_percent  = $plannedCoverage
    has_matrix_row            = $hasMatrixRow
    has_attack_surface        = $hasAttackSurface
    has_invariant_plan        = $hasInvariantPlan
    has_value_task            = $hasValueTask
    has_invariant_evidence    = $hasInvariantEvidence
    has_adversarial_evidence  = $hasAdversarialEvidence
    has_fuzz_evidence         = $hasFuzzEvidence
    evidence_files            = $evidence.files
    invariant_files           = $evidence.invariant_files
    fuzz_files                = $evidence.fuzz_files
    adversarial_files         = $evidence.adversarial_files
    determinism_files         = $evidence.determinism_files
    safe                      = $safe
    safe_blockers             = $reasons
  }
}

$plannedCoverageTotal = 0.0
foreach ($row in $moduleRows) {
  $plannedCoverageTotal += [double]$row["planned_coverage_percent"]
}
$plannedCoverageAverage = [math]::Round(($plannedCoverageTotal / [double]$moduleRows.Count), 2)
$modulesBelowPlan = @($moduleRows | Where-Object { $_["planned_coverage_percent"] -lt 95 -or $_["task_count"] -lt 5 -or -not $_["has_matrix_row"] -or -not $_["has_invariant_plan"] })
$modulesWithoutInvariantEvidence = @($moduleRows | Where-Object { -not $_["has_invariant_evidence"] })
$modulesWithoutFuzzEvidence = @($moduleRows | Where-Object { -not $_["has_fuzz_evidence"] })
$modulesWithoutAdversarialEvidence = @($moduleRows | Where-Object { -not $_["has_adversarial_evidence"] })
$mandatoryInvariantPassRate = 0
$auditPassed = $false
$productionSafe = $false

$runtimeModes = @(
  "stateless fuzzing",
  "stateful multi-block fuzzing",
  "adversarial red-team fuzzing",
  "deterministic replay",
  "stress mode",
  "chaos mode"
)
$simulatorModes = @(
  "in-memory app runner",
  "single-validator localnet",
  "multi-validator localnet",
  "sharding simulator"
)
$fuzzSeeds = @(
  "aexs-auth-replay-0001",
  "aexs-fee-denom-spoof-0002",
  "aexs-tokenfactory-admin-0003",
  "aexs-dex-reserve-desync-0004",
  "aexs-identity-resolver-hijack-0005",
  "aexs-avm-malformed-bytecode-0006",
  "aexs-routing-load-poison-0007",
  "aexs-mesh-replay-0008"
)
$testCommands = @(
  "go test ./...",
  "go vet ./...",
  "buf lint",
  "powershell -NoProfile -ExecutionPolicy Bypass -File tests\scripts\determinism_gate_test.ps1",
  "go test -run '^$' -fuzz <target> -fuzztime <duration>"
)

$summary = [ordered]@{
  campaign_id                         = $campaignId
  output_dir                          = $campaignDir
  source_task_file                    = Get-AexsRelativePath -BasePath $repoRoot -TargetPath $taskPath
  source_pipeline_doc                 = Get-AexsRelativePath -BasePath $repoRoot -TargetPath $pipelinePath
  source_hash                         = $sourceHash
  git_commit                          = $commit
  git_branch                          = $branch
  git_dirty_status                    = $dirtyStatus
  go_version                          = Get-AexsGoVersion
  os                                  = [System.Runtime.InteropServices.RuntimeInformation]::OSDescription
  test_commands                       = $testCommands
  fuzz_seeds                          = $fuzzSeeds
  runtime_modes                       = $runtimeModes
  simulator_modes                     = $simulatorModes
  target_modules                      = @($moduleRows | ForEach-Object { $_["module"] })
  module_count                        = $moduleRows.Count
  planned_coverage_percent            = $plannedCoverageAverage
  modules_below_planned_threshold     = @($modulesBelowPlan | ForEach-Object { $_["module"] })
  modules_without_invariant_evidence  = @($modulesWithoutInvariantEvidence | ForEach-Object { $_["module"] })
  modules_without_fuzz_evidence       = @($modulesWithoutFuzzEvidence | ForEach-Object { $_["module"] })
  modules_without_adversarial_evidence = @($modulesWithoutAdversarialEvidence | ForEach-Object { $_["module"] })
  mandatory_invariant_pass_rate       = $mandatoryInvariantPassRate
  audit_passed                        = $auditPassed
  production_safe                     = $productionSafe
  decision                            = "NOT_SAFE_PRE_CAMPAIGN"
  decision_reason                     = "AEXS structural plan can be audited, but full fuzz campaign execution and 100 percent invariant pass evidence are not recorded yet."
  source_failures                     = $sourceFailures
}

$coveragePath = Join-Path $campaignDir "coverage-matrix.json"
$summaryPath = Join-Path $campaignDir "summary.json"
$resultPath = Join-Path $campaignDir "AUDIT_RESULT.md"
$taskCopyPath = Join-Path $campaignDir "TO_AUDIT.md"
$moduleRows | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $coveragePath
$summary | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $summaryPath
Copy-Item -LiteralPath $taskPath -Destination $taskCopyPath -Force

$report = @()
$report += "# AEXS Audit Result"
$report += ""
$report += "- campaign id: $campaignId"
$report += "- git commit: $commit"
$report += "- branch: $branch"
$report += "- output dir: $campaignDir"
$report += "- decision: NOT_SAFE_PRE_CAMPAIGN"
$report += "- planned coverage: $plannedCoverageAverage%"
$report += "- mandatory invariant pass rate: $mandatoryInvariantPassRate%"
$report += ""
$report += "## Gate Decision"
$report += ""
$report += "The audit is not passed yet. This preflight validates that the campaign plan and matrix are machine-checkable, but production-safe status requires an executed fuzz/invariant campaign with 100% mandatory invariant pass rate and no untriaged Critical or High exploit."
$report += ""
$report += "## Coverage Gaps"
$report += ""
$report += "- modules below 95% planned coverage: $(@($modulesBelowPlan | ForEach-Object { $_["module"] }) -join ', ')"
$report += "- modules without invariant evidence: $(@($modulesWithoutInvariantEvidence | ForEach-Object { $_["module"] }) -join ', ')"
$report += "- modules without fuzz evidence: $(@($modulesWithoutFuzzEvidence | ForEach-Object { $_["module"] }) -join ', ')"
$report += "- modules without adversarial evidence: $(@($modulesWithoutAdversarialEvidence | ForEach-Object { $_["module"] }) -join ', ')"
$report += ""
$report += "## Module Matrix"
$report += ""
$report += "| Module | Tasks | Planned coverage | Invariant evidence | Fuzz evidence | Adversarial evidence | Safe |"
$report += "| --- | ---: | ---: | --- | --- | --- | --- |"
foreach ($row in $moduleRows) {
  $report += "| $($row["module"]) | $($row["task_count"]) | $($row["planned_coverage_percent"])% | $($row["has_invariant_evidence"]) | $($row["has_fuzz_evidence"]) | $($row["has_adversarial_evidence"]) | $($row["safe"]) |"
}
$report += ""
$report += "## Required Next Step"
$report += ""
$report += "Run the AEXS fuzzing and invariant campaign, write generated scenarios, minimized exploits, state diffs, and final results under `.work/aexs/`, then update this result with executed invariant pass rates and triage status."
$report | Set-Content -LiteralPath $resultPath

if ($sourceFailures.Count -gt 0) {
  throw "AEXS audit source validation failed: $($sourceFailures -join '; ')"
}
if ($modulesBelowPlan.Count -gt 0) {
  throw "AEXS planned coverage gate failed for module(s): $(@($modulesBelowPlan | ForEach-Object { $_["module"] }) -join ', ')"
}
if ($EnforceSafe -and -not $auditPassed) {
  throw "AEXS audit is not production safe; see $resultPath"
}

if ($Json) {
  $summary | ConvertTo-Json -Depth 8
} else {
  Write-Host "AEXS audit preflight complete"
  Write-Host "Campaign: $campaignId"
  Write-Host "Planned coverage: $plannedCoverageAverage%"
  Write-Host "Decision: NOT_SAFE_PRE_CAMPAIGN"
  Write-Host "Report: $resultPath"
}
