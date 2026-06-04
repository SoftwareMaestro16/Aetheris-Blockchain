package tx

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func AtomicStateChange(ctx context.Context, fn func(context.Context) error) error {
	cacheCtx, write := sdk.UnwrapSDKContext(ctx).CacheContext()
	if err := fn(cacheCtx); err != nil {
		return err
	}
	write()
	return nil
}
