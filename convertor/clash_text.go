package convertor

import (
	"bufio"
	"bytes"
	"context"
	"net/netip"
	"strings"

	boxConstant "github.com/sagernet/sing-box/constant"
	boxOption "github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
)

var _ adapter.Convertor = (*ClashTextRuleProvider)(nil)

type ClashTextRuleProvider struct{}

func (s *ClashTextRuleProvider) Type() string {
	return C.ConvertorTypeClashTextRuleProvider
}

func (s *ClashTextRuleProvider) ContentType(options adapter.ConvertOptions) string {
	return "plain/text"
}

func (s *ClashTextRuleProvider) From(ctx context.Context, binary []byte, options adapter.ConvertOptions) (*boxOption.PlainRuleSetCompat, error) {
	scanner := bufio.NewScanner(bytes.NewReader(binary))
	if !scanner.Scan() {
		return nil, E.New("empty rule set")
	}
	var (
		rule      boxOption.DefaultHeadlessRule
		ipPayload bool
	)
	for i := 0; scanner.Scan(); i++ {
		ruleLine := scanner.Text()
		if ruleLine == "" || ruleLine[0] == '#' {
			continue
		}
		if i == 0 {
			if _, err := netip.ParsePrefix(ruleLine); err == nil {
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

func (s *ClashTextRuleProvider) To(ctx context.Context, source *boxOption.PlainRuleSetCompat, options adapter.ConvertOptions) ([]byte, error) {
	var output bytes.Buffer
	for _, rule := range source.Options.Rules[0].DefaultOptions.IPCIDR {
		output.WriteString(rule + "\n")
	}
	for _, rule := range source.Options.Rules[0].DefaultOptions.Domain {
		output.WriteString(rule + "\n")
	}
	for _, rule := range source.Options.Rules[0].DefaultOptions.DomainSuffix {
		output.WriteString("+." + rule + "\n")
	}
	return output.Bytes(), nil
}
