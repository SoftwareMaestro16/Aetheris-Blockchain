# Account and transaction safety

This document records the Phase 2 Orbitalis transaction safety policy.

## Account creation paths

Runtime account creation is handled through the Cosmos SDK account keeper and
standard ante flow. Orbitalis adds protocol guards around address parsing and
genesis import:

- Auth genesis accounts are rejected when the account address is the Orbitalis
  zero address.
- `SimGenesisAccount.Validate` rejects nil base accounts and zero addresses.
- Custom modules parse actor fields with `app/addressing` helpers, so empty,
  malformed Bech32, malformed raw, malformed userfriendly, and zero addresses
  cannot become module actors.

## Replay and signer safety

All signed transactions run through the SDK ante handler after the Orbitalis fee
policy wrapper. SDK ante verifies account number, account sequence, chain ID in
the sign bytes, signer/public-key binding, signature count, signature gas, and
then increments sequence only after signature verification succeeds.

Acceptance evidence:

- Same signed bank tx succeeds once and replay fails with the stale sequence.
- A tx signed for a different chain ID fails before balance mutation.
- A tx signed by an account that is not the message signer fails before balance
  mutation.
- Malformed protobuf tx bytes produce failed tx results without fee accounting.

## Ante handler order

`app/handlers.go` installs:

1. `x/fees` ante decorator.
2. Cosmos SDK `x/auth/ante.NewAnteHandler`.

The `x/fees` decorator runs first:

1. At block height `0`, allow genesis `MsgCreateValidator` txs through.
2. Require the tx to implement `sdk.FeeTx`.
3. Load fee params.
4. Require valid `norb` fees meeting `min_fee_amount`.
5. Call the wrapped SDK ante handler.
6. Only after wrapped ante success and non-simulation, record protocol fee
   accounting.

The SDK ante handler order in Cosmos SDK v0.54.3 is:

1. `SetUpContext`
2. `ExtensionOptions`
3. `ValidateBasic`
4. `TxTimeoutHeight`
5. `ValidateMemo`
6. `ConsumeGasForTxSize`
7. `DeductFee`
8. `SetPubKey`
9. `ValidateSigCount`
10. `SigGasConsume`
11. `SigVerification`
12. `IncrementSequence`

Because Orbitalis records protocol fee accounting only after the wrapped SDK
ante succeeds, missing fees, invalid fee denoms, insufficient fee funds, invalid
signatures, wrong chain ID, and stale sequence cannot update protocol fee
accounting or message state.

## Required tests

Run:

```powershell
.\.work\tools\go1.25.11\go\bin\go.exe test .\tests\integration
.\.work\tools\go1.25.11\go\bin\go.exe test .\x\fees\...
.\.work\tools\go1.25.11\go\bin\go.exe test .\...
```
