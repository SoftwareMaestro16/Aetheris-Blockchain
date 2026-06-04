# Module Account And Bank Movement Audit

This inventory covers tracked custom L1 modules. Ecosystem-only modules are outside the tracked release artifact.

## Module Accounts

| Module | Permissions | Purpose | Blocked recipient |
| --- | --- | --- | --- |
| `tokenfactory` | `Minter`, `Burner` | Mint and burn factory denoms under admin control. | Yes |
| `fees` | none | Fee policy params and ante policy state. | Yes |

## Bank Movements

| Function | Module | Movement | Validation before movement | State ordering | Error handling | Tests |
| --- | --- | --- | --- | --- | --- | --- |
| `Mint` | tokenfactory | `MintCoins(tokenfactory, amount)` then `SendCoinsFromModuleToAccount(tokenfactory, mint_to, amount)`. | Existing denom, current admin signer, recipient Bech32, positive valid coin. | No tokenfactory state write. Event emitted in cached block after bank success. | All bank errors checked; cached block prevents minted supply leak if recipient send fails. | `TestTokenfactoryCreateMintBurnAdminFlow`, `TestMintToBlockedModuleAddressDoesNotLeakSupply`, unauthorized admin tests. |
| `Burn` | tokenfactory | `SendCoinsFromAccountToModule(burn_from, tokenfactory, amount)` then `BurnCoins(tokenfactory, amount)`. | Existing denom, current admin signer, `burn_from == sender`, positive valid coin. | No tokenfactory state write. Event emitted in cached block after bank success. | All bank errors checked; cached block prevents token custody changes if burn fails. | `TestAdminCanBurnOwnFactoryTokens`, unauthorized burn tests. |

## Audit Decisions

- No unchecked custom bank movement errors remain in `x/tokenfactory`.
- `tokenfactory` is the only tracked custom module with mint/burn permissions.
- Failed bank movement must leave custom module metadata unchanged and must not emit success events.
- Future modules must add an inventory row before adding mint, burn, send, escrow, or reserve custody behavior.
