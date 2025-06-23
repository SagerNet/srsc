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
