package convertor

import (
	"bufio"
	"bytes"
	"context"
	"strings"

	boxConstant "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/convertor/clash"
)

var (
	_ adapter.Convertor = (*SurgeRuleSet)(nil)
	_ adapter.Convertor = (*SurgeDomainSet)(nil)
)

type SurgeRuleSet struct{}

func (s *SurgeRuleSet) Type() string {
	return C.ConvertorTypeSurgeRuleSet
}

func (s *SurgeRuleSet) ContentType(options adapter.ConvertOptions) string {
	return "text/plain"
}

func (s *SurgeRuleSet) From(ctx context.Context, content []byte, options adapter.ConvertOptions) (*option.PlainRuleSetCompat, error) {
	var rules []option.HeadlessRule
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		rule, _ := clash.FromSurgeLine(scanner.Text())
		if rule != nil {
			rules = append(rules, *rule)
		}
	}
	return &option.PlainRuleSetCompat{
		Version: boxConstant.RuleSetVersionCurrent,
		Options: option.PlainRuleSet{Rules: rules},
	}, nil
}

func (s *SurgeRuleSet) To(ctx context.Context, source *option.PlainRuleSetCompat, options adapter.ConvertOptions) ([]byte, error) {
	var lines []string
	for _, rule := range source.Options.Rules {
		lines = append(lines, clash.ToSurgeLines(&rule)...)
	}
	return []byte(strings.Join(lines, "\n")), nil
}

type SurgeDomainSet struct{}

func (s *SurgeDomainSet) Type() string {
	return C.ConvertorTypeSurgeDomainSet
}

func (s *SurgeDomainSet) ContentType(options adapter.ConvertOptions) string {
	return "text/plain"
}

func (s *SurgeDomainSet) From(ctx context.Context, binary []byte, options adapter.ConvertOptions) (*option.PlainRuleSetCompat, error) {
	var rule option.DefaultHeadlessRule
	scanner := bufio.NewScanner(bytes.NewReader(binary))
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
	return &option.PlainRuleSetCompat{
		Version: boxConstant.RuleSetVersionCurrent,
		Options: option.PlainRuleSet{
			Rules: []option.HeadlessRule{{
				Type:           boxConstant.RuleTypeDefault,
				DefaultOptions: rule,
			}},
		},
	}, nil
}

func (s *SurgeDomainSet) To(ctx context.Context, source *option.PlainRuleSetCompat, options adapter.ConvertOptions) ([]byte, error) {
	var output bytes.Buffer
	for _, rule := range source.Options.Rules {
		if rule.Type == boxConstant.RuleTypeDefault {
			for _, domain := range rule.DefaultOptions.Domain {
				output.WriteString(domain + "\n")
			}
			for _, domainSuffix := range rule.DefaultOptions.DomainSuffix {
				output.WriteString("." + domainSuffix + "\n")
			}
		}
	}
	return output.Bytes(), nil
}
