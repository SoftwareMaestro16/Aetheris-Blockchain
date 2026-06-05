# Fee model hardening

Orbitalis v1 fee policy is intentionally narrow and deterministic.

## Fixed native fee denom

The only accepted transaction fee denom is `norb`. The `allowed_fee_denoms`
parameter is validated as exactly one value in v1:

```text
allowed_fee_denoms = ["norb"]
```

Factory denoms, LP denoms, display denom `ORB`, `testtoken`, and all other
denoms are rejected before the wrapped SDK ante handler can mutate state.

## Zero-fee policy

Localnet validator `minimum-gas-prices` can remain `0norb` for operator
convenience, but Orbitalis protocol policy still requires delivered
transactions to pay at least `1norb`.

The only v1 exception is height-0 genesis `MsgCreateValidator` transactions.
Public testnets should keep the protocol minimum at `>= 1norb` and may also set
non-zero validator min-gas-prices for mempool filtering.

## Bounded params

Governance cannot expand fee policy beyond the v1 bounds:

- allowed fee denom count: exactly `1`
- allowed fee denom value: `norb`
- `min_fee_amount`: positive integer, capped at `1000000000000000000`
- validator/community split ratios: each between `0` and `1`
- split ratio sum: exactly `1`
- fee collector module: `fee_collector`
- validator rewards target: `distribution/validator_rewards`
- community pool target: `protocolpool/community_pool`

Invalid params are rejected before state writes.

## Deduction and accounting model

The `x/fees` ante decorator runs before the SDK ante handler:

1. Reject missing, empty, malformed, below-minimum, and non-`norb` fees.
2. Call the wrapped SDK ante handler.
3. After SDK ante succeeds and deducts fees into `fee_collector`, record
   protocol fee accounting.

The v1 split model records:

- `98%` validator rewards target
- `2%` community pool target

Integer truncation applies to the community share, and the remainder goes to
validator rewards, so `total_collected == validator_rewards + community_pool`
always holds.
