# Initial Release Non-Goals

Initial release should not attempt:

- PoH;
- Solana-level TPS;
- 1-second blocks;
- mandatory KYC validator admission;
- EVM at genesis unless separately approved;
- subjective slashing;
- unlimited validator set;
- unbounded contract execution;
- high inflation APR marketing.

## Scope Rule

These are hard scope boundaries for the initial release. Aetra initial release is a balanced BFT PoS L1, not a performance-first network and not a high-APR marketing product.

The initial release must keep:

- consensus on CometBFT BFT without PoH;
- block time targets above 1 second;
- validator participation open without mandatory KYC admission;
- bounded validator set growth;
- objective slashing only;
- bounded contract execution;
- EVM out of genesis scope unless separately approved;
- APR and inflation messaging conservative and evidence-based.

## Implementation Contract

The implementation gate is `app/params/initial_release_non_goals.go`.

Required catalog properties:

- `DefaultInitialReleaseScopePolicy` must not enable any non-goal;
- `ValidateInitialReleaseScope` must reject any initial release policy that attempts a non-goal;
- `BuildInitialReleaseNonGoalsReport` must report each violated non-goal;
- unit tests must cover performance, validator, execution, security, and marketing scope creep.
