package adapter

import (
	"context"

	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/option"
)

type Convertor interface {
	Type() string
	ContentType(options ConvertOptions) string
	From(ctx context.Context, content []byte, options ConvertOptions) ([]Rule, error)
	To(ctx context.Context, contentRules []Rule, options ConvertOptions) ([]byte, error)
}

type ConvertOptions struct {
	Options  option.ConvertOptions
	Metadata C.Metadata
}
