package convertor

import (
	"bytes"
	"context"
	"strings"

	boxConstant "github.com/sagernet/sing-box/constant"
	boxOption "github.com/sagernet/sing-box/option"
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

func (s *RuleSetSource) ContentType() string {
	return "application/json"
}

func (s *RuleSetSource) From(ctx context.Context, binary []byte) (*boxOption.PlainRuleSetCompat, error) {
	if !strings.HasPrefix(string(binary), "{") {
		return nil, E.New("source is not a JSON object")
	}
	options, err := json.UnmarshalExtendedContext[boxOption.PlainRuleSetCompat](ctx, binary)
	if err != nil {
		return nil, err
	}
	return &options, nil
}

func (s *RuleSetSource) To(ctx context.Context, source *boxOption.PlainRuleSetCompat, options adapter.ConvertOptions) ([]byte, error) {
	if options.Metadata.Platform == C.PlatformSingBox && options.Metadata.Version != nil {
		Downgrade(source, options.Metadata.Version)
	}
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(source)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Downgrade(source *boxOption.PlainRuleSetCompat, version *semver.Version) {
	if version.LessThan(semver.ParseVersion("1.11.0")) {
		source.Version = boxConstant.RuleSetVersion2
		source.Options.Rules = common.Filter(source.Options.Rules, filter1100Rule)
	}
	if version.LessThan(semver.ParseVersion("1.10.0")) {
		source.Version = boxConstant.RuleSetVersion1
	}
}

func filter1100Rule(it boxOption.HeadlessRule) bool {
	return !hasRule([]boxOption.HeadlessRule{it}, func(it boxOption.DefaultHeadlessRule) bool {
		return len(it.NetworkType) > 0 || it.NetworkIsExpensive || it.NetworkIsConstrained
	})
}

func hasRule(rules []boxOption.HeadlessRule, cond func(rule boxOption.DefaultHeadlessRule) bool) bool {
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
