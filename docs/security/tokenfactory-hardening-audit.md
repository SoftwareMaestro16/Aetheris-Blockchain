# Tokenfactory Hardening Audit

Date: 2026-06-04

Scope:

- `x/tokenfactory/types`
- `x/tokenfactory/keeper`
- `x/tokenfactory/client/cli`
- `proto/l1/tokenfactory/v1/query.proto`

## Findings

- Subdenom validation accepted slash and colon, enabling ambiguous names such as nested factory denoms or LP-like subdenoms.
- `Query/Denoms` returned the full store without pagination, creating a state-bloat DoS surface for public RPC.
- Mint and burn authorization checked the admin, but did not assert bank supply deltas after module accounting.
- Admin transfer had no explicit renounce lifecycle. Empty admin is now a terminal renounced state.
- Genesis validation only checked a prefix and did not validate the canonical `factory/{creator}/{subdenom}` shape.
- Keeper read paths used `MustUnmarshal`; corrupted store bytes could panic instead of returning an error.

## Fixes

- Added a single factory denom validation path in `types`.
- Restricted subdenoms to letters, digits, dot, underscore, and dash; slash and colon are rejected.
- Reserved native-token-like subdenoms `norb` and `orb`, and LP-like `lp`, `lp-*`, `lp_*`, `lp.*`.
- Added bank metadata collision checks before denom creation.
- Added mint and burn supply delta assertions using `x/bank` supply before and after accounting.
- Added admin renounce by setting `new_admin` to an empty string; renounced denoms cannot mint, burn, or recover admin.
- Added bounded `Denoms` pagination with default limit `100` and max limit `500`.
- Replaced keeper `MustUnmarshal` read paths with explicit error returns.

## Compatibility

- Existing valid denoms of the form `factory/{canonical-bech32-creator}/{subdenom}` remain valid when their subdenom matches the stricter safe character set.
- Admin may differ from the denom creator after `ChangeAdmin`; genesis import/export preserves that state.
- Empty admin is valid only as a renounced lifecycle state.

## Residual Risk

- `ExportGenesis` and internal `GetAllDenoms` still iterate all denoms. This is acceptable for operator/export paths, but public queries must use `Denoms` pagination.
- Mint and burn still rely on SDK tx cache rollback if a downstream bank operation errors after an earlier state write. Handlers validate addresses and amounts before writes to minimize this surface.
- Further work can add simulation invariants that compare tokenfactory registry state against bank metadata and supply at app level.
