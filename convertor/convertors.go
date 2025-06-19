package convertor

import (
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/convertor/clash"
)

var Convertors = map[string]adapter.Convertor{
	C.ConvertorTypeRuleSetSource:     (*RuleSetSource)(nil),
	C.ConvertorTypeRuleSetBinary:     (*RuleSetBinary)(nil),
	C.ConvertorTypeAdGuardRuleSet:    (*AdGuardRuleSet)(nil),
	C.ConvertorTypeClashRuleProvider: (*clash.RuleProvider)(nil),
}
