# Prototype Transaction Lifecycle Matrix

This matrix traces every tracked L1 prototype transaction from operator CLI input to final state query. Ecosystem-only flows are kept outside the tracked L1 contract.

Rules:

- Transaction behavior changes require module-boundary review and targeted tests in the row being changed.
- Public proto changes require the buf lint/generation workflow before merge.
- All example fees use `1000000norb`; `ORB`, factory denoms, and `testtoken` are not valid fee denoms.
- Transaction paths must use direct key lookups or bounded structures. No transaction path may require a full store scan.
- Custom module events are supporting evidence only; state queries remain the source of truth.

## Security Review Lens

The Cosmos-specific review for every row must cover:

- signer mismatch or missing authorization,
- invalid Bech32 account/operator addresses,
- invalid or spoofed denoms,
- insufficient funds or failed bank keeper calls,
- duplicate state and replay/sequence failure,
- malformed amount, param, or query fields,
- ABCI panic and nondeterministic state-write risk.

## Lifecycle Matrix

| Tx | Actor | Signer | CLI input | Funds and fee | State writes | Observable events | Verification queries | Tests |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Bank send | funded local account `node0` sends native funds to `node1` or another `orb1...` account | `--from node0`; SDK `MsgSend` signer must be sender | `tx bank send node0 $NODE1 1000norb --fees 1000000norb` | sender must hold `1000norb` plus fee; fee denom must be `norb` | `x/bank` sender/receiver balances and fee deduction | SDK tx/message and bank transfer events | `query bank balance $NODE1 norb`; `query tx <hash>` | `tests/e2e/native_token_smoke.ps1`, `tests/e2e/localnet_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1`, `x/fees/keeper/ante_test.go` |
| Staking delegate | local delegator bonds `norb` to a bonded validator | `--from node0`; SDK `MsgDelegate` delegator signer | `tx staking delegate $VALOPER 5000000norb --fees 1000000norb` | delegator must hold delegated `norb` plus fee; bond denom must be `norb` | `x/staking` delegation, shares, validator token/power state; bank bonded pool movement | SDK tx/message, staking delegation, and coin movement events | `query staking delegation $DELEGATOR $VALOPER`; `query staking validator $VALOPER`; CometBFT validator set/voting power | `app/pos_test.go`, `tests/e2e/pos_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Tokenfactory create denom | creator defines a new factory denom under own address | `--from node0`; `MsgCreateDenom.Creator` from CLI account | `tx tokenfactory create-denom gold --fees 1000000norb` | creator pays `norb` fee only; no token funds moved | `x/tokenfactory` denom authority metadata; bank denom metadata for `factory/<creator>/gold` | `tokenfactory_create_denom` plus SDK tx/message events | `query tokenfactory denom $DENOM`; `query bank denom-metadata $DENOM` | `x/tokenfactory/keeper/msg_server_test.go`, `x/tokenfactory/keeper/query_server_test.go`, `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Tokenfactory mint | current denom admin mints factory supply to a recipient | `--from <admin>`; `MsgMint.Sender` must equal current admin | `tx tokenfactory mint "1000000$DENOM" $TO --fees 1000000norb` | admin pays `norb` fee; module mints factory denom; recipient may be any valid `orb1...` | bank supply and recipient balance for factory denom | `tokenfactory_mint` plus SDK tx/message, bank mint, and transfer events | `query bank balance $TO $DENOM`; `query bank total-supply-of $DENOM` | `x/tokenfactory/keeper/msg_server_test.go`, `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Tokenfactory burn | current denom admin burns own factory balance | `--from <admin>`; `MsgBurn.Sender` must equal current admin and `burn_from_address` | `tx tokenfactory burn "250000$DENOM" $ADMIN --fees 1000000norb` | admin must hold burned amount plus `norb` fee | bank supply and admin balance decrease for factory denom | `tokenfactory_burn` plus SDK tx/message, bank send-to-module, and burn events | `query bank balance $ADMIN $DENOM`; `query bank total-supply-of $DENOM` | `x/tokenfactory/keeper/msg_server_test.go`, `tests/e2e/tokenfactory_smoke.ps1` |
| Tokenfactory change admin | current denom admin transfers authority | `--from <old-admin>`; `MsgChangeAdmin.Sender` must equal current admin | `tx tokenfactory change-admin $DENOM $NEW_ADMIN --fees 1000000norb` | old admin pays `norb` fee only | `x/tokenfactory` denom authority metadata admin field | `tokenfactory_change_admin` plus SDK tx/message events | `query tokenfactory denom $DENOM`; follow-up old-admin failure and new-admin mint success | `x/tokenfactory/keeper/msg_server_test.go`, `tests/e2e/tokenfactory_smoke.ps1` |
| Fees update params | governance authority updates fee policy params | governance module authority; no operator CLI business logic | gRPC/proposal path for `l1.fees.v1.MsgUpdateParams`; no local operator shortcut | governance tx/proposal fees in `norb`; params must validate before write | `x/fees` params store | `fees_update_params` plus SDK tx/message events | `query fees params`; REST `/l1/fees/v1/params` | `x/fees/keeper/msg_server_test.go`, `x/fees/types/genesis_test.go`, `x/fees/keeper/query_server_test.go`, `tests/e2e/fees_ante_smoke.ps1` |

## Negative And Adversarial Matrix

| Tx | Signer/auth failures | Field and denom failures | Balance/state failures | Replay/sequence coverage | Scale/scan note |
| --- | --- | --- | --- | --- | --- |
| Bank send | SDK rejects missing/wrong sender signature | invalid Bech32 receiver; invalid amount denom; wrong fee denom rejected by `x/fees` ante | insufficient funds fails in bank keeper without receiver mutation | SDK ante sequence; replay explicitly covered in PoS signed tx smoke and inherited by all SDK txs | direct account balance updates |
| Staking delegate | SDK requires delegator signature | malformed `orbvaloper`; wrong delegation denom | insufficient funds and invalid validator fail safely | `app/pos_test.go` and `tests/e2e/pos_smoke.ps1` cover signed replay/sequence | staking keeper direct validator/delegation keys |
| Tokenfactory create denom | creator signer from tx `--from`; malformed creator rejected | invalid subdenom, duplicate denom, native spoofing | duplicate denom state rejected before metadata write | SDK ante inherited; SHOULD FIX reusable replay e2e for custom module txs | denom key lookup; no denom list scan |
| Tokenfactory mint | non-admin sender rejected | missing denom, invalid amount, invalid recipient address | bank mint/send errors returned; supply checked by bank query | SDK ante inherited; SHOULD FIX reusable replay e2e for custom module txs | denom metadata lookup only |
| Tokenfactory burn | non-admin and burn-from mismatch rejected | missing denom, invalid amount, invalid burn address | insufficient balance fails on send-to-module; bank supply decreases only after successful transfer | SDK ante inherited; SHOULD FIX reusable replay e2e for custom module txs | denom metadata lookup only |
| Tokenfactory change admin | non-admin sender rejected | missing denom, invalid new admin | metadata write only after checks | SDK ante inherited; SHOULD FIX reusable replay e2e for custom module txs | denom metadata lookup only |
| Fees update params | invalid authority rejected | empty, duplicate, multi-denom, invalid denom params rejected | params write only after `Validate` | governance/SDK ante sequence inherited; no operator shortcut | single params key |

## Coverage Index

| Layer | Evidence |
| --- | --- |
| CLI construction | `cmd/l1d/cmd/root_test.go`, `x/tokenfactory/client/cli/tx.go`, `x/fees/client/cli/query.go` |
| Msg server authorization and state writes | `x/tokenfactory/keeper/msg_server_test.go`, `x/fees/keeper/msg_server_test.go`, `app/pos_test.go` |
| Fee ante policy | `x/fees/keeper/ante_test.go`, `tests/e2e/fees_ante_smoke.ps1` |
| Query verification | `x/tokenfactory/keeper/query_server_test.go`, `x/fees/keeper/query_server_test.go`, `tests/e2e/query_surface_smoke.ps1` |
| E2E lifecycle | `tests/e2e/native_token_smoke.ps1`, `tests/e2e/pos_smoke.ps1`, `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Determinism and audit | `scripts/security/determinism-gate.ps1`, `scripts/security/prototype-audit.ps1`, `docs/security/cosmos-security-checklist.md` |
| Bench/perf | `BenchmarkEmptyBlockFinalizeCommit` |

## Gaps

MUST FIX before public release or high-cardinality testnet:

- A reusable signed-tx replay/sequence e2e helper should be applied to bank and tokenfactory txs, not only the current staking/PoS smoke path.

SHOULD FIX for stronger operator observability:

- Add high-cardinality query pagination load evidence for tokenfactory before public explorer/API load testing.
- Add per-row transcript artifacts to release evidence so each lifecycle can be audited without rerunning the localnet.
- Add targeted CLI negative tests for malformed Bech32 and malformed coin arguments at command construction boundaries where Cosmos SDK validation does not already cover the path.

NICE TO HAVE:

- Cross-module invariant generator that runs tokenfactory mint/burn across many denoms/accounts.
- Multi-OS localnet lifecycle transcript beyond the current Windows-focused operator evidence.
