# Fee Model Hardening

Aetra v1 fee policy is intentionally narrow and deterministic.

## Fixed Native Fee Denom

The only accepted transaction fee denom is `naet`. The `allowed_fee_denoms`
parameter is validated as exactly one value in v1:

```text
allowed_fee_denoms = ["naet"]
```

Factory denoms, token wallets, DEX LP denoms, NFT/SBT assets, display denom
`AET`, `testtoken`, and all other denoms are rejected before the wrapped SDK
ante handler can mutate state. Owning the non-native token does not make it a
valid protocol fee.

Gasless or user-friendly flows must use relayers that pay `naet` on-chain. Any
other token collection belongs off-chain or inside a separate contract flow and
does not bypass base-chain fee policy.

## Zero-Fee Policy

Localnet validator `minimum-gas-prices` can remain `0naet` for operator
convenience, but Aetra protocol policy still requires delivered
transactions to pay at least `1naet`.

The only v1 exception is height-0 genesis `MsgCreateValidator` transactions.
Public testnets should keep the protocol minimum at `>= 1naet` and may also set
non-zero validator min-gas-prices for mempool filtering.

## Bounded Params

Governance cannot expand fee policy beyond the v1 bounds:

- allowed fee denom count: exactly `1`
- allowed fee denom value: `naet`
- `min_fee_amount`: positive integer, capped at `1000000000000000000`
- `base_fee_amount`: positive integer, `>= min_fee_amount`
- `max_fee_amount`: positive integer, `>= base_fee_amount`, capped at
  `1000000000000000000`
- target utilization: `1..9999` bps
- congestion threshold: greater than target and `<= 10000` bps
- `max_tx_gas <= max_block_gas`
- block tx, sender tx, and stake-weighted sender tx limits must be positive
- fee/stake priority weights must sum to `10000` bps
- validator/community split ratios: each between `0` and `1`
- split ratio sum: exactly `1`
- fee collector module: `fee_collector`
- validator rewards target: `distribution/validator_rewards`
- community pool target: `protocolpool/community_pool`

Invalid params are rejected before state writes. Valid governance updates also
synchronize `community_pool_ratio` into `x/distribution` as `community_tax`.

## Deduction And Accounting Model

The `x/fees` ante decorator runs before the SDK ante handler:

1. Reject zero-address fee participants, missing, empty, malformed,
   below-minimum, over-cap, and non-`naet` fees.
2. Compute the dynamic required fee from deterministic block gas utilization.
3. Enforce per-tx gas, per-block gas, per-block tx count, and sender rate
   limits.
4. Call the wrapped SDK ante handler.
5. After SDK ante succeeds and deducts fees into `fee_collector`, record
   protocol fee accounting.

The v1 split model records:

- `98%` validator rewards target
- `2%` community pool target

Integer truncation applies to the community share, and the remainder goes to
validator rewards, so `total_collected == validator_rewards + community_pool`
always holds. Accounting state only supports `naet` in v1.

## Low-Fee Congestion Policy

The dynamic fee curve is capped and intentionally revenue-suboptimal:

```text
required_fee = base_fee + ceil((max_fee - base_fee) * over_target^2 / room^2)
required_fee = min(required_fee, max_fee)
```

Normal traffic at or below target utilization pays the base fee. Extreme
congestion can never exceed the hard cap. Transactions above the hard cap are
invalid, and fee overpayment does not increase priority.

Spam is controlled by protocol admission limits instead of unbounded fee
escalation:

- fixed max tx gas;
- fixed max block gas;
- fixed max tx count per block;
- rolling sender count per block;
- stake-weighted allowance and priority formulas for PrepareProposal/mempool
  integration.
