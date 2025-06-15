package convertor

import (
	"bytes"
	"context"
	"net/netip"
	"strings"

	boxConstant "github.com/sagernet/sing-box/constant"
	boxOption "github.com/sagernet/sing-box/option"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"

	"github.com/stretchr/testify/assert/yaml"
)

var _ adapter.Convertor = (*ClashYamlRuleProvider)(nil)

type ClashYamlRuleProvider struct{}

func (s *ClashYamlRuleProvider) Type() string {
	return C.ConvertorTypeClashYamlRuleProvider
}

func (s *ClashYamlRuleProvider) ContentType(options adapter.ConvertOptions) string {
	return "application/x-yaml"
}

func (s *ClashYamlRuleProvider) From(ctx context.Context, binary []byte, options adapter.ConvertOptions) (*boxOption.PlainRuleSetCompat, error) {
	var ruleProvider struct {
		Payload []string `yaml:"payload"`
	}
	err := yaml.Unmarshal(binary, &ruleProvider)
	if err != nil {
		return nil, err
	}
	var (
		rule      boxOption.DefaultHeadlessRule
		ipPayload bool
	)
	for i, ruleLine := range ruleProvider.Payload {
		if i == 0 {
			if _, err = netip.ParsePrefix(ruleLine); err == nil {
				ipPayload = true
			}
		}
		if ipPayload {
			rule.IPCIDR = append(rule.IPCIDR, ruleLine)
		} else {
			var domainSuffix bool
			if strings.HasPrefix(ruleLine, "+.") {
				domainSuffix = true
				ruleLine = strings.TrimPrefix(ruleLine, "+.")
			}
			if strings.Contains(ruleLine, "+") || strings.Contains(ruleLine, "*") {
				continue
			}
			if domainSuffix {
				rule.DomainSuffix = append(rule.DomainSuffix, ruleLine)
			} else {
				rule.Domain = append(rule.Domain, ruleLine)
			}
		}
	}
	return &boxOption.PlainRuleSetCompat{
		Version: boxConstant.RuleSetVersionCurrent,
		Options: boxOption.PlainRuleSet{
			Rules: []boxOption.HeadlessRule{{
				Type:           boxConstant.RuleTypeDefault,
				DefaultOptions: rule,
			}},
		},
	}, nil
}

func (s *ClashYamlRuleProvider) To(ctx context.Context, source *boxOption.PlainRuleSetCompat, options adapter.ConvertOptions) ([]byte, error) {
	var output bytes.Buffer
	output.WriteString("payload:\n")
	for _, rule := range source.Options.Rules[0].DefaultOptions.IPCIDR {
		output.WriteString("  - " + rule + "\n")
	}
	for _, rule := range source.Options.Rules[0].DefaultOptions.Domain {
		output.WriteString("  - " + rule + "\n")
	}
	for _, rule := range source.Options.Rules[0].DefaultOptions.DomainSuffix {
		output.WriteString("  - +." + rule + "\n")
	}
	return output.Bytes(), nil
}
