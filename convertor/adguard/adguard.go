package adguard

import (
	"bytes"
	"context"

	boxConstant "github.com/sagernet/sing-box/constant"
	boxOption "github.com/sagernet/sing-box/option"
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

func (a *RuleSet) From(ctx context.Context, binary []byte, options adapter.ConvertOptions) (*boxOption.PlainRuleSetCompat, error) {
	if options.Options.AdGuardOptions.AcceptExtendedRules && options.Options.TargetType != C.ConvertorTypeAdGuardRuleSet && options.Options.TargetType != C.ConvertorTypeRuleSetBinary {
		return nil, E.New("AdGuard rule-set can only be converted to sing-box rule-set binary with `accept_extended_rules` enabled")
	}
	rules, err := ToOptions(bytes.NewReader(binary), options.Options.AdGuardOptions.AcceptExtendedRules, logger.NOP())
	if err != nil {
		return nil, err
	}
	return &boxOption.PlainRuleSetCompat{
		Version: boxConstant.RuleSetVersionCurrent,
		Options: boxOption.PlainRuleSet{
			Rules: rules,
		},
	}, nil
}

func (a *RuleSet) To(ctx context.Context, source *boxOption.PlainRuleSetCompat, options adapter.ConvertOptions) ([]byte, error) {
	return FromOptions(source.Options.Rules)
}
