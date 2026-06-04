param(
  [string]$OutputDir = ""
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"

& (Join-Path $PSScriptRoot "stop.ps1") -OutputDir $OutputDir
if (Test-Path $OutputDir) {
  Remove-LocalnetDirectory -OutputDir $OutputDir
  Write-Host "Removed $OutputDir"
}
