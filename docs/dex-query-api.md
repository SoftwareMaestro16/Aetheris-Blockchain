# DEX Query API Notes

Date: 2026-06-04

## Pagination

`Query/Pools` is wire-compatible with previous clients: callers may still send an empty `QueryPoolsRequest`, but the response is now bounded.

- Default page size: `100` pools.
- Maximum page size: `500` pools.
- Maximum offset: `500` pools.
- `pagination.next_key` should be passed back as `pagination.key` for the next page.
- `pagination.reverse` is rejected for v1.
- `pagination.count_total` is rejected for v1 because it requires a full pool-store scan.
- `pagination.key` and `pagination.offset` cannot be used together.

CLI:

```powershell
build\orbitalisd.exe query dex pools --limit 100 --node tcp://127.0.0.1:26657
```

REST:

```text
GET /l1/dex/v1/pools?pagination.limit=100
```

## Pair Lookup

`Query/PoolByPair` resolves the canonical pair through a deterministic KV index. It does not scan all pools.

CLI:

```powershell
build\orbitalisd.exe query dex pool-by-pair norb uatom --node tcp://127.0.0.1:26657
```

REST:

```text
GET /l1/dex/v1/pool_by_pair?denom_a=norb&denom_b=uatom
```

## Security And Scale Notes

DEX queries expose pool metadata and reserves only; they do not expose account private data, keys, mnemonics, or local node files.

`count_total` is intentionally disabled because a correct total requires scanning every pool. Offset pagination is capped for the same reason. High-cardinality clients should use `next_key` pagination or `PoolByPair`.

## Operational Note

The pair index is populated during pool creation and genesis import. Existing stores created before this index need a migration before enabling pair lookup on a live network.
