package option

import (
	_ "unsafe"

	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badjson"
	"github.com/sagernet/sing/common/json/badoption"
	C "github.com/sagernet/srsc/constant"
)

type _FileEndpoint struct {
	SourceOptions
	ConvertOptions
}

type FileEndpoint _FileEndpoint

func (e FileEndpoint) MarshalJSON() ([]byte, error) {
	return badjson.MarshallObjects(e.SourceOptions, e.ConvertOptions)
}

func (e *FileEndpoint) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &e.SourceOptions)
	if err != nil {
		return err
	}
	return badjson.UnmarshallExcludedMulti(bytes, &e.SourceOptions, &e.ConvertOptions)
}

type _SourceOptions struct {
	Source        string       `json:"source,omitempty"`
	LocalOptions  LocalSource  `json:"-"`
	RemoteOptions RemoteSource `json:"-"`
}

type SourceOptions _SourceOptions

func (o SourceOptions) MarshalJSON() ([]byte, error) {
	var v any
	switch o.Source {
	case C.EndpointSourceLocal:
		v = o.LocalOptions
	case C.EndpointSourceRemote:
		v = o.RemoteOptions
	case "":
		return nil, E.New("missing endpoint source")
	default:
		return nil, E.New("unknown endpoint source: " + o.Source)
	}
	return badjson.MarshallObjects((_SourceOptions)(o), v)
}

func (o *SourceOptions) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, (*_SourceOptions)(o))
	if err != nil {
		return err
	}
	var v any
	switch o.Source {
	case C.EndpointSourceLocal:
		v = &o.LocalOptions
	case C.EndpointSourceRemote:
		v = &o.RemoteOptions
	case "":
		return E.New("missing endpoint source")
	default:
		return E.New("unknown endpoint source: " + o.Source)
	}
	return json.Unmarshal(bytes, v)
}

type LocalSource struct {
	Path string `json:"path,omitempty"`
}

type RemoteSource struct {
	URL       string             `json:"url,omitempty"`
	UserAgent string             `json:"user_agent,omitempty"`
	TTL       badoption.Duration `json:"ttl,omitempty"`
	option.OutboundTLSOptionsContainer
	option.DialerOptions
}

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
}

type SourceConvertOptions _SourceConvertOptions

func (o SourceConvertOptions) MarshalJSON() ([]byte, error) {
	var v any
	switch o.SourceType {
	case C.ConvertorTypeAdGuardRuleSet:
		v = o.AdGuardOptions
	case C.ConvertorTypeClashRuleProvider:
		v = o.ClashOptions
	case C.ConvertorTypeRuleSetSource, C.ConvertorTypeRuleSetBinary, C.ConvertorTypeSurgeRuleSet, C.ConvertorTypeSurgeDomainSet:
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
	case C.ConvertorTypeAdGuardRuleSet:
		v = &o.AdGuardOptions
	case C.ConvertorTypeClashRuleProvider:
		v = &o.ClashOptions
	case C.ConvertorTypeRuleSetSource, C.ConvertorTypeRuleSetBinary, C.ConvertorTypeSurgeRuleSet, C.ConvertorTypeSurgeDomainSet:
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
}

type TargetConvertOptions _TargetConvertOptions

func (o TargetConvertOptions) MarshalJSON() ([]byte, error) {
	var v any
	switch o.TargetType {
	case C.ConvertorTypeClashRuleProvider:
		v = o.ClashOptions
	case C.ConvertorTypeRuleSetSource, C.ConvertorTypeRuleSetBinary, C.ConvertorTypeAdGuardRuleSet:
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
	case C.ConvertorTypeClashRuleProvider:
		v = &o.ClashOptions
	case C.ConvertorTypeRuleSetSource, C.ConvertorTypeRuleSetBinary, C.ConvertorTypeAdGuardRuleSet:
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
