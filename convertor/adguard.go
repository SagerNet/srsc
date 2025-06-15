package convertor

import (
	"bytes"
	"context"

	"github.com/sagernet/sing-box/common/convertor/adguard"
	boxConstant "github.com/sagernet/sing-box/constant"
	boxOption "github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
)

var _ adapter.Convertor = (*AdGuardRuleSet)(nil)

type AdGuardRuleSet struct{}

func (a *AdGuardRuleSet) Type() string {
	return C.ConvertorTypeAdGuardRuleSet
}

func (a *AdGuardRuleSet) ContentType(options adapter.ConvertOptions) string {
	return "text/plain"
}

func (a *AdGuardRuleSet) From(ctx context.Context, binary []byte, options adapter.ConvertOptions) (*boxOption.PlainRuleSetCompat, error) {
	rules, err := adguard.ToOptions(bytes.NewReader(binary), logger.NOP())
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

func (a *AdGuardRuleSet) To(ctx context.Context, source *boxOption.PlainRuleSetCompat, options adapter.ConvertOptions) ([]byte, error) {
	return adguard.FromOptions(source.Options.Rules)
}
