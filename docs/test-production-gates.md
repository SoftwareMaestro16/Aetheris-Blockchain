# Test And Production Gates

Track 10 defines the required testing layers, release gates, production gate,
immediate build order, and final economic architecture summary for Aetra.
It complements the prototype test pyramid and public testnet gate ledger.

## Unit Tests

Required unit coverage:

- fee formulas;
- inflation formulas;
- burn accounting;
- staking ratio controller;
- reputation scoring;
- domain pricing;
- resolver validation;
- memo validation;
- token/NFT/SBT storage rules.

Unit tests should stay fast, deterministic, and close to pure formulas or
state transition helpers. They must not require localnet or external services.

## Keeper Tests

Required keeper coverage:

- `x/fee`;
- `x/token`;
- `x/identity`;
- `x/resolver`;
- `x/reputation`;
- `x/execution`;
- `x/messaging`;
- `x/queue`;
- `x/events`;
- `x/actors`;
- `x/storage`.

Keeper tests must cover authority checks, signer checks, zero-address
rejection, `naet` fee policy, malformed input rejection, state writes, events,
genesis validation, export/import, and bounded iteration.

## Integration Tests

Required integration coverage:

- bank transfer with memo;
- resolver-based payment;
- domain auction to ownership;
- token creation and transfer;
- NFT mint and transfer;
- SBT mint and transfer rejection;
- async contract call;
- queue bounce/refund;
- reputation rate limit;
- dynamic fee under load.

Integration tests should prove cross-module behavior before e2e localnet
scripts are treated as acceptance evidence.

## E2E Smoke

Required e2e smoke coverage:

- 3-validator localnet;
- 5-validator localnet;
- staking lifecycle;
- fee distribution;
- domain lifecycle;
- resolver payment;
- AVM counter contract;
- AFT token transfer;
- ANFT mint/transfer;
- ASBT mint/prove/revoke;
- memo indexing;
- restart persistence;
- snapshot/state-sync.

E2E smoke scripts must run from clean state and must not require private
operator data, local-only secrets, or non-reproducible manual steps.

## Security Gates

Required security gates:

- `go test ./...`;
- `go vet ./...`;
- `buf lint`;
- deterministic execution gate;
- state export/import gate;
- govulncheck;
- gosec;
- gitleaks;
- CodeQL;
- dependency review;
- independent audit before production claim.

Security gates block public testnet or production when high/critical
fund-safety, consensus-safety, or secret-leak findings are untriaged.

## Production Gate

Production cannot be claimed until:

- long-running public testnet has no untriaged consensus or fund-safety issues;
- validator set can upgrade safely;
- staking, fees, DEX, AVM, domains, reputation, memo, and contract standards
  have adversarial tests;
- state export/import is deterministic;
- snapshot/state-sync works;
- emergency governance and halt/restart process tested;
- audit findings are triaged.

Sharding, partitioning, AVM production execution, and CosmWasm production use
remain excluded unless their own gates are explicitly passed.

## Immediate Build Order

1. Finish base-chain rename, address policy, and `naet` cleanup.
2. Finish base-chain safety and validation helpers.
3. Implement production fee formulas in `x/fee`.
4. Implement adaptive mint/burn controller in `x/token`.
5. Harden PoS/staking and distribution.
6. Add deterministic memo metadata support.
7. Add `x/reputation` with deterministic score only.
8. Add `.aet` domain registry and auction model.
9. Add resolver and resolver-based payment.
10. Build deterministic async queue without AVM.
11. Build minimal AVM with a counter contract.
12. Add actor model and messaging.
13. Implement AW-5 wallet.
14. Implement AFT-44 token master/wallet.
15. Implement ANFT-66 NFT collection/item.
16. Implement ASBT-67 soulbound item.
17. Add scheduler parallelism only after deterministic sequential async
    execution is stable.
18. Add compute/storage/market modules after baseline abuse controls exist.
19. Gate CosmWasm behind explicit config and tests.
20. Start partitioning/sharding simulator and spec only after async queue and
    AVM are audited.

## Final Economic Architecture Summary

Aetra economy is controlled by four feedback loops:

```text
staking participation -> adaptive inflation -> validator/delegator rewards
network activity      -> burn             -> supply pressure reduction
network load          -> soft fees/queues -> congestion control
account behavior      -> reputation       -> anti-spam and priority
```

The intended long-term behavior:

- low usage: mint rewards keep validators paid;
- normal usage: fees stay low and burn offsets part of mint;
- high usage: burn rises, queue controls activate, fees rise softly but stay
  capped;
- low staking: inflation rises within cap to attract staking;
- high staking: inflation falls to reduce dilution;
- spam: rate limits, reputation, deposits, and queue caps absorb abuse before
  fees become punitive.

This creates a self-regulating production L1 model where `AET` has uncapped
but bounded PoS supply, `naet` remains the only protocol fee asset, users keep
cheap transactions, and validators have a durable security budget.
