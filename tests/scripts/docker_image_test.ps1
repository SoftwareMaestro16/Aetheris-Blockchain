param(
  [string]$Dockerfile = "Dockerfile",
  [string]$Compose = "docker-compose.localnet.yml",
  [string]$Workflow = ".github\workflows\testnet-readiness.yml"
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

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$dockerfileText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Dockerfile)
$composeText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Compose)
$workflowText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Workflow)

foreach ($term in @(
    'FROM golang:1.25.11-bookworm AS builder',
    'ARG VERSION=dev',
    'ARG COMMIT=unknown',
    'ARG BUILD_DATE=1970-01-01T00:00:00Z',
    'ARG DIRTY=false',
    'go build -trimpath -buildvcs=false',
    '-X github.com/sovereign-l1/l1/cmd/l1d/cmd.appVersion=${VERSION}',
    '-X github.com/sovereign-l1/l1/cmd/l1d/cmd.gitCommit=${COMMIT}',
    'COPY assets/aetra.png',
    'useradd --create-home',
    'USER aetra',
    'HEALTHCHECK',
    'status --home "$DAEMON_HOME/localnet/node0/aetrad" --node tcp://127.0.0.1:26657 --output json',
    'CMD ["start", "--home", "/home/aetra/localnet/node0/aetrad", "--log_level", "info"]'
  )) {
  Assert-Contains -Text $dockerfileText -Pattern ([regex]::Escape($term)) -Message "Dockerfile missing term: $term"
}

foreach ($term in @(
    'services:',
    'init:',
    'node0:',
    'init-localnet',
    'aetra-local-1',
    '/home/aetra/localnet',
    'condition: service_completed_successfully',
    '26657:26657',
    '9090:9090'
  )) {
  Assert-Contains -Text $composeText -Pattern ([regex]::Escape($term)) -Message "docker compose sample missing term: $term"
}

foreach ($term in @(
    'docker build',
    'version --long --output json',
    'docker inspect',
    'status --node tcp://127.0.0.1:26657',
    'aetra-node-ci',
    'docker volume create'
  )) {
  Assert-Contains -Text $workflowText -Pattern ([regex]::Escape($term)) -Message "workflow missing docker image term: $term"
}

Write-Host "docker image contract test passed"
