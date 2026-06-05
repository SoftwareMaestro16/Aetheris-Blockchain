# Memo / Note System

The memo system allows optional human-readable text on transactions without
letting metadata affect consensus execution logic or bloat state.

## Core Fields

```text
TxMetadata {
  memo: string optional
  memo_hash: bytes optional
  memo_visible: bool
}
```

Supported surfaces:

- native bank transfer;
- resolver/domain payment;
- token transfer;
- NFT transfer;
- SBT proof/revoke;
- contract call;
- domain auction bid;
- domain renewal;
- DEX swap/liquidity action.

Requirements:

- optional;
- UTF-8 only;
- max length default: `200` characters;
- hard max memo chars: `500`;
- default max memo bytes: `1024`;
- max length governance configurable within hard protocol bound;
- immutable after block inclusion;
- stored as transaction metadata, not execution input;
- cannot change state transition result.

Validation:

```text
if memo == "":
  ok
else:
  require valid_utf8(memo)
  require char_count(memo) <= max_memo_chars
  require byte_len(memo) <= max_memo_bytes
  require no prohibited control chars
```

`memo_hash` is optional. If it is present with a visible memo, it must be a
32-byte SHA-256 hash of the memo bytes. Hash-only metadata is allowed for hidden
notes but still cannot affect execution.

## Memo Economics

Recommended cost:

```text
memo_fee =
  memo_base_fee
  + memo_byte_fee * memo_bytes
  * reputation_multiplier(sender)
  * congestion_multiplier(load)
```

Reputation multiplier:

```text
score >= 80: 0.75
score >= 50: 1.00
score >= 20: 1.50
score < 20:  3.00
```

Rules:

- memo fee paid only in `naet`;
- memo fee can be zero for empty memo;
- memo size contributes to tx byte cost;
- low reputation can be rate-limited or delayed;
- memo cannot become cheap spam storage;
- memo fee and memo validation must run before inclusion, but memo text is not
  passed into keeper state transition logic.

## Implementation Boundary

`x/memo/types` is a pure metadata validation and fee specification package. It
does not register SDK stores, keepers, module accounts, genesis state, CLI, or
ABCI hooks. Future wiring must keep memo data immutable after block inclusion
and must prove through tests that changing memo text cannot change execution
results.

## Storage And Indexing

Store:

- tx hash;
- sender;
- receiver if known;
- asset type;
- related domain if any;
- memo or memo hash depending on storage policy;
- block height;
- timestamp.

Indexes:

- by tx hash;
- by sender;
- by receiver;
- by domain;
- by contract;
- by asset;
- by event type.

Privacy option:

- allow full memo stored on-chain for transparency; or
- store memo hash on-chain and full memo in indexer only.

Production default:

- store bounded memo on-chain only if below configured byte limit;
- indexer may provide search;
- consensus does not depend on search index.

Events:

```text
EventMemoAttached {
  tx_hash
  from
  to
  domain
  memo_hash
  memo
}
```

The memo event is deterministic and derived from the stored memo projection.
Empty memo is accepted. Valid memo is accepted. Invalid UTF-8 and oversized
memo are rejected before indexing. Low reputation memo cost is higher through
the memo economics multiplier.
