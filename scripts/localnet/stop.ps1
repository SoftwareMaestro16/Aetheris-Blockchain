param(
  [string]$OutputDir = "",
  [int]$TimeoutSeconds = 10
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"
$pidDir = Join-Path $OutputDir "pids"

if (!(Test-Path -LiteralPath $pidDir)) {
  Write-Host "No pid directory found at $pidDir"
}

Stop-LocalnetProcesses -OutputDir $OutputDir -PidDir $pidDir
