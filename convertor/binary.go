package convertor

import (
	"bytes"
	"context"

	"github.com/sagernet/sing-box/common/srs"
	boxConstant "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
)

var _ adapter.Convertor = (*RuleSetBinary)(nil)

type RuleSetBinary struct{}

func (s *RuleSetBinary) Type() string {
	return C.ConvertorTypeRuleSetBinary
}

func (s *RuleSetBinary) ContentType(_ adapter.ConvertOptions) string {
	return "application/octet-stream"
}

func (s *RuleSetBinary) From(ctx context.Context, content []byte, _ adapter.ConvertOptions) ([]adapter.Rule, error) {
	options, err := srs.Read(bytes.NewReader(content), true)
	if err != nil {
		return nil, err
	}
	return common.Map(options.Options.Rules, adapter.RuleFrom), nil
}

func (s *RuleSetBinary) To(ctx context.Context, contentRules []adapter.Rule, options adapter.ConvertOptions) ([]byte, error) {
	convertedRules, err := adapter.EmbedResourceRules(ctx, contentRules)
	if err != nil {
		return nil, err
	}
	ruleSet := &option.PlainRuleSetCompat{
		Version: boxConstant.RuleSetVersionCurrent,
		Options: option.PlainRuleSet{
			Rules: common.Map(common.Filter(convertedRules, func(it adapter.Rule) bool {
				return it.Headlessable()
			}), adapter.Rule.ToHeadless),
		},
	}
	if options.Metadata.Platform == C.PlatformSingBox && options.Metadata.Version != nil {
		Downgrade(ruleSet, options.Metadata.Version)
	}
	buffer := new(bytes.Buffer)
	err = srs.Write(buffer, ruleSet.Options, ruleSet.Version)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
