package types

const (
	EventTypeCreateDenom = "tokenfactory_create_denom"
	EventTypeMint        = "tokenfactory_mint"
	EventTypeBurn        = "tokenfactory_burn"
	EventTypeChangeAdmin = "tokenfactory_change_admin"

	AttributeKeyDenom           = "denom"
	AttributeKeyCreator         = "creator"
	AttributeKeyAdmin           = "admin"
	AttributeKeySender          = "sender"
	AttributeKeyAmount          = "amount"
	AttributeKeyMintToAddress   = "mint_to_address"
	AttributeKeyBurnFromAddress = "burn_from_address"
	AttributeKeyNewAdmin        = "new_admin"
)
