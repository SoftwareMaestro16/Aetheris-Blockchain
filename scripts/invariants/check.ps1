param(
  [string]$Package = "./app",
  [string]$Run = "Invariant",
  [switch]$IncludeCLI
)

$ErrorActionPreference = "Stop"
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
Push-Location $RepoRoot
try {
  & go test $Package -run $Run -count=1
  if ($IncludeCLI) {
    & go test ./cmd/l1d/cmd -run "TestInvariant(List|Check|Command)" -count=1
  }
} finally {
  Pop-Location
}
