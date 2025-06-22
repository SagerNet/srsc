package convertor

import (
	"bytes"
	"context"
	"strings"

	boxConstant "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/srsc/adapter"
	"github.com/sagernet/srsc/common/semver"
	C "github.com/sagernet/srsc/constant"
)

var _ adapter.Convertor = (*RuleSetSource)(nil)

type RuleSetSource struct{}

func (s *RuleSetSource) Type() string {
	return C.ConvertorTypeRuleSetSource
}

func (s *RuleSetSource) ContentType(_ adapter.ConvertOptions) string {
	return "application/json"
}

func (s *RuleSetSource) From(ctx context.Context, content []byte, _ adapter.ConvertOptions) ([]adapter.Rule, error) {
	if !strings.HasPrefix(string(content), "{") {
		return nil, E.New("source is not a JSON object")
	}
	options, err := json.UnmarshalExtendedContext[option.PlainRuleSetCompat](ctx, content)
	if err != nil {
		return nil, err
	}
	return common.Map(options.Options.Rules, adapter.RuleFrom), nil
}

func (s *RuleSetSource) To(ctx context.Context, contentRules []adapter.Rule, options adapter.ConvertOptions) ([]byte, error) {
	ruleSet := &option.PlainRuleSetCompat{
		Version: boxConstant.RuleSetVersionCurrent,
		Options: option.PlainRuleSet{
			Rules: common.Map(common.Filter(contentRules, func(it adapter.Rule) bool {
				return it.Headlessable()
			}), adapter.Rule.ToHeadless),
		},
	}
	if options.Metadata.Platform == C.PlatformSingBox && options.Metadata.Version != nil {
		Downgrade(ruleSet, options.Metadata.Version)
	}
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(ruleSet)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Downgrade(source *option.PlainRuleSetCompat, version *semver.Version) {
	if version.LessThan(semver.ParseVersion("1.11.0")) {
		source.Version = boxConstant.RuleSetVersion2
		source.Options.Rules = common.Filter(source.Options.Rules, filter1100Rule)
	}
	if version.LessThan(semver.ParseVersion("1.10.0")) {
		source.Version = boxConstant.RuleSetVersion1
	}
}

func filter1100Rule(it option.HeadlessRule) bool {
	return !hasRule([]option.HeadlessRule{it}, func(it option.DefaultHeadlessRule) bool {
		return len(it.NetworkType) > 0 || it.NetworkIsExpensive || it.NetworkIsConstrained
	})
}

func hasRule(rules []option.HeadlessRule, cond func(rule option.DefaultHeadlessRule) bool) bool {
	for _, rule := range rules {
		switch rule.Type {
		case boxConstant.RuleTypeDefault:
			if cond(rule.DefaultOptions) {
				return true
			}
		case boxConstant.RuleTypeLogical:
			if hasRule(rule.LogicalOptions.Rules, cond) {
				return true
			}
		}
	}
	return false
}
