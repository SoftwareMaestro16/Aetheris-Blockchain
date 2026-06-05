# PoS and staking correctness

This document records the Phase 4 Orbitalis PoS/staking acceptance policy.

## Native staking denom

Orbitalis staking uses only `norb` as the bond denom. Validator creation,
delegation, unbonding, redelegation, rewards, slashing, and localnet genesis
tests must not rely on `stake`, `uatom`, display denom `ORB`, or factory denoms.

## Validator lifecycle coverage

Unit tests in `app/pos_test.go` cover:

- validator creation with `norb` self-delegation
- delegation increasing validator tokens and consensus power
- unbonding entries and delayed balance return
- redelegation entries between bonded validators
- invalid delegation denom, funds, and address rejection
- slashing params and downtime missed-block bitmap persistence
- delegator reward withdrawal through the distribution module

Integration tests in `tests/integration/pos_lifecycle_test.go` cover:

- signed staking tx delivery through ante and `FinalizeBlock`
- validator-set updates returned to CometBFT after staking power changes
- staking delegation state surviving export/import restart

## Local 3-validator acceptance

The localnet acceptance path is:

```powershell
.\scripts\build-orbitalisd.ps1
.\tests\e2e\pos_smoke.ps1 -Binary .\build\orbitalisd.exe
```

`tests/e2e/pos_smoke.ps1` validates a 3-validator local network by default:

- initializes and validates localnet genesis
- starts the network and waits for blocks
- checks CometBFT validator count and bonded staking validators
- confirms staking bond denom `norb`
- checks slashing params and signing info
- submits a delegation and confirms total voting power increases
- confirms invalid delegation paths fail
- confirms replayed signed tx fails

`scripts/localnet/validate-genesis.ps1` is the genesis-specific guard for the
3-node localnet. It asserts matching genesis hashes across nodes, exactly three
gentxs by default, `MsgCreateValidator` self-delegation in `norb`, and empty
prototype ecosystem state for tokenfactory and dex genesis.
