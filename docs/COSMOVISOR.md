# Aetra Cosmovisor Guide

Cosmovisor is the recommended upgrade supervisor for validator operators who
want automatic binary switching at a scheduled height.

## Install

Install Cosmovisor from the official Cosmos SDK tooling or the release
environment approved by the operator team. Verify the installed version before
using it in production.

## Directory Layout

Cosmovisor expects a stable home directory layout:

```text
$DAEMON_HOME/
  cosmovisor/
    genesis/bin/aetrad
    upgrades/<upgrade-name>/bin/aetrad
  data/upgrade-info.json
  config/
  data/
```

The active home directory must remain stable across restarts and upgrades.

## Current Binary

Cosmovisor runs the currently active binary from the `genesis` or
`upgrades/<name>` directory. Keep the binary checksum and release notes aligned
with the announced plan.

## Upgrades Directory

Place the new binary under the upgrade slot that matches the approved handler
name. The upgrade name and the handler name must be identical.

Example names:

- `v053-to-v054`
- `rehearsal-noop`

The `rehearsal-noop` handler is for dry-run and rehearsal only.

## Environment

Use explicit environment variables:

```text
DAEMON_HOME=<validator-home>
DAEMON_NAME=aetrad
DAEMON_ALLOW_DOWNLOAD_BINARIES=false
DAEMON_RESTART_AFTER_UPGRADE=true
```

Keep the same home directory in the process manager, service unit, or shell
wrapper that launches Cosmovisor.

## Upgrade Handler Naming

The plan name, handler name, and Cosmovisor upgrade directory must match. If
the name differs, the node will not switch binaries at the intended height.

Rehearsal runs should use the canonical no-op handler name rather than a
production binary switch.

## Rollback Policy

Keep the previous binary available until after the upgrade is validated.

Rollback rules:

- do not delete the old binary before validation;
- preserve the same home directory on restart;
- if the new binary fails, restore the previous binary and restart from the
  same home;
- never use `--unsafe-skip-upgrades` except as an explicitly documented
  emergency recovery action approved by governance or incident response.

After any successful upgrade or rehearsal, export the state and validate the
exported genesis again.
