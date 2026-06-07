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

Upgrade readiness is not optional for public testnet. A module with persistent state is not production-ready until its migration path, dry-run checklist, export/import behavior, and migration tests are present.
