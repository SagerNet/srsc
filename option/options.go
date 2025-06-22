package option

import (
	"bytes"
	"context"

	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badjson"
	"github.com/sagernet/sing/common/json/badoption"
	C "github.com/sagernet/srsc/constant"
)

type _Options struct {
	Log        *option.LogOptions                   `json:"log,omitempty"`
	Listen     *badoption.Addr                      `json:"listen,omitempty"`
	ListenPort uint16                               `json:"listen_port,omitempty"`
	Endpoints  *badjson.TypedMap[string, *Endpoint] `json:"endpoints,omitempty"`
	Resources  *ResourceOptions                     `json:"resources,omitempty"`
	option.InboundTLSOptionsContainer
	Cache      *CacheOptions `json:"cache,omitempty"`
	RawMessage []byte        `json:"-"`
}

type Options _Options

func (o *Options) UnmarshalJSONContext(ctx context.Context, content []byte) error {
	decoder := json.NewDecoderContext(ctx, bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	err := decoder.Decode((*_Options)(o))
	if err != nil {
		return err
	}
	o.RawMessage = content
	return nil
}

type _Endpoint struct {
	Type        string       `json:"type,omitempty"`
	FileOptions FileEndpoint `json:"-"`
}

type Endpoint _Endpoint

func (o Endpoint) MarshalJSON() ([]byte, error) {
	var v any
	switch o.Type {
	case C.EndpointTypeFile:
		v = o.FileOptions
	case "":
		return nil, E.New("missing endpoint type")
	default:
		return nil, E.New("unknown endpoint type: " + o.Type)
	}
	return badjson.MarshallObjects((_Endpoint)(o), v)
}

func (o *Endpoint) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, (*_Endpoint)(o))
	if err != nil {
		return err
	}
	var v any
	switch o.Type {
	case C.EndpointTypeFile:
		v = &o.FileOptions
	default:
		return E.New("unknown endpoint type: " + o.Type)
	}
	return badjson.UnmarshallExcluded(bytes, (*_Endpoint)(o), v)
}
