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

## Endpoint Matrix

All examples use JSON output and the default localnet values above. Sample responses are intentionally minimal; real responses include additional Cosmos SDK fields such as pagination and consensus metadata.

| Query | CLI | gRPC method | REST path | Sample request | Sample response |
| --- | --- | --- | --- | --- | --- |
| Latest block | `build\orbitalisd.exe query block --node $NODE --output json` | `cosmos.base.tendermint.v1beta1.Service/GetLatestBlock` | `GET /cosmos/base/tendermint/v1beta1/blocks/latest` | `{}` | `{"block":{"header":{"height":"12","chain_id":"orbitalis-local-1"}}}` |
| Node info | `build\orbitalisd.exe status --node $NODE --output json` | `cosmos.base.tendermint.v1beta1.Service/GetNodeInfo` | `GET /cosmos/base/tendermint/v1beta1/node_info` | `{}` | `{"default_node_info":{"network":"orbitalis-local-1"}}` |
| Bank balance | `build\orbitalisd.exe query bank balance $NODE0 norb --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `cosmos.bank.v1beta1.Query/Balance` | `GET /cosmos/bank/v1beta1/balances/{address}/by_denom?denom=norb` | `{"address":"orb1...","denom":"norb"}` | `{"balance":{"denom":"norb","amount":"1000000"}}` |
| Bank balances | `build\orbitalisd.exe query bank balances $NODE0 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `cosmos.bank.v1beta1.Query/AllBalances` | `GET /cosmos/bank/v1beta1/balances/{address}` | `{"address":"orb1...","pagination":{"limit":"100"}}` | `{"balances":[{"denom":"norb","amount":"1000000"}]}` |
| Staking validators | `build\orbitalisd.exe query staking validators --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `cosmos.staking.v1beta1.Query/Validators` | `GET /cosmos/staking/v1beta1/validators?pagination.limit=100` | `{"pagination":{"limit":"100"}}` | `{"validators":[{"operator_address":"orbvaloper1...","status":"BOND_STATUS_BONDED"}]}` |
| Fees params | `build\orbitalisd.exe query fees params --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.fees.v1.Query/Params` | `GET /l1/fees/v1/params` | `{}` | `{"params":{"allowed_fee_denoms":["norb"],"validator_rewards_ratio":"0.98","community_pool_ratio":"0.02"}}` |
| Factory denoms | `build\orbitalisd.exe query tokenfactory denoms --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.tokenfactory.v1.Query/Denoms` | `GET /l1/tokenfactory/v1/denoms` | `{}` | `{"denoms":[{"denom":"factory/orb1.../gold","admin":"orb1..."}]}` |
| Factory denom | `build\orbitalisd.exe query tokenfactory denom $GOLD --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.tokenfactory.v1.Query/Denom` | `GET /l1/tokenfactory/v1/denom/{denom}` | `{"denom":"factory/orb1.../gold"}` | `{"metadata":{"denom":"factory/orb1.../gold","admin":"orb1..."}}` |
| DEX pools | `build\orbitalisd.exe query dex pools --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.dex.v1.Query/Pools` | `GET /l1/dex/v1/pools` | `{}` | `{"pools":[{"id":"1","denom0":"factory/orb1.../gold","denom1":"norb","lp_denom":"lp/1"}]}` |
| DEX pool | `build\orbitalisd.exe query dex pool 1 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.dex.v1.Query/Pool` | `GET /l1/dex/v1/pools/{pool_id}` | `{"pool_id":"1"}` | `{"pool":{"id":"1","reserve0":"10000000","reserve1":"10000000","lp_denom":"lp/1"}}` |

CometBFT RPC remains available for node-level checks that are not gRPC services:

| Query | Endpoint | Expected result |
| --- | --- | --- |
| RPC status | `GET http://127.0.0.1:26657/status` | `result.node_info.network = orbitalis-local-1` |
| RPC peers | `GET http://127.0.0.1:26657/net_info` | peer count for multi-validator localnet |
| RPC validator set | `GET http://127.0.0.1:26657/validators?per_page=100` | expected validator count with positive voting power |

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
