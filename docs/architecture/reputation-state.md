# Reputation State

`x/reputation` is the planned deterministic reputation module for anti-spam,
scheduler weighting, and progressive account limits. It is currently a pure
types package only; no SDK store, keeper, module account, genesis state, CLI, or
ABCI hook is registered.

Reputation is never a fee-market auction replacement and cannot be directly
purchased. Reputation staking may exist only as a bonded signal with
slashing/risk. Scores are based only on deterministic on-chain events.

## State

```text
ReputationRecord {
  account: address
  score: uint8 // 0..100
  age_score
  staking_score
  tx_success_score
  volume_score
  domain_score
  contract_score
  spam_penalty
  failed_tx_penalty
  slash_penalty
  last_updated
}
```

## Score Levels

```text
0-20   restricted
20-50  new
50-80  normal
80-95  trusted
95-100 elite
```

The implementation treats lower bounds as inclusive except the first bucket:
`0..19 restricted`, `20..49 new`, `50..79 normal`, `80..94 trusted`,
`95..100 elite`.

## Deterministic Score

```text
score =
  age_score
  + staking_time_score
  + tx_success_rate_score
  + bounded_volume_score
  + domain_reputation_score
  + contract_reputation_score
  - spam_penalty
  - failed_tx_penalty
  - slash_events_penalty
```

Bounds:

```text
score = clamp(score, 0, 100)
```

Domain ownership can add bounded reputation, but cannot dominate score.
Contract reputation is bounded separately. Contracts also have reputation
records and progressive limits.

Decay:

```text
if inactive:
  score -= inactivity_decay_rate * inactive_epochs
```

In code, decay begins only after the configured inactivity threshold and clamps
at zero.

## Progressive Limits

New accounts have progressive limits. These limits are intentionally separate
from fee price escalation:

- restricted: very small tx, gas, and async queue allowance;
- new: limited but usable allowance;
- normal: default user allowance;
- trusted: larger allowance for active users and contracts;
- elite: highest bounded allowance, still capped.

These limits can feed `x/fees` admission checks, async queue scheduling, and
future mempool prioritization. They must not bypass `naet` fee validation,
signer checks, zero-address rejection, sequence replay protection, or module
authorization.

## Reputation Usage

Anti-spam:

- low score means lower tx rate limit;
- low score means lower async queue quota;
- low score means higher memo/storage byte cost;
- low score means stricter contract deploy limits.

Execution priority:

- high score can improve queue priority within deterministic bounds;
- priority cannot bypass fees, signatures, or validation;
- validators must compute identical priority ordering.

Access control:

- token creation may require score threshold or deposit;
- contract deployment may require score threshold or deposit;
- DEX pool creation may require score threshold or deposit;
- domain auction spam can be rate-limited by score.

Usage rules:

- high-score users still must pay required protocol fees;
- reputation priority is a bounded integer weight with tx-index tie-breaker;
- deposits can satisfy access gates but do not increase score directly;
- failed contract executions add deterministic penalties;
- successful contract executions can increase bounded contract reputation;
- reputation usage must never bypass base transaction validation.
