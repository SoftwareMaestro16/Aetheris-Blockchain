# Observability and Public Metrics

Aetra observability must cover consensus health, validator behavior, economics, slashing, AVM execution, transaction failures, and node sync status. Metrics are part of public network trust: validators, delegators, explorers, indexers, governance participants, and incident responders must be able to verify network behavior without private operator access.

## Required Metrics

Required metrics:

- block time;
- finality latency;
- missed blocks;
- validator uptime;
- validator concentration;
- top-10/top-20/top-33 voting power;
- inflation;
- bonded ratio;
- estimated APR;
- burned fees;
- treasury balance;
- slashing events;
- jail/unjail events;
- contract execution gas;
- failed tx reasons;
- node sync status.

Each required metric must be available through public or operator-safe surfaces. Prometheus alone is not enough; public users need query and indexer-compatible access too.

## Required Surfaces

Required surfaces:

- CLI queries;
- gRPC queries;
- REST queries where applicable;
- Prometheus metrics;
- explorer/indexer compatibility events;
- public testnet dashboards.

CLI, gRPC, and REST are used for direct verification. Prometheus is used for operator alerting and dashboards. Explorer/indexer compatibility events are used for public network visibility, historical analysis, and user-facing wallets. Public testnet dashboards are required before a broad validator or delegator campaign.

## Prometheus Metric Names

Required Prometheus names:

```text
block time: aetra_block_time_seconds
finality latency: aetra_finality_latency_seconds
missed blocks: aetra_validator_missed_blocks_total
validator uptime: aetra_validator_uptime_bps
validator concentration: aetra_validator_concentration_bps
top-10/top-20/top-33 voting power: aetra_validator_top_n_power_bps
inflation: aetra_economy_inflation_bps
bonded ratio: aetra_economy_bonded_ratio_bps
estimated APR: aetra_economy_estimated_apr_bps
burned fees: aetra_economy_burned_fees_naet
treasury balance: aetra_economy_treasury_balance_naet
slashing events: aetra_slashing_events_total
jail/unjail events: aetra_validator_jail_events_total and aetra_validator_unjail_events_total
contract execution gas: aetra_contract_execution_gas
failed tx reasons: aetra_failed_tx_reasons_total
node sync status: aetra_node_sync_status
```

Labels must be bounded and redacted. Do not use addresses, tx hashes, contract addresses, pool ids, denoms outside an allowlist, free-form error strings, or user-controlled metadata as labels without explicit cardinality review.

## Event and Query Requirements

Explorer/indexer compatibility events must be deterministic and stable across releases. Event names and attributes should be documented and versioned when they represent public protocol behavior.

Required query behavior:

- CLI queries must expose current values or point to the exact module query;
- gRPC queries must be the canonical machine-readable query surface;
- REST queries should be available through gRPC gateway where applicable;
- failed tx reasons must use bounded reason codes;
- slashing and jail/unjail events must include bounded reason, validator identity, height, and amount when applicable;
- top-N voting power must support top-10, top-20, and top-33 views;
- node sync status must distinguish catching up from caught up.

## Dashboard Requirements

Public testnet dashboards must include:

- consensus health panel: block time, finality latency, node sync status;
- validator reliability panel: missed blocks, uptime, jail/unjail events, slashing events;
- decentralization panel: validator concentration and top-10/top-20/top-33 voting power;
- economics panel: inflation, bonded ratio, estimated APR, burned fees, treasury balance;
- VM and tx panel: contract execution gas and failed tx reasons.

Dashboards are not consensus-critical, but dashboard readiness is a public testnet gate because operators and delegators need shared visibility during incidents.

## Implementation Contract

The implementation catalog is `DefaultPublicMetricSpecs` in `observability/public_metrics.go`.

Required catalog properties:

- every required metric has a Prometheus definition;
- every required metric is exposed through CLI, gRPC, REST where applicable, Prometheus, indexer-compatible events, and public dashboards;
- all labels are bounded;
- explorer/indexer compatibility is explicit;
- public testnet dashboard readiness is explicit;
- missing surfaces fail readiness tests.
