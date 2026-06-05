# Phase 9 Aether Core Wiring Gate Security Notes

## Scope

Phase 9 wires accepted prototype modules into Aether Core as disabled-by-default
SDK modules:

- `x/load`
- `x/routing`
- `x/zones`
- `x/mesh`

The wiring adds store keys, keepers, module manager registration, default
genesis, validation, export, and migration skeletons. It does not add public Msg
services, contract execution, production sharding, or cross-zone settlement.

## Routing Decision Point

Routing is fixed to `ANTE_ADMISSION_ONLY` for this gate. That means routing is an
auditable admission/classification spec and is not executed in
`PrepareProposal`, `ProcessProposal`, `FinalizeBlock`, or a production Msg
server. A later coordinated upgrade must explicitly change this policy before
routing can mutate consensus state.

## Consensus Safety

- Prototype feature gates are disabled in default genesis.
- Prototype modules have no module account permissions.
- Prototype modules do not implement BeginBlocker or EndBlocker.
- BeginBlocker and EndBlocker order lists include prototype module names as
  explicit no-op lifecycle positions.
- Aether Core stores only prototype genesis/config state at this phase.
- Aether Core does not execute smart contracts or application logic.
- Contract-zone execution remains target architecture, not live behavior.

## Security Audit Notes

- No randomness, wall-clock time, goroutines, external API calls, or local
  latency inputs were added to app lifecycle paths.
- Store-backed genesis import/export validates state before write and after
  read.
- Export/import state remains deterministic for disabled modules.
- Query response bounds remain in the prototype keeper layer.
- Migration handlers are no-op validators from consensus version `1` to `2`.

## Remaining Production Gates

- Add protobuf Msg/Query contracts only after API review.
- Add prefix-bounded KV iteration before public queries.
- Add production routing persistence only through governance and software
  version gate.
- Re-run determinism, export/import, localnet restart, and long-run testnet
  checks before enabling any mutating prototype feature.
- Keep public docs wording as `experimental sharding` until the production gate
  passes.
