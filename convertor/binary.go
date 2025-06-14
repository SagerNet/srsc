package convertor

import (
	"bytes"
	"context"

	"github.com/sagernet/sing-box/common/srs"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
)

var _ adapter.Convertor = (*RuleSetBinary)(nil)

type RuleSetBinary struct{}

func (s *RuleSetBinary) Type() string {
	return C.ConvertorTypeRuleSetBinary
}

func (s *RuleSetBinary) ContentType() string {
	return "application/octet-stream"
}

func (s *RuleSetBinary) From(ctx context.Context, binary []byte) (*option.PlainRuleSetCompat, error) {
	options, err := srs.Read(bytes.NewReader(binary), true)
	if err != nil {
		return nil, err
	}
	return &options, nil
}

func (s *RuleSetBinary) To(ctx context.Context, source *option.PlainRuleSetCompat, options adapter.ConvertOptions) ([]byte, error) {
	if options.Metadata.Platform == C.PlatformSingBox && options.Metadata.Version != nil {
		Downgrade(source, options.Metadata.Version)
	}
	buffer := new(bytes.Buffer)
	err := srs.Write(buffer, source.Options, source.Version)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
