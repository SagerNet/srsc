package source

import (
	"context"

	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/option"
)

func New(ctx context.Context, options option.SourceOptions) (adapter.Source, error) {
	switch options.Source {
	case C.EndpointSourceLocal:
		return NewLocal(ctx, options)
	case C.EndpointSourceRemote:
		return NewRemote(ctx, options)
	default:
		return nil, E.New("unknown source type: " + options.Source)
	}
}
