package resource

import (
	"context"
	"os"

	boxConstant "github.com/sagernet/sing-box/constant"
	boxOption "github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/convertor"
	"github.com/sagernet/srsc/option"
	"github.com/sagernet/srsc/source"
)

var _ adapter.ResourceManager = (*Manager)(nil)

type Manager struct {
	ctx     context.Context
	logger  logger.ContextLogger
	cache   adapter.Cache
	geoip   *Resource
	geosite *Resource
	ipasn   *Resource
}

type Resource struct {
	adapter.Source
	adapter.Convertor
	option.SourceConvertOptions
}

func NewResource(ctx context.Context, options *option.Resource) (*Resource, error) {
	resSource, err := source.New(ctx, options.SourceOptions)
	if err != nil {
		return nil, err
	}
	resConvertor, loaded := convertor.Convertors[options.SourceType]
	if !loaded {
		return nil, E.New("unknown source type: ", options.SourceType)
	}
	return &Resource{
		Source:               resSource,
		Convertor:            resConvertor,
		SourceConvertOptions: options.SourceConvertOptions,
	}, nil
}

func NewManager(ctx context.Context, logger logger.ContextLogger, options option.ResourceOptions) (*Manager, error) {
	m := &Manager{
		ctx:    ctx,
		logger: logger,
		cache:  service.FromContext[adapter.Cache](ctx),
	}
	if options.GEOIP != nil {
		geoip, err := NewResource(ctx, options.GEOIP)
		if err != nil {
			return nil, E.Cause(err, "create resource for GEOIP")
		}
		m.geoip = geoip
	}
	if options.GEOSite != nil {
		geosite, err := NewResource(ctx, options.GEOSite)
		if err != nil {
			return nil, E.Cause(err, "create resource for GEOSite")
		}
		m.geosite = geosite
	}
	if options.IPASN != nil {
		ipasn, err := NewResource(ctx, options.IPASN)
		if err != nil {
			return nil, E.Cause(err, "create resource for IPASN")
		}
		m.ipasn = ipasn
	}
	return m, nil
}

func (m *Manager) GEOIPConfigured() bool {
	return m.geoip != nil
}

func (m *Manager) GEOIP(code string) (*boxOption.DefaultHeadlessRule, error) {
	if m.geoip == nil {
		return nil, E.New("GEOIP resource source is not configured")
	}
	cachePath, err := m.geoip.Path(map[string]string{
		"code": code,
	})
	if err != nil {
		return nil, E.Cause(err, "evaluate source path")
	}
	return m.fetch(m.geoip, cachePath, "res.geoip."+cachePath)
}

func (m *Manager) GEOSiteConfigured() bool {
	return m.geosite != nil
}

func (m *Manager) GEOSite(code string) (*boxOption.DefaultHeadlessRule, error) {
	if m.geosite == nil {
		return nil, E.New("GEOSite resource source is not configured")
	}
	cachePath, err := m.geosite.Path(map[string]string{
		"code": code,
	})
	if err != nil {
		return nil, E.Cause(err, "evaluate source path")
	}
	return m.fetch(m.geosite, cachePath, "res.geosite."+cachePath)
}

func (m *Manager) IPASNConfigured() bool {
	return m.ipasn != nil
}

func (m *Manager) IPASN(asn string) (*boxOption.DefaultHeadlessRule, error) {
	if m.ipasn == nil {
		return nil, E.New("IPASN resource source is not configured")
	}
	cachePath, err := m.ipasn.Path(map[string]string{
		"asn": asn,
	})
	if err != nil {
		return nil, E.Cause(err, "evaluate source path")
	}
	return m.fetch(m.ipasn, cachePath, "res.ipasn."+cachePath)
}

func (m *Manager) fetch(r *Resource, cachePath string, cacheKey string) (*boxOption.DefaultHeadlessRule, error) {
	cachedBinary, err := m.cache.LoadBinary(cacheKey)
	if err != nil && !os.IsNotExist(err) {
		return nil, E.Cause(err, "load cache binary")
	}
	lastUpdated := r.LastUpdated(cachePath)
	if cachedBinary != nil && !lastUpdated.IsZero() && cachedBinary.LastUpdated.Equal(lastUpdated) {
		return m.loadCache(cachedBinary)
	}
	var fetchBody adapter.FetchRequestBody
	if cachedBinary != nil {
		fetchBody.ETag = cachedBinary.LastEtag
		fetchBody.LastUpdated = cachedBinary.LastUpdated
	}
	response, err := r.Fetch(cachePath, fetchBody)
	if err != nil {
		return nil, E.Cause(err, "fetch source")
	}
	if response.NotModified {
		if cachedBinary == nil {
			return nil, E.New("fetch source: unexpected not modified response")
		}
		if response.LastUpdated != cachedBinary.LastUpdated {
			cachedBinary.LastUpdated = response.LastUpdated
			err = m.cache.SaveBinary(cacheKey, cachedBinary)
			if err != nil {
				return nil, E.Cause(err, "save cache binary")
			}
		}
		return m.loadCache(cachedBinary)
	}
	if len(response.Content) == 0 {
		return nil, E.Cause(err, "fetch source: empty content")
	}
	var rules []adapter.Rule
	rules, err = r.From(m.ctx, response.Content, adapter.ConvertOptions{
		Options: option.ConvertOptions{
			SourceConvertOptions: r.SourceConvertOptions,
		},
	})
	if err != nil {
		return nil, E.Cause(err, "decode source")
	}
	if len(rules) != 1 {
		return nil, E.New("unexpected resource rule count: ", len(rules))
	} else if rules[0].Type != boxConstant.RuleTypeDefault {
		return nil, E.New("unexpected complex resource: logical rule")
	} else if !rules[0].Headlessable() {
		return nil, E.New("unexpected complex resource: unsupported by sing-box")
	}
	binary, err := convertor.Convertors[C.ConvertorTypeRuleSetBinary].To(m.ctx, rules, adapter.ConvertOptions{
		Options: option.ConvertOptions{TargetConvertOptions: option.TargetConvertOptions{TargetType: C.ConvertorTypeRuleSetBinary}},
	})
	if err != nil {
		return nil, E.Cause(err, "encode JSON")
	}
	cachedBinary = &adapter.SavedBinary{
		Content:     binary,
		LastUpdated: response.LastUpdated,
		LastEtag:    response.ETag,
	}
	err = m.cache.SaveBinary(cacheKey, cachedBinary)
	if err != nil {
		return nil, E.Cause(err, "save cache binary")
	}
	return m.loadCache(cachedBinary)
}

func (m *Manager) loadCache(cachedBinary *adapter.SavedBinary) (*boxOption.DefaultHeadlessRule, error) {
	rules, err := convertor.Convertors[C.ConvertorTypeRuleSetBinary].From(m.ctx, cachedBinary.Content, adapter.ConvertOptions{
		Options: option.ConvertOptions{SourceConvertOptions: option.SourceConvertOptions{SourceType: C.ConvertorTypeRuleSetBinary}},
	})
	if err != nil {
		return nil, err
	}
	if len(rules) != 1 || rules[0].Type != boxConstant.RuleTypeDefault || !rules[0].Headlessable() {
		return nil, E.New("unexpected complex resource")
	}
	return &rules[0].DefaultOptions.DefaultHeadlessRule, nil
}
