package convertor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"net/netip"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	boxOption "github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/srsc/adapter"
	"github.com/sagernet/srsc/convertor/internal/meta_cidr"
	"github.com/sagernet/srsc/convertor/internal/meta_domainset"

	"github.com/klauspost/compress/zstd"
)

var MrsMagicBytes = [4]byte{'M', 'R', 'S', 1} // MRSv1

var _ adapter.Convertor = (*ClashRuleProvider)(nil)

type ClashRuleProvider struct{}

func (s *ClashRuleProvider) Type() string {
	return "clash"
}

func (s *ClashRuleProvider) ContentType(options adapter.ConvertOptions) string {
	switch options.Options.TargetType {
	case "mrs":
		return "application/octet-stream"
	default:
		return "plain/text"
	}
}

func (s *ClashRuleProvider) From(ctx context.Context, binary []byte, options adapter.ConvertOptions) (*boxOption.PlainRuleSetCompat, error) {
	//if bytes.HasPrefix(binary, MrsMagicBytes[:]) {
	// TODO: read mrs
	//}
	scanner := bufio.NewScanner(bytes.NewReader(binary))
	if !scanner.Scan() {
		return nil, E.New("empty rule set")
	}
	var (
		yamlPayload bool
		rule        boxOption.DefaultHeadlessRule
		ipPayload   bool
	)
	if strings.TrimSpace(scanner.Text()) == "payload:" {
		yamlPayload = true
	}
	for i := 0; scanner.Scan(); i++ {
		ruleLine := scanner.Text()
		if ruleLine == "" || ruleLine[0] == '#' {
			continue
		}
		if yamlPayload {
			ruleLine = strings.TrimSpace(common.SubstringAfter(ruleLine, "-"))
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
				// unsupported rule
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
		Version: C.RuleSetVersionCurrent,
		Options: boxOption.PlainRuleSet{
			Rules: []boxOption.HeadlessRule{{
				Type:           C.RuleTypeDefault,
				DefaultOptions: rule,
			}},
		},
	}, nil
}

func (s *ClashRuleProvider) To(ctx context.Context, source *boxOption.PlainRuleSetCompat, options adapter.ConvertOptions) ([]byte, error) {
	targetFormat := options.Options.ClashOptions.TargetFormat
	var output bytes.Buffer
	switch targetFormat {
	case "", "yaml":
		output.WriteString("payload:\n")
		for _, rule := range source.Options.Rules[0].DefaultOptions.IPCIDR {
			output.WriteString("- " + rule + "\n")
		}
		for _, rule := range source.Options.Rules[0].DefaultOptions.Domain {
			output.WriteString("- " + rule + "\n")
		}
		for _, rule := range source.Options.Rules[0].DefaultOptions.DomainSuffix {
			output.WriteString("- +." + rule + "\n")
		}
	case "text":
		for _, rule := range source.Options.Rules[0].DefaultOptions.IPCIDR {
			output.WriteString(rule + "\n")
		}
		for _, rule := range source.Options.Rules[0].DefaultOptions.Domain {
			output.WriteString(rule + "\n")
		}
		for _, rule := range source.Options.Rules[0].DefaultOptions.DomainSuffix {
			output.WriteString("+." + rule + "\n")
		}
	case "mrs":
		encoder, err := zstd.NewWriter(&output, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		if err != nil {
			return nil, err
		}
		_, err = encoder.Write(MrsMagicBytes[:])
		if err != nil {
			return nil, err
		}
		if len(source.Options.Rules[0].DefaultOptions.IPCIDR) > 0 {
			encoder.Write([]byte{1}) // IPCIDR

			err = binary.Write(encoder, binary.BigEndian, int64(len(source.Options.Rules[0].DefaultOptions.IPCIDR)))
			if err != nil {
				return nil, err
			}
		} else {
			encoder.Write([]byte{0}) // Domain

			err = binary.Write(encoder, binary.BigEndian, int64(len(source.Options.Rules[0].DefaultOptions.Domain)+len(source.Options.Rules[0].DefaultOptions.DomainSuffix)))
			if err != nil {
				return nil, err
			}
		}
		err = binary.Write(encoder, binary.BigEndian, int64(0))
		if err != nil {
			return nil, err
		}
		if len(source.Options.Rules[0].DefaultOptions.IPCIDR) > 0 {
			ipCidrTrie := cidr.NewIpCidrSet()
			for _, rule := range source.Options.Rules[0].DefaultOptions.IPCIDR {
				ipCidrTrie.AddIpCidrForString(rule)
			}
			err = ipCidrTrie.WriteBin(encoder)
			if err != nil {
				return nil, E.Cause(err, "compile mrs")
			}
		} else {
			domainTrie := trie.New[struct{}]()
			for _, rule := range source.Options.Rules[0].DefaultOptions.Domain {
				domainTrie.Insert(rule, struct{}{})
			}
			for _, rule := range source.Options.Rules[0].DefaultOptions.DomainSuffix {
				domainTrie.Insert("+."+rule, struct{}{})
			}
			domainSet := domainTrie.NewDomainSet()
			err = domainSet.WriteBin(encoder)
			if err != nil {
				return nil, E.Cause(err, "compile mrs")
			}
		}
		err = encoder.Close()
		if err != nil {
			return nil, err
		}
	default:
		return nil, E.New("unknown target format: " + targetFormat)
	}
	return output.Bytes(), nil
}
