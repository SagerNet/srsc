package clash

import (
	"regexp"
	"strconv"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/ranges"
	"github.com/sagernet/srsc/convertor/internal/meta_utils"

	"github.com/bahlo/generic-list-go"
	"golang.org/x/exp/slices"
)

func toClassicalLine(rule *option.HeadlessRule) []string {
	if rule.Type == C.RuleTypeLogical {
		var subRules []string
		for _, subRule := range rule.LogicalOptions.Rules {
			subRules = append(subRules, "("+strings.Join(toClassicalLine(&subRule), ",")+")")
		}
		if rule.LogicalOptions.Mode == C.LogicalTypeAnd {
			if rule.LogicalOptions.Invert {
				return []string{"NOT,(" + strings.Join(subRules, ","), ")"}
			} else {
				return []string{"AND,(" + strings.Join(subRules, ","), ")"}
			}
		} else {
			if rule.LogicalOptions.Invert {
				return []string{"NOT,(AND,(" + strings.Join(subRules, ","), "))"}
			} else {
				return []string{"OR,(" + strings.Join(subRules, ","), ")"}
			}
		}
	} else {
		var lines []string
		for _, domain := range rule.DefaultOptions.Domain {
			lines = append(lines, "DOMAIN,"+domain)
		}
		for _, domainSuffix := range rule.DefaultOptions.DomainSuffix {
			lines = append(lines, "DOMAIN-SUFFIX,"+domainSuffix)
		}
		for _, domainKeyword := range rule.DefaultOptions.DomainKeyword {
			lines = append(lines, "DOMAIN-KEYWORD,"+domainKeyword)
		}
		for _, domainRegex := range rule.DefaultOptions.DomainRegex {
			lines = append(lines, "DOMAIN-REGEX,"+domainRegex)
		}
		for _, ipCidr := range rule.DefaultOptions.IPCIDR {
			lines = append(lines, "IP-CIDR,"+ipCidr)
		}
		for _, sourceIPCIDR := range rule.DefaultOptions.SourceIPCIDR {
			lines = append(lines, "SRC-IP-CIDR,"+sourceIPCIDR)
		}
		for _, port := range rule.DefaultOptions.Port {
			lines = append(lines, "DST-PORT,"+F.ToString(port))
		}
		if len(rule.DefaultOptions.PortRange) > 0 {
			rangeList, err := convertPortRangeList(rule.DefaultOptions.PortRange)
			if err != nil {
				return nil
			}
			lines = append(lines, common.Map(rangeList, func(it string) string {
				return "DST-PORT," + it
			})...)
		}
		for _, sourcePort := range rule.DefaultOptions.SourcePort {
			lines = append(lines, "SRC-PORT,"+F.ToString(sourcePort))
		}
		if len(rule.DefaultOptions.SourcePortRange) > 0 {
			rangeList, err := convertPortRangeList(rule.DefaultOptions.SourcePortRange)
			if err != nil {
				return nil
			}
			lines = append(lines, common.Map(rangeList, func(it string) string {
				return "SRC-PORT," + it
			})...)
		}
		for _, processName := range rule.DefaultOptions.ProcessName {
			lines = append(lines, "PROCESS-NAME,"+processName)
		}
		for _, processPath := range rule.DefaultOptions.ProcessPath {
			lines = append(lines, "PROCESS-PATH,"+processPath)
		}
		for _, processPathRegex := range rule.DefaultOptions.ProcessPathRegex {
			lines = append(lines, "PROCESS-PATH-REGEX,"+processPathRegex)
		}
		for _, network := range rule.DefaultOptions.Network {
			switch strings.ToLower(network) {
			case N.NetworkTCP:
				lines = append(lines, "NETWORK,TCP")
			case N.NetworkUDP:
				lines = append(lines, "NETWORK,UDP")
			default:
				return nil
			}
		}
		return lines
	}
}

func fromClassicalLine(ruleLine string) (*option.HeadlessRule, error) {
	ruleType, payload, params := parseRule(ruleLine)
	var boxRule option.DefaultHeadlessRule
	switch ruleType {
	case "MATCH", "RULE-SET", "SUB-RULE":
		return nil, E.New("unsupported rule type on classical rule-set: ", ruleType)
	case "GEOSITE",
		"GEOIP", "SRC-GEOIP",
		"IP-ASN", "SRC-IP-ASN",
		"IP-SUFFIX", "SRC-IP-SUFFIX",
		"IN-PORT",
		"DSCP",
		"PROCESS-NAME-REGEX",
		"UID",
		"IN-TYPE",
		"IN-USER",
		"IN-NAME":
		return nil, E.New("unsupported rule type in sing-box: ", ruleType)
	case "DOMAIN":
		boxRule.Domain = append(boxRule.Domain, payload)
	case "DOMAIN-SUFFIX":
		boxRule.DomainSuffix = append(boxRule.DomainSuffix, payload)
	case "DOMAIN-KEYWORD":
		boxRule.DomainKeyword = append(boxRule.DomainKeyword, payload)
	case "DOMAIN-REGEX":
		boxRule.DomainRegex = append(boxRule.DomainRegex, payload)
	case "IP-CIDR", "IP-CIDR6":
		isSrc := slices.Contains(params, "src")
		if isSrc {
			boxRule.SourceIPCIDR = append(boxRule.SourceIPCIDR, payload)
		} else {
			boxRule.IPCIDR = append(boxRule.IPCIDR, payload)
		}
	case "SRC-IP-CIDR":
		boxRule.SourceIPCIDR = append(boxRule.SourceIPCIDR, payload)
	case "SRC-PORT":
		portRanges, err := utils.NewUnsignedRanges[uint16](payload)
		if err == nil {
			return nil, err
		}
		for _, portRange := range portRanges {
			if portRanges[0].Start() == portRanges[0].End() {
				boxRule.SourcePort = append(boxRule.SourcePort, portRange.Start())
			} else {
				boxRule.SourcePortRange = append(boxRule.SourcePortRange, F.ToString(portRange.Start(), ":", portRange.End()))
			}
		}
	case "DST-PORT":
		portRanges, err := utils.NewUnsignedRanges[uint16](payload)
		if err == nil {
			return nil, err
		}
		for _, portRange := range portRanges {
			if portRanges[0].Start() == portRanges[0].End() {
				boxRule.Port = append(boxRule.Port, portRange.Start())
			} else {
				boxRule.PortRange = append(boxRule.PortRange, F.ToString(portRange.Start(), ":", portRange.End()))
			}
		}
	case "PROCESS-NAME":
		boxRule.ProcessName = append(boxRule.ProcessName, payload)
	case "PROCESS-PATH":
		boxRule.ProcessPath = append(boxRule.ProcessPath, payload)
	case "PROCESS-PATH-REGEX":
		boxRule.ProcessPathRegex = append(boxRule.ProcessPathRegex, payload)
	case "NETWORK":
		switch strings.ToLower(payload) {
		case N.NetworkTCP:
			boxRule.Network = append(boxRule.Network, N.NetworkTCP)
		case N.NetworkUDP:
			boxRule.Network = append(boxRule.Network, N.NetworkUDP)
		default:
			return nil, E.New("unknown network: ", payload)
		}
	case "AND", "OR", "NOT":
		return parseLogicLine(ruleType, payload)
	}
	return &option.HeadlessRule{
		Type:           C.RuleTypeDefault,
		DefaultOptions: boxRule,
	}, nil
}

func parseRule(ruleRaw string) (string, string, []string) {
	item := strings.Split(ruleRaw, ",")
	if len(item) == 1 {
		return "", item[0], nil
	} else if len(item) == 2 {
		return item[0], item[1], nil
	} else if len(item) > 2 {
		if item[0] == "NOT" || item[0] == "OR" || item[0] == "AND" || item[0] == "SUB-RULE" || item[0] == "DOMAIN-REGEX" || item[0] == "PROCESS-NAME-REGEX" || item[0] == "PROCESS-PATH-REGEX" {
			return item[0], strings.Join(item[1:], ","), nil
		} else {
			return item[0], item[1], item[2:]
		}
	}
	return "", "", nil
}

func parseLogicLine(name string, payload string) (*option.HeadlessRule, error) {
	regex, err := regexp.Compile("\\(.*\\)")
	if err != nil {
		return nil, err
	}
	if !regex.MatchString(payload) {
		return nil, E.New("payload format error")
	}
	subAllRanges, err := logicFormat(payload)
	if err != nil {
		return nil, err
	}
	subRanges := findSubRuleRange(payload, subAllRanges)
	var rules []option.HeadlessRule
	for _, subRange := range subRanges {
		subPayload := payload[subRange.start+1 : subRange.end]
		subRule, err := fromClassicalLine(subPayload)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *subRule)
	}
	var mode string
	if name == "OR" {
		mode = C.LogicalTypeOr
	} else {
		mode = C.LogicalTypeAnd
	}
	return &option.HeadlessRule{
		Type: C.RuleTypeLogical,
		LogicalOptions: option.LogicalHeadlessRule{
			Mode:   mode,
			Rules:  rules,
			Invert: name == "NOT",
		},
	}, nil
}

type Range struct {
	start int
	end   int
	index int
}

func (r Range) containRange(preStart, preEnd int) bool {
	return preStart < r.start && preEnd > r.end
}

func logicFormat(payload string) ([]Range, error) {
	stack := list.New[Range]()
	num := 0
	subRanges := make([]Range, 0)
	for i, c := range payload {
		if c == '(' {
			sr := Range{
				start: i,
				index: num,
			}

			num++
			stack.PushBack(sr)
		} else if c == ')' {
			if stack.Len() == 0 {
				return nil, E.New("missing '('")
			}

			sr := stack.Back()
			stack.Remove(sr)
			sr.Value.end = i
			subRanges = append(subRanges, sr.Value)
		}
	}
	if stack.Len() != 0 {
		return nil, E.New("format error is missing )")
	}
	sortResult := make([]Range, len(subRanges))
	for _, sr := range subRanges {
		sortResult[sr.index] = sr
	}
	return sortResult, nil
}

func findSubRuleRange(payload string, ruleRanges []Range) []Range {
	payloadLen := len(payload)
	subRuleRange := make([]Range, 0)
	for _, rr := range ruleRanges {
		if rr.start == 0 && rr.end == payloadLen-1 {
			continue
		}
		containInSub := false
		for _, r := range subRuleRange {
			if rr.containRange(r.start, r.end) {
				containInSub = true
				break
			}
		}
		if !containInSub {
			subRuleRange = append(subRuleRange, rr)
		}
	}
	return subRuleRange
}

var errBadPortRange = E.New("bad port range")

func convertPortRangeList(rangeList []string) ([]string, error) {
	var portRangeList []ranges.Range[uint16]
	for _, portRange := range rangeList {
		if !strings.Contains(portRange, ":") {
			return nil, E.Extend(errBadPortRange, portRange)
		}
		subIndex := strings.Index(portRange, ":")
		var start, end uint64
		var err error
		if subIndex > 0 {
			start, err = strconv.ParseUint(portRange[:subIndex], 10, 16)
			if err != nil {
				return nil, E.Cause(err, E.Extend(errBadPortRange, portRange))
			}
		}
		if subIndex == len(portRange)-1 {
			end = 0xFFFF
		} else {
			end, err = strconv.ParseUint(portRange[subIndex+1:], 10, 16)
			if err != nil {
				return nil, E.Cause(err, E.Extend(errBadPortRange, portRange))
			}
		}
		portRangeList = append(portRangeList, ranges.New(uint16(start), uint16(end)))
	}
	portRangeList = ranges.Merge(portRangeList)
	return common.Map(portRangeList, func(it ranges.Range[uint16]) string {
		return F.ToString(it.Start, "-", it.End)
	}), nil
}
