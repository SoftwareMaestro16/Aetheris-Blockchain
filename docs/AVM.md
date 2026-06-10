# Aetra AVM Guide

AVM v1 is the canonical contract execution environment for Aetra application
logic.

## Canonical Bytecode And Module Format

AVM modules use a canonical bytecode format with a stable module header,
deterministic serialization, and explicit versioning. The compiler, verifier,
and runtime must agree on the same module identity.

The canonical implementation lives under:

- `x/aetravm/avm`
- `x/aetravm/messageabi`
- `x/aetravm/async`
- `x/aetravm/chunk`
- `x/aetravm/standards`
- `x/contracts`

## Verifier

Every module must pass verification before deployment or execution. The
verifier rejects:

- malformed bytecode;
- unsupported versions;
- forbidden opcodes;
- forbidden host calls;
- over-large modules or instruction data;
- nondeterministic execution features.

## Instruction Set And Gas Schedule

AVM instructions are stack-based and deterministic. Gas is charged from the
published gas schedule, not from wall clock time, local mempool contents, or
randomness.

The runtime must produce the same gas usage for the same code, state, and
message sequence.

## Typed Values

Typed values are part of the module and message ABI. The runtime must not rely
on untyped host-dependent interpretation.

## Deterministic Stack Execution

Execution must be reproducible across nodes:

- same code;
- same state;
- same input message;
- same block height;
- same gas limit;
- same host function allowlist.

The same input must produce the same exit code, state root, receipts, and
outgoing messages.

## Chunk And ChunkMap Persistent State

AVM persistent state is stored in deterministic chunked structures. Export and
import must preserve state roots exactly.

Chunk iteration and export paths must stay bounded and deterministic.

## Deploy, Execute, And Get Methods

AVM must support:

- deployment;
- message execution;
- get/query methods.

Get methods must not mutate state or emit consensus messages.

## Receipts, Events, And Proofs

Execution should emit deterministic receipts, events, and proofs. These must
be stable enough for export/import, indexing, and verification tooling.

## Host Function Allowlist

Only explicitly approved host functions are allowed. Forbidden host behavior
includes nondeterministic time sources, random sources, ambient network access,
and unordered iteration over external state.

## Examples In CI

The CI surface should compile and run the minimal AVM examples:

- counter;
- domain registry;
- asset ledger;
- internal message;
- bounce/refund;
- get methods.

Recommended checks:

```powershell
go test ./x/aetravm/avm ./x/aetravm/async ./x/aetravm/messageabi ./x/aetravm/standards/... ./x/contracts
```

## Non-Goals

AVM is the runtime for application logic, not a native asset module system.
Application-level asset behavior belongs in AVM contracts and standards.
