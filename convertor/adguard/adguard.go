package adguard

import (
	"bytes"
	"context"

	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
)

var _ adapter.Convertor = (*RuleSet)(nil)

type RuleSet struct{}

func (a *RuleSet) Type() string {
	return C.ConvertorTypeAdGuardRuleSet
}

func (a *RuleSet) ContentType(options adapter.ConvertOptions) string {
	return "text/plain"
}

func (a *RuleSet) From(ctx context.Context, content []byte, options adapter.ConvertOptions) ([]adapter.Rule, error) {
	if options.Options.AdGuardOptions.AcceptExtendedRules && options.Options.TargetType != C.ConvertorTypeAdGuardRuleSet && options.Options.TargetType != C.ConvertorTypeRuleSetBinary {
		return nil, E.New("AdGuard rule-set can only be converted to sing-box rule-set binary when `accept_extended_rules` enabled")
	}
	return ToRules(bytes.NewReader(content), options.Options.AdGuardOptions.AcceptExtendedRules, logger.NOP())
}

func (a *RuleSet) To(ctx context.Context, contentRules []adapter.Rule, options adapter.ConvertOptions) ([]byte, error) {
	return FromRules(contentRules)
}
