package convertor

import (
	"bufio"
	"bytes"
	"context"
	"strings"

	boxConstant "github.com/sagernet/sing-box/constant"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/convertor/clash"
)

var _ adapter.Convertor = (*SurgeRuleSet)(nil)

type SurgeRuleSet struct{}

func (s *SurgeRuleSet) Type() string {
	return C.ConvertorTypeSurgeRuleSet
}

func (s *SurgeRuleSet) ContentType(options adapter.ConvertOptions) string {
	return "text/plain"
}

func (s *SurgeRuleSet) From(ctx context.Context, content []byte, options adapter.ConvertOptions) ([]adapter.Rule, error) {
	behavior := options.Options.SourceConvertOptions.SurgeOptions.SourceBehavior
	switch behavior {
	case "", "classical":
		var rules []adapter.Rule
		scanner := bufio.NewScanner(bytes.NewReader(content))
		for scanner.Scan() {
			rule, _ := clash.FromSurgeLine(scanner.Text())
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
		return adapter.MergeRules(rules), nil
	case "domain":
		var rule adapter.DefaultRule
		scanner := bufio.NewScanner(bytes.NewReader(content))
		for scanner.Scan() {
			ruleLine := strings.TrimSpace(scanner.Text())
			if ruleLine == "" || strings.HasPrefix(ruleLine, "#") {
				continue
			}
			if strings.HasPrefix(ruleLine, ".") {
				rule.DomainSuffix = append(rule.DomainSuffix, strings.TrimPrefix(ruleLine, "."))
			} else {
				rule.Domain = append(rule.Domain, ruleLine)
			}
		}
		return []adapter.Rule{{Type: boxConstant.RuleTypeDefault, DefaultOptions: rule}}, nil
	default:
		return nil, E.New("unknown Surge source behavior: " + behavior)
	}
}

func (s *SurgeRuleSet) To(ctx context.Context, contentRules []adapter.Rule, options adapter.ConvertOptions) ([]byte, error) {
	behavior := options.Options.TargetConvertOptions.SurgeOptions.TargetBehavior
	switch behavior {
	case "", "classical":
		convertedRules, err := adapter.EmbedResourceRules(ctx, contentRules)
		if err != nil {
			return nil, err
		}
		var lines []string
		for _, rule := range convertedRules {
			ruleLines, err := clash.ToSurgeLines(rule)
			if err != nil {
				continue
			}
			lines = append(lines, ruleLines...)
		}
		return []byte(strings.Join(lines, "\n")), nil
	case "domain":
		var output bytes.Buffer
		for _, rule := range contentRules {
			if rule.Type != boxConstant.RuleTypeDefault || !adapter.IsDestinationAddressRule(rule.DefaultOptions) {
				continue
			}
			for _, domain := range rule.DefaultOptions.Domain {
				output.WriteString(domain + "\n")
			}
			for _, domainSuffix := range rule.DefaultOptions.DomainSuffix {
				output.WriteString("." + domainSuffix + "\n")
			}
		}
		return output.Bytes(), nil
	default:
		return nil, E.New("unknown Surge target behavior: " + behavior)
	}
}
