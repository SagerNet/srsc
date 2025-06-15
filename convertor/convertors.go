package convertor

import (
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
)

var Convertors = map[string]adapter.Convertor{
	C.ConvertorTypeRuleSetSource:         (*RuleSetSource)(nil),
	C.ConvertorTypeRuleSetBinary:         (*RuleSetBinary)(nil),
	C.ConvertorTypeAdGuardRuleSet:        (*AdGuardRuleSet)(nil),
	C.ConvertorTypeClashTextRuleProvider: (*ClashTextRuleProvider)(nil),
	C.ConvertorTypeClashYamlRuleProvider: (*ClashYamlRuleProvider)(nil),
	C.ConvertorTypeMetaRuleSetBinary:     (*MetaRuleSetBinary)(nil),
}
