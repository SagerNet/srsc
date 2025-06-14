package cache

import (
	"context"

	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/option"
)

func New(ctx context.Context, options option.CacheOptions) (adapter.Cache, error) {
	switch options.Type {
	case C.CacheTypeMemory, "":
		return NewMemory(options.Timeout), nil
	case C.CacheTypeRedis:
		return NewRedis(ctx, options.Timeout, options.RedisOptions)
	default:
		return nil, E.New("unknown cache type: ", options.Type)
	}
}
