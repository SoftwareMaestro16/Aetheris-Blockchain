# Genesis And Params Model

## Global Defaults

All chain constants that affect state transitions must be represented in genesis or module params. No consensus-critical value should be hardcoded in keeper logic.

Initial Orbitalis chain choices:
- Address prefix: `orb`
- Native base denom: `norb`
- Display denom: `ORB`
- Native token name: `Orbitalis`
- Native token decimals: `9`
- Governance authority: the `x/gov` module account or SDK authority configured at genesis.

## `x/tokenfactory`

Params:
- `denom_creation_fee`
- `max_denom_name_len`
- `max_metadata_len`
- `minting_enabled`
- `burning_enabled`

Genesis state:
- `params`
- `denoms`
- `admins`
- `metadata`

Validation:
- Denom IDs must be canonical and collision-free.
- Admin addresses must decode with the chain address codec.
- Existing supply must match bank state when imported.

## `x/dex`

Params:
- `swap_fee_bps`
- `protocol_fee_bps`
- `max_fee_bps`
- `min_initial_liquidity`
- `max_pools`
- `pool_creation_fee`

Genesis state:
- `params`
- `pools`
- `lp_positions`
- `next_pool_id`

Validation:
- Pool IDs must be unique and monotonic.
- Asset pairs must be canonical sorted pairs.
- Reserves and LP supply must be positive for active pools.
- Fee values must be bounded by `max_fee_bps`.

## `x/fees`

Params:
- `allowed_fee_denoms`
- `validator_rewards_ratio`
- `community_pool_ratio`
- `min_fee_amount`
- `fee_collector_module`
- `validator_rewards_target`
- `community_pool_target`

Genesis state:
- `params`
- `protocol_fee_state`

Validation:
- v1 allows only `norb` fees.
- Minimum fee must be positive.
- Split ratios must be decimals between `0` and `1` and sum exactly to `1`.
- Fee collector and target routes are explicit and fixed to safe v1 module routes.
- Protocol fee accounting must satisfy `total_collected == validator_rewards + community_pool`.

## Upgrade Policy

Every module with persistent state must include a migrations package once implementation begins. Initial migrations may be no-ops, but the upgrade path must exist before testnet launch.
