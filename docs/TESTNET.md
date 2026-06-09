# Aetra Testnet Reference

This document is the canonical public testnet summary for operators and wallet
integrators.

## Chain ID

The launch announcement publishes the active chain ID. Use the exact announced
value for all txs, queries, snapshots, and state-sync commands.

Example placeholder:

```text
aetra-<network>-<height>
```

## AWCE-1 Wallet Compatibility Summary

- user-facing addresses are `AE...`;
- raw/internal/proof addresses are `4:...`;
- user staking is pool/index based;
- normal users do not choose validators directly;
- direct user delegation to validators is disabled;
- private keys and seed phrases are never accepted in user-facing docs, logs,
  or genesis data.

## Genesis URL And Checksum

Publish both the genesis URL and the SHA256 checksum before validators join.

Example placeholders:

```text
Genesis URL: https://<release-host>/genesis.json
Genesis SHA256: <sha256>
Genesis height: <height>
```

Every operator should validate the downloaded genesis before first start.

## Seed Nodes And Persistent Peers

The launch announcement must publish:

- seed node addresses;
- persistent peer addresses;
- `docs/testnet/peers.example.json` for JSON-based peer metadata templates;
- `docs/testnet/seeds.example.txt` for CometBFT `node_id@host:port` seed
  lines;
- any RPC peers required for state sync;
- the port profile for the public testnet launch.

Operators should keep validator peers explicit and avoid copying untrusted peer
sets into private validator nodes.

## RPC Endpoints

Publish the canonical RPC, REST, and gRPC endpoints for the network.

Example placeholders:

```text
RPC:  tcp://<host>:26657
REST: http://<host>:1317
gRPC: 127.0.0.1:9090
```

Validator operators should keep a local health check and avoid depending on a
single external RPC endpoint.

## Faucet Path

If a faucet is enabled, it must be off-chain and rate limited. The chain does
not mint faucet funds through a native faucet module.

Publish:

- faucet URL or command path;
- rate limit policy;
- max grant per request;
- cooldown window;
- key custody and incident contact.

## Minimum Fees

The launch announcement publishes the minimum required tx fee policy in `naet`.
Prototype and launch examples should use the configured floor from genesis or
the release notes, not an ad hoc token denom.

## Expected Block Time

Publish the expected block time, commit timeout, and any validator-set caveats.
Operators should measure health by height progression, not wall-clock optimism.

## Launch Profile

The launch profile must state the validator count and rehearsal tier:

- `ValidatorProfile 3` for fast CI smoke;
- `ValidatorProfile 4` and `ValidatorProfile 5` for launch rehearsal;
- `ValidatorProfile All` only when every published profile is required.

## Known Non-Goals

Public testnet v1 does not promise:

- direct user delegation to validators;
- native token/NFT/DEX application modules;
- production faucet custody;
- sharding without an explicit governance gate;
- mainnet security or uptime guarantees.

Token, NFT, and DEX behavior belongs in AVM contracts and standards, not in
native app asset modules.
