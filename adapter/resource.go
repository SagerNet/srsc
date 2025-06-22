package adapter

import (
	"context"

	boxConstant "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/service"
)

type ResourceManager interface {
	GEOIPConfigured() bool
	GEOIP(country string) (*option.DefaultHeadlessRule, error)
	IPASNConfigured() bool
	IPASN(asn string) (*option.DefaultHeadlessRule, error)
}

func EmbedResourceRules(ctx context.Context, rules []Rule) ([]Rule, error) {
	resourceManager := service.FromContext[ResourceManager](ctx)
	for index, rule := range rules {
		err := embedResourceRule(ctx, resourceManager, &rule)
		if err != nil {
			return nil, err
		}
		rules[index] = rule
	}
	return rules, nil
}

func embedResourceRule(ctx context.Context, resourceManager ResourceManager, rule *Rule) error {
	if rule.Type == boxConstant.RuleTypeLogical {
		for index, subRule := range rule.LogicalOptions.Rules {
			err := embedResourceRule(ctx, resourceManager, &subRule)
			if err != nil {
				return err
			}
			rule.LogicalOptions.Rules[index] = subRule
		}
		return nil
	}
	if resourceManager.GEOIPConfigured() {
		for _, geoip := range rule.DefaultOptions.GEOIP {
			geoipRule, err := resourceManager.GEOIP(geoip)
			if err != nil {
				return E.Cause(err, "fetch GEOIP resource: ", rule.DefaultOptions.GEOIP)
			}
			/*newHeadlessRule, err := badjson.Merge(ctx, rule.DefaultOptions.DefaultHeadlessRule, *geoipRule, false)
			if err != nil {
				return E.Cause(err, "merge GEOIP resource")
			}*/
			if len(geoipRule.IPCIDR) > 0 {
				rule.DefaultOptions.IPCIDR = append(rule.DefaultOptions.IPCIDR, geoipRule.IPCIDR...)
			} else if len(geoipRule.SourceIPCIDR) > 0 {
				rule.DefaultOptions.IPCIDR = append(rule.DefaultOptions.IPCIDR, geoipRule.SourceIPCIDR...)
			}
		}
		rule.DefaultOptions.GEOIP = nil
		for _, sourceGeoip := range rule.DefaultOptions.SourceGEOIP {
			sourceGeoipRule, err := resourceManager.GEOIP(sourceGeoip)
			if err != nil {
				return E.Cause(err, "fetch GEOIP resource: ", rule.DefaultOptions.SourceGEOIP)
			}
			if len(sourceGeoipRule.IPCIDR) > 0 {
				rule.DefaultOptions.SourceIPCIDR = append(rule.DefaultOptions.SourceIPCIDR, sourceGeoipRule.IPCIDR...)
			} else if len(sourceGeoipRule.SourceIPCIDR) > 0 {
				rule.DefaultOptions.SourceIPCIDR = append(rule.DefaultOptions.SourceIPCIDR, sourceGeoipRule.SourceIPCIDR...)
			}
		}
		rule.DefaultOptions.SourceGEOIP = nil
	}
	if resourceManager.IPASNConfigured() {
		for _, ipasn := range rule.DefaultOptions.IPASN {
			ipasnRule, err := resourceManager.IPASN(ipasn)
			if err != nil {
				return E.Cause(err, "fetch IPASN resource: ", rule.DefaultOptions.IPASN)
			}
			if len(ipasnRule.IPCIDR) > 0 {
				rule.DefaultOptions.IPCIDR = append(rule.DefaultOptions.IPCIDR, ipasnRule.IPCIDR...)
			} else if len(ipasnRule.SourceIPCIDR) > 0 {
				rule.DefaultOptions.IPCIDR = append(rule.DefaultOptions.IPCIDR, ipasnRule.SourceIPCIDR...)
			}
		}
		rule.DefaultOptions.IPASN = nil
		for _, sourceIPASN := range rule.DefaultOptions.SourceIPASN {
			sourceIPASNRule, err := resourceManager.IPASN(sourceIPASN)
			if err != nil {
				return E.Cause(err, "fetch IPASN resource: ", rule.DefaultOptions.SourceIPASN)
			}
			if len(sourceIPASNRule.IPCIDR) > 0 {
				rule.DefaultOptions.SourceIPCIDR = append(rule.DefaultOptions.SourceIPCIDR, sourceIPASNRule.IPCIDR...)
			} else if len(sourceIPASNRule.SourceIPCIDR) > 0 {
				rule.DefaultOptions.SourceIPCIDR = append(rule.DefaultOptions.SourceIPCIDR, sourceIPASNRule.SourceIPCIDR...)
			}
		}
		rule.DefaultOptions.SourceIPASN = nil
	}
	return nil
}
