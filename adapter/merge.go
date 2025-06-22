package adapter

import (
	"reflect"

	C "github.com/sagernet/sing-box/constant"
)

func MergeRules(rules []Rule) []Rule {
	return mergeDestinationAddressRules(rules)
}

func mergeDestinationAddressRules(rules []Rule) []Rule {
	var outputRules []Rule
	var destinationRule *DefaultRule
	for _, rule := range rules {
		if rule.Type == C.RuleTypeDefault && IsDestinationAddressRule(rule.DefaultOptions) {
			if destinationRule == nil {
				destinationRule = &rule.DefaultOptions
			} else {
				destinationRule.Domain = append(destinationRule.Domain, rule.DefaultOptions.Domain...)
				destinationRule.DomainSuffix = append(destinationRule.DomainSuffix, rule.DefaultOptions.DomainSuffix...)
				destinationRule.DomainKeyword = append(destinationRule.DomainKeyword, rule.DefaultOptions.DomainKeyword...)
				destinationRule.DomainRegex = append(destinationRule.DomainRegex, rule.DefaultOptions.DomainRegex...)
				destinationRule.IPCIDR = append(destinationRule.IPCIDR, rule.DefaultOptions.IPCIDR...)
				destinationRule.GEOIP = append(destinationRule.GEOIP, rule.DefaultOptions.GEOIP...)
				destinationRule.IPASN = append(destinationRule.IPASN, rule.DefaultOptions.IPASN...)
			}
		} else {
			outputRules = append(outputRules, rule)
		}
	}
	if destinationRule != nil {
		outputRules = append([]Rule{{Type: C.RuleTypeDefault, DefaultOptions: *destinationRule}}, outputRules...)
	}
	return outputRules
}

func IsDestinationAddressRule(rule DefaultRule) bool {
	var defaultRule DefaultRule
	defaultRule.Domain = rule.Domain
	defaultRule.DomainSuffix = rule.DomainSuffix
	defaultRule.DomainKeyword = rule.DomainKeyword
	defaultRule.DomainRegex = rule.DomainRegex
	defaultRule.IPCIDR = rule.IPCIDR
	defaultRule.GEOIP = rule.GEOIP
	defaultRule.IPASN = rule.IPASN
	return reflect.DeepEqual(rule, defaultRule)
}
