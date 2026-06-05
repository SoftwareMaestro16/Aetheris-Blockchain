# Orbitalis zero address policy

Orbitalis reserves the zero address as a protocol-invalid address:

```text
raw:          0:0000000000000000000000000000000000000000000000000000000000000000
userfriendly: ORBAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
```

The zero address is forbidden by default. It must not be accepted as a signer,
admin, authority, recipient, fee collector, pool creator, liquidity provider,
swap actor, tokenfactory creator, tokenfactory admin, or genesis account.

Orbitalis does not currently define a zero-address burn sink. If a burn sink is
needed later, it must be introduced as an explicitly named protocol sink, with
state transitions that never require a private key or future action by that
address.

Custom modules must use the shared helpers in `app/addressing`:

- `ParseUserAddress`
- `ValidateUserAddress`
- `ParseAuthorityAddress`
- `ValidateAuthorityAddress`
- `ParseOptionalAdminAddress`
- `ValidateOptionalAdminAddress`
- `ValidateNoZeroFactoryDenomAdmin`

Genesis validation must reject zero address state even when the address also
passes generic Cosmos SDK address-format validation.
