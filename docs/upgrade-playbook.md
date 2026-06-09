# Aetra Upgrade Playbook

This playbook is the operational checklist for coordinated Aetra testnet and
mainnet upgrades.

## 1. Pre-Proposal Checks

- Choose a unique upgrade name matching the binary handler.
- Pick an upgrade height, not wall-clock time.
- Verify the upgrade height leaves enough blocks for validators to deploy the
  binary.
- Run a dry-run migration from the latest exported state.
- Confirm `StoreUpgrades` lists every added, renamed, or deleted store. Use an
  explicit empty list when no store layout changes are expected.
- Confirm `ModuleManager.GetVersionMap()` includes every active `x/` module.
- Confirm every module version bump has a registered migration handler.
- Export after migration and validate genesis.

Required local checks:

```powershell
go test ./app/upgrades ./app -run "Upgrade|ModuleVersion|PreBlocker|CustomModuleMigrations" -count=1
go test ./app/keeperwiring ./app/modulewiring ./app/wiring/... -count=1
```

Run broader CI before public scheduling:

```powershell
go test ./...
```

## 2. Governance Proposal

Submit an SDK `MsgSoftwareUpgrade` proposal with:

- `name`: exact handler name;
- `height`: planned halt/upgrade height;
- `info`: release tag, binary checksum, migration summary, and rollback notes.

The proposal must be rejected or withdrawn if:

- the plan name does not match a registered handler;
- the height is zero or already passed;
- the version-map dry run fails;
- store migrations are not documented;
- export/import after migration fails.

## 3. Validator Rollout

Before halt height:

- validators download and verify the new binary checksum;
- validators keep the old binary available for rollback before the halt height;
- validators keep `priv_validator_state.json` intact;
- sentries and RPC nodes are upgraded according to operator policy;
- at least one snapshot/export is retained before the upgrade height.

At halt height:

- old binaries stop at the scheduled upgrade;
- operators replace the binary;
- operators restart with the same home directory;
- the new binary runs the upgrade handler and migrations.

### 3.1 Cosmovisor Path

When the chain is supervised by Cosmovisor, the same plan still applies:

- keep the active home directory unchanged across the restart;
- write the upgrade plan to the node's `data/upgrade-info.json` file before
  restart or let Cosmovisor persist it through the normal upgrade flow;
- keep the replacement binary available under the Cosmovisor-managed upgrade
  slot for the plan name, or use the canonical no-op rehearsal handler
  `rehearsal-noop` when the binary does not change;
- restart the service from the same home directory so Cosmovisor can relaunch
  the daemon at the scheduled height;
- export the state after the upgrade and validate the exported genesis again.

## 4. Post-Upgrade Validation

After restart:

- verify blocks resume;
- verify app hash agreement across validators;
- query the stored module version map;
- verify the upgrade is marked done;
- check validator set and staking queries;
- check fee collector, treasury, burn, and protection fund accounting;
- check AVM and contract query surfaces if touched by the release;
- export state after a few blocks and validate it.

## 5. Mainnet Extra Gates

Mainnet upgrades require:

- public testnet rehearsal with the same handler name or a clearly equivalent
  dry-run handler;
- signed release checksums;
- validator communication window;
- incident response owner;
- rollback decision point before halt height;
- post-upgrade monitoring dashboard.

Mainnet must not use `--unsafe-skip-upgrades` except as an explicitly documented
emergency recovery action approved by governance or the incident process.
