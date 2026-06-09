# Aetra Validator Guide

This is the canonical validator operator guide for public testnet and later
launches.

## Hardware

Public validators should start from a medium server profile and increase only
after load testing proves the need.

Baseline guidance:

```text
CPU: 4-8 modern cores
RAM: 16-32 GB
Storage: NVMe SSD
Network: stable 100 Mbps+ with low packet loss
```

## OS

Linux is the preferred production environment. Windows is supported for local
tooling, build verification, and rehearsal scripts.

## Build Or Download Binary

Build from source or use the signed release artifact published for the target
network.

Build example:

```powershell
.\scripts\build-aetrad.ps1
build\aetrad.exe version --long --output json
```

Release operators should verify the release checksum and commit before first
start.

## Version Verification

Verify the binary version, commit, and dirty flag before joining a network:

```powershell
build\aetrad.exe version --long --output json
```

The verified version must match the published release notes and upgrade plan.

## Chain ID

Use the exact published chain ID. Do not invent a local testnet chain ID for a
public join.

## Genesis Validation

After downloading genesis, validate it locally before the first node start:

```powershell
build\aetrad.exe genesis validate-genesis $HOME\config\genesis.json --home $HOME
```

Genesis validation must pass before state sync, snapshot restore, or a
full-from-genesis start.

## Keyring

Use a secure keyring backend for validator operator keys.

```powershell
build\aetrad.exe keys add <key-name> --home $HOME --keyring-backend os
build\aetrad.exe keys show <key-name> -a --home $HOME --keyring-backend os
```

## Validator Key Safety

Never copy `priv_validator_key.json` to sentries or external hosts. Preserve
`priv_validator_state.json` across restarts and upgrades. Never publish keyring
files, mnemonics, or node keys.

## State Sync

State sync is the preferred fast join path when the network publishes a trust
height, trust hash, and supported RPC list.

Example:

```powershell
# edit $HOME\config\config.toml with trust height/hash and rpc_servers
build\aetrad.exe start --home $HOME
```

Operators must verify the trusted height and hash from the launch announcement.

## Snapshots

Publishers should provide snapshot height, archive checksum, and source
validator identity. Joiners should keep at least one recent snapshot and one
fallback snapshot available until after the next coordinated upgrade.

## Create Validator

Fund the validator account first, then create the validator with the published
chain ID and `naet` fees.

```powershell
$VAL_PUBKEY = build\aetrad.exe comet show-validator --home $HOME
build\aetrad.exe tx staking create-validator `
  --amount 100000000naet `
  --pubkey $VAL_PUBKEY `
  --moniker <moniker> `
  --chain-id $CHAIN_ID `
  --from <key-name> `
  --home $HOME `
  --keyring-backend os `
  --fees 1000000naet `
  --commission-rate 0.05 `
  --commission-max-rate 0.20 `
  --commission-max-change-rate 0.01 `
  --min-self-delegation 1 `
  --node tcp://127.0.0.1:26657 `
  -y
```

## Monitor

Monitor at least:

- block height;
- validator voting power;
- peer count;
- disk usage;
- process restart count;
- RPC, REST, and gRPC health;
- missed blocks and any jail/slash warnings.

Useful checks:

```powershell
build\aetrad.exe query staking validators --node tcp://127.0.0.1:26657 --output json
build\aetrad.exe status --node tcp://127.0.0.1:26657 --output json
```

## Restart

Restart with the same home directory. Preserve:

- `config\priv_validator_key.json`;
- `config\priv_validator_state.json`;
- `config\node_key.json`;
- `data\`;
- logs needed for incident review.

Do not roll back validator state to an earlier height unless the incident owner
has explicitly approved recovery steps.

## Upgrade

Follow the coordinated upgrade plan and rehearse it before public launch.
Upgrade handler names must match the announced plan name. Use
[docs/COSMOVISOR.md](COSMOVISOR.md) for Cosmovisor-managed nodes and
[docs/upgrade-playbook.md](upgrade-playbook.md) for rehearsal flow and
post-upgrade validation.

## Incident Response

If block production stalls, voting power diverges, or a validator key is
suspected compromised, follow the published incident response runbook and keep
evidence intact.

See [Testnet Incident Response](testnet-incident-response.md).
