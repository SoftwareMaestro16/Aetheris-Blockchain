# Data Migration and Upgrade Strategy

Aetra is expected to evolve. Upgrade safety is part of the architecture.

Every persistent module must define how state is versioned, migrated, validated, exported, imported, and rolled forward during a coordinated chain upgrade. Upgrade logic is consensus-critical: a migration that is non-deterministic, partially applied, missing from the version map, or not replay-safe can halt the network or fork app hash.

## Required Strategy

Required upgrade strategy:

- module consensus versions are explicit;
- migrations are registered before public testnet;
- every migration has deterministic state transformations;
- every migration validates old state before writing new state;
- every migration validates new state after writing;
- export/import after upgrade must produce valid genesis;
- app hash after restart must remain stable;
- dry-run upgrade procedure must be documented;
- rollback limits must be documented;
- unsafe downgrade behavior must be rejected;
- upgrade handlers must emit stable events;
- tests must cover migration success and failure paths.

## 31.1 Upgrade Requirements

Every new module or state-breaking change must include:

- store key decision;
- genesis import/export;
- migration handler;
- version map update;
- upgrade test;
- rollback notes where possible;
- operator instructions.

The store key decision must state whether the module introduces a new store key, reuses an existing key, migrates prefixes, or intentionally remains stateless. Genesis import/export must be implemented before public testnet so state can be replayed, audited, and recovered. Migration handlers must be registered in the upgrade path and must update the module version map exactly once.

Rollback notes are required even when rollback is limited. If rollback is not possible after a state write, the note must say so explicitly and point operators to the supported recovery path. Operator instructions must include binary version, upgrade height, expected halt behavior, restart procedure, and post-upgrade checks.

The implementation gate is `DefaultAetraUpgradeStrategyEvidence` in `app/params/data_migration_upgrade_strategy.go`.

## 31.2 Migration Tests

Required tests:

- old genesis imports into new binary;
- migration initializes params;
- migration preserves balances;
- migration preserves staking state;
- migration preserves slashing state;
- migration preserves contract state if applicable;
- app hash after migration is deterministic.

These tests must be automated where feasible. If a test is not feasible in unit scope, it must move to integration, simulation, or localnet upgrade testing and remain tracked as a production gate. Manual notes are not enough for balances, staking, slashing, contract state, or app hash determinism.

## Module Requirements

Every Aetra module with persistent state must provide:

- current consensus version;
- migration handlers from every supported previous version;
- version map sanity checks;
- genesis validation for migrated state;
- export/import compatibility tests;
- deterministic replay tests;
- invariant tests after migration;
- documentation of removed, renamed, or transformed fields.

Modules without persistent state must explicitly document that no migration handler is required.

## Upgrade Plan Requirements

Every network upgrade plan must define:

- upgrade name;
- target height;
- binary version;
- expected module version map before upgrade;
- expected module version map after upgrade;
- state migration list;
- parameter changes if any;
- operator runbook;
- halt/restart expectations;
- post-upgrade smoke tests;
- public rollback boundary.

The rollback boundary must be honest: after a migration writes new state and blocks continue, rollback may require social coordination and state export tooling rather than a simple binary downgrade.

## Tests

Required tests:

- default genesis upgrade no-op where applicable;
- migration from previous version to current version;
- invalid old state rejected;
- invalid new state rejected;
- missing module version rejected;
- future module version rejected;
- export/import after migration;
- restart after migration;
- app hash stability after migration;
- invariant checks after migration;
- event emission for upgrade handler execution.

## Acceptance Gate

Upgrade readiness is not optional for public testnet. A module with persistent state is not production-ready until its migration path, dry-run checklist, export/import behavior, and migration tests are present. `BuildAetraUpgradeStrategyReport` must pass for every new persistent module and every state-breaking change before that work can be considered complete.
