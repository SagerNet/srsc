package option

import (
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badjson"
	C "github.com/sagernet/srsc/constant"
)

type ConvertOptions struct {
	SourceConvertOptions
	TargetConvertOptions
}

func (o ConvertOptions) MarshalJSON() ([]byte, error) {
	return badjson.MarshallObjects(o.SourceConvertOptions, o.TargetConvertOptions)
}

func (o *ConvertOptions) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &o.SourceConvertOptions)
	if err != nil {
		return err
	}
	return badjson.UnmarshallExcludedMulti(bytes, &o.SourceConvertOptions, &o.TargetConvertOptions)
}

type _SourceConvertOptions struct {
	SourceType     string                         `json:"source_type,omitempty"`
	AdGuardOptions AdGuardRuleSetSourceOptions    `json:"-"`
	ClashOptions   ClashRuleProviderSourceOptions `json:"-"`
	SurgeOptions   SurgeRuleProviderSourceOptions `json:"-"`
}

type SourceConvertOptions _SourceConvertOptions

func (o SourceConvertOptions) MarshalJSON() ([]byte, error) {
	var v any
	switch o.SourceType {
	case C.ConvertorTypeRuleSetSource, C.ConvertorTypeRuleSetBinary:
	case C.ConvertorTypeAdGuardRuleSet:
		v = o.AdGuardOptions
	case C.ConvertorTypeClashRuleProvider:
		v = o.ClashOptions
	case C.ConvertorTypeSurgeRuleSet:
		v = o.SurgeOptions
	case "":
		return nil, E.New("missing source type")
	default:
		return nil, E.New("unknown source type: " + o.SourceType)
	}
	if v == nil {
		return json.Marshal((_SourceConvertOptions)(o))
	}
	return badjson.MarshallObjects((_SourceConvertOptions)(o), v)
}

func (o *SourceConvertOptions) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, (*_SourceConvertOptions)(o))
	if err != nil {
		return err
	}
	var v any
	switch o.SourceType {
	case C.ConvertorTypeRuleSetSource, C.ConvertorTypeRuleSetBinary:
	case C.ConvertorTypeAdGuardRuleSet:
		v = &o.AdGuardOptions
	case C.ConvertorTypeClashRuleProvider:
		v = &o.ClashOptions
	case C.ConvertorTypeSurgeRuleSet:
		v = &o.SurgeOptions
	case "":
		return E.New("missing source type")
	default:
		return E.New("unknown source type: " + o.SourceType)
	}
	if v == nil {
		return nil
	}
	return json.Unmarshal(bytes, v)
}

type _TargetConvertOptions struct {
	TargetType   string                         `json:"target_type,omitempty"`
	ClashOptions ClashRuleProviderTargetOptions `json:"-"`
	SurgeOptions SurgeRuleProviderTargetOptions `json:"-"`
}

type TargetConvertOptions _TargetConvertOptions

func (o TargetConvertOptions) MarshalJSON() ([]byte, error) {
	var v any
	switch o.TargetType {
	case C.ConvertorTypeRuleSetSource, C.ConvertorTypeRuleSetBinary:
	case C.ConvertorTypeClashRuleProvider:
		v = o.ClashOptions
	case C.ConvertorTypeSurgeRuleSet:
		v = o.SurgeOptions
	case "":
		return nil, E.New("missing target type")
	default:
		return nil, E.New("unknown target type: " + o.TargetType)
	}
	if v == nil {
		return json.Marshal((_TargetConvertOptions)(o))
	}
	return badjson.MarshallObjects((_TargetConvertOptions)(o), v)
}

func (o *TargetConvertOptions) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, (*_TargetConvertOptions)(o))
	if err != nil {
		return err
	}
	var v any
	switch o.TargetType {
	case C.ConvertorTypeRuleSetSource, C.ConvertorTypeRuleSetBinary:
	case C.ConvertorTypeClashRuleProvider:
		v = &o.ClashOptions
	case C.ConvertorTypeSurgeRuleSet:
		v = &o.SurgeOptions
	case "":
		return E.New("missing target type")
	default:
		return E.New("unknown target type: " + o.TargetType)
	}
	if v == nil {
		return nil
	}
	return badjson.UnmarshallExcluded(bytes, (*_TargetConvertOptions)(o), v)
}

type AdGuardRuleSetSourceOptions struct {
	AcceptExtendedRules bool `json:"accept_extended_rules,omitempty"`
}

type ClashRuleProviderSourceOptions struct {
	SourceFormat   string `json:"source_format,omitempty"`
	SourceBehavior string `json:"source_behavior,omitempty"`
}

type ClashRuleProviderTargetOptions struct {
	TargetFormat   string `json:"target_format,omitempty"`
	TargetBehavior string `json:"target_behavior,omitempty"`
}

type SurgeRuleProviderSourceOptions struct {
	SourceBehavior string `json:"source_behavior,omitempty"`
}

type SurgeRuleProviderTargetOptions struct {
	TargetBehavior string `json:"target_behavior,omitempty"`
}
