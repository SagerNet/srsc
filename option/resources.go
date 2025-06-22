package option

import (
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badjson"
)

type ResourceOptions struct {
	GEOIP *Resource `json:"geoip,omitempty"`
	IPASN *Resource `json:"ipasn,omitempty"`
}

type Resource struct {
	SourceOptions
	SourceConvertOptions
}

func (e Resource) MarshalJSON() ([]byte, error) {
	return badjson.MarshallObjects(e.SourceOptions, e.SourceConvertOptions)
}

func (e *Resource) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &e.SourceOptions)
	if err != nil {
		return err
	}
	return badjson.UnmarshallExcludedMulti(bytes, &e.SourceOptions, &e.SourceConvertOptions)
}
