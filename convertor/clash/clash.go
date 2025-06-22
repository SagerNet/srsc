package clash

import (
	"bufio"
	"bytes"
	"context"
	"reflect"
	"strings"

	boxConstant "github.com/sagernet/sing-box/constant"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/convertor/adguard"

	"gopkg.in/yaml.v3"
)

var _ adapter.Convertor = (*RuleProvider)(nil)

type RuleProvider struct{}

func (c *RuleProvider) Type() string {
	return C.ConvertorTypeClashRuleProvider
}

func (c *RuleProvider) ContentType(options adapter.ConvertOptions) string {
	switch options.Options.TargetConvertOptions.ClashOptions.TargetFormat {
	case "yaml":
		return "application/x-yaml"
	case "mrs":
		return "application/octet-stream"
	default:
		return "text/plain"
	}
}

func (c *RuleProvider) From(ctx context.Context, content []byte, options adapter.ConvertOptions) ([]adapter.Rule, error) {
	format := options.Options.SourceConvertOptions.ClashOptions.SourceFormat
	var lines []string
	switch format {
	case "text":
	case "yaml":
		var ruleProvider struct {
			Payload []string `yaml:"payload"`
		}
		err := yaml.Unmarshal(content, &ruleProvider)
		if err != nil {
			return nil, err
		}
		lines = ruleProvider.Payload
	case "mrs":
		return fromMrs(content)
	case "":
		return nil, E.New("missing source format in options")
	default:
		return nil, E.New("unknown source format: ", format)
	}
	behavior := options.Options.SourceConvertOptions.ClashOptions.SourceBehavior
	switch behavior {
	case "domain":
		var rule adapter.DefaultRule
		if len(lines) > 0 {
			for _, line := range lines {
				fromDomainLine(&rule, line)
			}
		} else {
			scanner := bufio.NewScanner(bytes.NewReader(content))
			for scanner.Scan() {
				fromDomainLine(&rule, scanner.Text())
			}
		}
		return []adapter.Rule{{Type: boxConstant.RuleTypeDefault, DefaultOptions: rule}}, nil
	case "ipcidr":
		var rule adapter.DefaultRule
		if len(lines) > 0 {
			for _, line := range lines {
				fromIPCIDRLine(&rule, line)
			}
		} else {
			scanner := bufio.NewScanner(bytes.NewReader(content))
			for scanner.Scan() {
				fromIPCIDRLine(&rule, scanner.Text())
			}
		}
		return []adapter.Rule{{Type: boxConstant.RuleTypeDefault, DefaultOptions: rule}}, nil
	case "classical":
		var rules []adapter.Rule
		if len(lines) > 0 {
			for _, line := range lines {
				rule, _ := fromClassicalLine(line)
				if rule != nil {
					rules = append(rules, *rule)
				}
			}
		} else {
			scanner := bufio.NewScanner(bytes.NewReader(content))
			for scanner.Scan() {
				rule, _ := fromClassicalLine(scanner.Text())
				if rule != nil {
					rules = append(rules, *rule)
				}
			}
		}
		return rules, nil
	case "":
		return nil, E.New("missing source behavior in options")
	default:
		return nil, E.New("unknown source behavior: ", behavior)
	}
}

func (c *RuleProvider) To(ctx context.Context, contentRules []adapter.Rule, options adapter.ConvertOptions) ([]byte, error) {
	format := options.Options.TargetConvertOptions.ClashOptions.TargetFormat
	behavior := options.Options.TargetConvertOptions.ClashOptions.TargetBehavior
	if format == "mrs" {
		return toMrs(behavior, contentRules)
	}
	ruleLines, err := toLines(behavior, contentRules)
	if err != nil {
		return nil, err
	}
	switch format {
	case "text":
		var output bytes.Buffer
		for _, line := range ruleLines {
			output.WriteString(line + "\n")
		}
		return output.Bytes(), nil
	case "yaml":
		var output bytes.Buffer
		ruleProvider := struct {
			Payload []string `yaml:"payload"`
		}{
			Payload: ruleLines,
		}
		encoder := yaml.NewEncoder(&output)
		encoder.SetIndent(2)
		err = encoder.Encode(ruleProvider)
		if err != nil {
			return nil, err
		}
		return output.Bytes(), nil
	case "":
		return nil, E.New("missing target format in options")
	default:
		return nil, E.New("unknown target format: ", format)
	}
}

func fromDomainLine(rule *adapter.DefaultRule, ruleLine string) {
	if ruleLine == "" || strings.HasPrefix(ruleLine, "#") {
		return
	}
	var domainSuffix bool
	if strings.HasPrefix(ruleLine, "+.") {
		domainSuffix = true
		ruleLine = strings.TrimPrefix(ruleLine, "+.")
	}
	if strings.Contains(ruleLine, "+") || strings.Contains(ruleLine, "*") {
		return
	}
	if domainSuffix {
		rule.DomainSuffix = append(rule.DomainSuffix, ruleLine)
	} else {
		rule.Domain = append(rule.Domain, ruleLine)
	}
}

func fromIPCIDRLine(rule *adapter.DefaultRule, ruleLine string) {
	if ruleLine == "" || strings.HasPrefix(ruleLine, "#") {
		return
	}
	rule.IPCIDR = append(rule.IPCIDR, ruleLine)
}

func toLines(behavior string, rules []adapter.Rule) ([]string, error) {
	var lines []string
	switch behavior {
	case "domain":
		for _, rule := range rules {
			if rule.Type != boxConstant.RuleTypeDefault || !adguard.IsDestinationAddressRule(rule.DefaultOptions) {
				continue
			}
			for _, domain := range rule.DefaultOptions.Domain {
				lines = append(lines, domain)
			}
			for _, domainSuffix := range rule.DefaultOptions.DomainSuffix {
				lines = append(lines, domainSuffix)
			}
		}
		return lines, nil
	case "ipcidr":
		for _, rule := range rules {
			if rule.Type != boxConstant.RuleTypeDefault || !adguard.IsDestinationAddressRule(rule.DefaultOptions) {
				continue
			}
			for _, ipCidr := range rule.DefaultOptions.IPCIDR {
				lines = append(lines, ipCidr)
			}
		}
	case "classical":
		for _, rule := range rules {
			ruleLines, err := toClassicalLine(rule)
			if err == nil {
				continue
			}
			lines = append(lines, ruleLines...)
		}
	}
	return lines, nil
}

func IsSimpleDomainRule(rule adapter.DefaultRule) bool {
	var defaultRule adapter.DefaultRule
	defaultRule.Domain = rule.Domain
	defaultRule.DomainSuffix = rule.DomainSuffix
	return reflect.DeepEqual(rule, defaultRule)
}
