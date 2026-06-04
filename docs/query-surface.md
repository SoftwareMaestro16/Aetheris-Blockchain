# Prototype Query Surface

Orbitalis prototype nodes must be observable through CLI, gRPC, REST gateway, and CometBFT RPC. The default localnet exposes node0 at:

```powershell
$NODE = "tcp://127.0.0.1:26657"
$GRPC = "127.0.0.1:9090"
$REST = "http://127.0.0.1:1317"
$HOME = ".localnet\node0\orbitalisd"
$FROM = "node0"
$KEYRING = "test"
$NODE0 = build\orbitalisd.exe keys show $FROM -a --home $HOME --keyring-backend $KEYRING
```

## Required Endpoints

| Surface | Query | Command or endpoint | Expected result |
| --- | --- | --- | --- |
| CometBFT RPC | node status | `Invoke-RestMethod http://127.0.0.1:26657/status` | `result.node_info.network = orbitalis-local-1` |
| CometBFT RPC | peers | `Invoke-RestMethod http://127.0.0.1:26657/net_info` | peer count for multi-validator localnet |
| CometBFT RPC | validator set | `Invoke-RestMethod http://127.0.0.1:26657/validators?per_page=100` | expected validator count with positive voting power |
| CLI | latest block | `build\orbitalisd.exe query block --node $NODE --output json` | `header.height` |
| REST | latest block | `Invoke-RestMethod "$REST/cosmos/base/tendermint/v1beta1/blocks/latest"` | `block.header.height` |
| REST | node info | `Invoke-RestMethod "$REST/cosmos/base/tendermint/v1beta1/node_info"` | `default_node_info.network` |
| CLI/REST | bank balance | `build\orbitalisd.exe query bank balance $NODE0 norb --node $NODE --output json` / `Invoke-RestMethod "$REST/cosmos/bank/v1beta1/balances/$NODE0"` | positive `norb` balance in the returned list |
| CLI/REST | staking validators | `build\orbitalisd.exe query staking validators --node $NODE --output json` / `Invoke-RestMethod "$REST/cosmos/staking/v1beta1/validators"` | validator list |
| CLI/gRPC/REST | fees params | `build\orbitalisd.exe query fees params --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` / `Invoke-RestMethod "$REST/l1/fees/v1/params"` | `allowed_fee_denoms` contains `norb` |
| CLI/gRPC/REST | factory denoms | `build\orbitalisd.exe query tokenfactory denoms --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` / `Invoke-RestMethod "$REST/l1/tokenfactory/v1/denoms"` | bounded denom list |
| CLI/gRPC/REST | factory denom | `build\orbitalisd.exe query tokenfactory denom $GOLD --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` / `Invoke-RestMethod "$REST/l1/tokenfactory/v1/denom/$GOLD"` | denom metadata and admin |
| CLI/gRPC/REST | DEX pools | `build\orbitalisd.exe query dex pools --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` / `Invoke-RestMethod "$REST/l1/dex/v1/pools"` | bounded pool list |
| CLI/gRPC/REST | DEX pool | `build\orbitalisd.exe query dex pool 1 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` / `Invoke-RestMethod "$REST/l1/dex/v1/pools/1"` | pool reserves and LP denom |

## Error Contract

Custom query servers return gRPC-status compatible errors:

- `InvalidArgument`: nil request, malformed denom, or `pool_id = 0`
- `NotFound`: valid tokenfactory denom or DEX pool id does not exist
- `ResourceExhausted`: unpaginated prototype list query exceeds the current cap

REST gateway maps these to HTTP status codes, for example `400` for invalid pool id and `404` for missing custom module objects.

## Bounded Lists

The current proto API has unpaginated list requests:

- `tokenfactory Denoms`
- `dex Pools`

For the prototype, query handlers cap each response at 100 items and return `ResourceExhausted` beyond that. Before public testnet or high-cardinality workloads, these RPCs are a MUST FIX for pagination/versioned replacement.

## Compatibility Policy

- Proto changes require the normal proto workflow: edit `.proto`, run generation, run `buf lint`, and review generated Go diffs.
- Generated `.pb.go` and `.pb.gw.go` files must not be edited manually.
- Breaking REST/gRPC path or field changes require versioning or an explicit compatibility exception.
- Query handlers must remain read-only: no state writes, no bank movement, and no business side effects.

## Security Notes

Query endpoints must not expose mnemonics, private keys, local environment secrets, or keyring material. Errors should be short status errors, not dumps of local process state. Iterator code must be deterministic and bounded on public query paths.

## Acceptance Check

```powershell
.\tests\e2e\query_surface_smoke.ps1
.\.work\tools\bin\buf.exe lint
```
