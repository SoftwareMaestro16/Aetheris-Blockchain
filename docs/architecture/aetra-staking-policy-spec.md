# x/aetra-staking-policy Module Specification

Purpose: control effective voting power, delegation overflow, commission policy, and anti-concentration incentives.

This module is the central anti-centralization module of Aetra.

## Responsibilities

The module must:

- calculate raw validator stake;
- calculate effective validator stake;
- calculate overflow stake;
- enforce or expose effective voting power cap;
- calculate reward multiplier for overflow stake;
- expose delegation concentration warnings;
- enforce commission floor;
- enforce max commission;
- enforce max commission change rate;
- expose top-N concentration metrics;
- validate governance param changes;
- emit events for cap/overflow/commission policy changes;
- remain deterministic and export/import safe.

## Production Rule

`x/aetra-staking-policy` is not complete when only cap math exists. The production definition is:

```text
staking policy = stake math + cap enforcement/exposure + overflow accounting + commission policy + concentration metrics + governance params + events + export/import safety + tests + docs
```

Every responsibility must be represented in code, genesis/governance parameter validation, query surface, events where state changes, and tests. If a responsibility is temporarily query-only instead of enforcement, the behavior must be explicit and covered by tests so the chain does not silently present fake anti-centralization guarantees.

## Implementation Contract

The implementation gate is `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `AetraStakingPolicyModuleName` must be `x/aetra-staking-policy`;
- `DefaultAetraStakingPolicySpecEvidence` must cover the module purpose and central anti-centralization role;
- `BuildAetraStakingPolicySpecReport` must require all responsibilities from this document;
- `ValidateAetraStakingPolicySpec` must reject incomplete evidence;
- missing stake math, cap, overflow, reward multiplier, concentration warnings, commission controls, top-N metrics, governance validation, policy-change events, or deterministic export/import safety must fail validation;
- module identity must fail validation when missing or not equal to `x/aetra-staking-policy`.

## Required Tests

The module specification tests must prove:

- the default evidence covers all responsibilities;
- raw stake, effective stake, overflow stake, effective voting power cap, and overflow reward multiplier are mandatory;
- delegation concentration warnings, commission floor, max commission, max commission change rate, top-N concentration metrics, governance param validation, policy events, and export/import safety are mandatory;
- purpose and central anti-centralization role are mandatory;
- wrong or missing module identity is rejected.
