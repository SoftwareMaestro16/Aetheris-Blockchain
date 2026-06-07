# Implementation Phases

This plan defines the first implementation phases for Aetra Core. Each phase has tasks, deliverables, tests, and acceptance criteria. A phase is not complete because code exists; it is complete only when its tests and evidence are present.

## Phase 0 - Baseline Audit

Tasks:

- inspect current Cosmos SDK and CometBFT versions;
- document current app module graph;
- identify existing modules overlapping with `aetra-staking-policy`, `aetra-validator-score`, and `aetra-economics`;
- decide which modules are renamed, reused, or wrapped;
- verify current `naet` staking denom;
- verify fee collector, burn, treasury, emissions, mint authority wiring;
- verify current localnet scripts and test coverage.

Deliverables:

- module inventory;
- gap analysis;
- risk list;
- updated implementation checklist.

Tests:

- current full unit test run;
- current integration test run;
- current localnet smoke test;
- current export/import test.

Acceptance:

- module inventory exists and maps current app, keeper, store, and module-account ownership;
- overlapping modules have a decision: rename, reuse, wrap, or replace;
- `naet` staking denom and fee denom assumptions are verified;
- fee collector, burn, treasury, emissions, and mint authority wiring have explicit evidence;
- localnet and export/import scripts are known-good or have tracked blockers.

## Phase 1 - Staking Policy and Validator Cap

Tasks:

- implement effective voting power cap;
- implement overflow stake accounting;
- implement commission floor/max/max-change policy;
- add concentration metrics;
- add queries for validator raw/effective/overflow stake;
- add governance params with validation;
- wire module into app lifecycle.

Tests:

- cap math unit tests;
- validator set transition tests;
- concentration query tests;
- commission bounds tests;
- integration tests with staking;
- export/import tests;
- invariant tests.

Acceptance:

- no validator can exceed configured effective power cap;
- excess stake does not increase voting power;
- params cannot be set outside safe bounds;
- state remains deterministic after export/import.

## Implementation Contract

The implementation phase gate is `app/params/implementation_phases.go`.

Required catalog properties:

- `DefaultImplementationPhasePlans` must include Phase 0 and Phase 1;
- Phase 0 must include baseline audit tasks, deliverables, and current test evidence;
- Phase 1 must include staking cap tasks, tests, and acceptance checks;
- phase items require evidence;
- missing tasks, deliverables, tests, or acceptance checks fail validation.
