package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidDenom      = errorsmod.Register(ModuleName, 2, "invalid denom")
	ErrUnauthorized      = errorsmod.Register(ModuleName, 3, "unauthorized")
	ErrDenomExists       = errorsmod.Register(ModuleName, 4, "denom already exists")
	ErrDenomMissing      = errorsmod.Register(ModuleName, 5, "denom not found")
	ErrInvalidAddress    = errorsmod.Register(ModuleName, 6, "invalid address")
	ErrSupplyInvariant   = errorsmod.Register(ModuleName, 7, "supply invariant violation")
	ErrInvalidPagination = errorsmod.Register(ModuleName, 8, "invalid pagination")
)
