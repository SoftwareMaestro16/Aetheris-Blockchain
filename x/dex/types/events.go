package types

const (
	EventTypeCreatePool        = "dex_create_pool"
	EventTypeAddLiquidity      = "dex_add_liquidity"
	EventTypeRemoveLiquidity   = "dex_remove_liquidity"
	EventTypeSwapExactAmountIn = "dex_swap_exact_amount_in"

	AttributeKeyPoolID       = "pool_id"
	AttributeKeyCreator      = "creator"
	AttributeKeyDepositor    = "depositor"
	AttributeKeyWithdrawer   = "withdrawer"
	AttributeKeyTrader       = "trader"
	AttributeKeyDenom0       = "denom0"
	AttributeKeyDenom1       = "denom1"
	AttributeKeyAmount0      = "amount0"
	AttributeKeyAmount1      = "amount1"
	AttributeKeyLPDenom      = "lp_denom"
	AttributeKeyMintedShares = "minted_shares"
	AttributeKeyShares       = "shares"
	AttributeKeyTokenIn      = "token_in"
	AttributeKeyTokenOut     = "token_out"
)
