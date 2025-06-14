package adapter

import (
	"context"

	boxOption "github.com/sagernet/sing-box/option"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/option"
)

type Convertor interface {
	Type() string
	ContentType(options ConvertOptions) string
	From(ctx context.Context, binary []byte, options ConvertOptions) (*boxOption.PlainRuleSetCompat, error)
	To(ctx context.Context, source *boxOption.PlainRuleSetCompat, options ConvertOptions) ([]byte, error)
}

type ConvertOptions struct {
	Options  option.ConvertOptions
	Metadata C.Metadata
}
