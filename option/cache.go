package option

import (
	"time"

	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badjson"
	"github.com/sagernet/sing/common/json/badoption"
	C "github.com/sagernet/srsc/constant"
)

type _CacheOptions struct {
	Type         string            `json:"type,omitempty"`
	RedisOptions RedisCacheOptions `json:"-"`
	Timeout      time.Duration     `json:"timeout,omitempty"`
}

type CacheOptions _CacheOptions

func (o CacheOptions) MarshalJSON() ([]byte, error) {
	var v any
	switch o.Type {
	case C.CacheTypeRedis:
		v = o.RedisOptions
	case "":
		return nil, E.New("missing cache type")
	default:
		return nil, E.New("unknown cache type: " + o.Type)
	}
	return badjson.MarshallObjects((_CacheOptions)(o), v)
}

func (o *CacheOptions) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, (*_CacheOptions)(o))
	if err != nil {
		return err
	}
	var v any
	switch o.Type {
	case C.CacheTypeRedis:
		v = &o.RedisOptions
	default:
		return E.New("unknown cache type: " + o.Type)
	}
	return badjson.UnmarshallExcluded(bytes, (*_CacheOptions)(o), v)
}

type RedisCacheOptions struct {
	Address  badoption.Listable[string] `json:"address,omitempty"`
	Username string                     `json:"username,omitempty"`
	Password string                     `json:"password,omitempty"`
	DB       int                        `json:"db,omitempty"`
	Protocol int                        `json:"protocol,omitempty"`
	PoolSize int                        `json:"pool_size,omitempty"`
	option.OutboundTLSOptionsContainer
}
