package clash

import (
	"net/netip"
	"strconv"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/ranges"
)

func ToSurgeLines(rule *option.HeadlessRule) []string {
	if rule.Type == C.RuleTypeLogical {
		var subRules []string
		for _, subRule := range rule.LogicalOptions.Rules {
			subRules = append(subRules, "("+strings.Join(ToSurgeLines(&subRule), ",")+")")
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
			if prefix, err := netip.ParsePrefix(ipCidr); err == nil {
				if prefix.Addr().Is6() {
					lines = append(lines, "IP-CIDR6,"+ipCidr)
				} else {
					lines = append(lines, "IP-CIDR,"+ipCidr)
				}
			}
			if addr, err := netip.ParseAddr(ipCidr); err == nil {
				addrPrefix := netip.PrefixFrom(addr, addr.BitLen())
				if addrPrefix.Addr().Is6() {
					lines = append(lines, "IP-CIDR6,"+addrPrefix.String())
				} else {
					lines = append(lines, "IP-CIDR,"+addrPrefix.String())
				}
			}
		}
		for _, sourceIPCIDR := range rule.DefaultOptions.SourceIPCIDR {
			if addr, err := netip.ParseAddr(sourceIPCIDR); err == nil {
				lines = append(lines, "SRC-IP,"+netip.PrefixFrom(addr, addr.BitLen()).String())
			} else {
				lines = append(lines, "SRC-IP,"+sourceIPCIDR)
			}
		}
		for _, port := range rule.DefaultOptions.Port {
			lines = append(lines, "DEST-PORT,"+F.ToString(port))
		}
		if len(rule.DefaultOptions.PortRange) > 0 {
			rangeList, err := convertPortRangeList(rule.DefaultOptions.PortRange)
			if err != nil {
				return nil
			}
			lines = append(lines, common.Map(rangeList, func(it string) string {
				return "DEST-PORT," + it
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
		return lines
	}
}

func FromSurgeLine(ruleLine string) (*option.HeadlessRule, error) {
	ruleType, payload, _ := parseRule(ruleLine)
	var boxRule option.DefaultHeadlessRule
	switch ruleType {
	case "DOMAIN":
		boxRule.Domain = append(boxRule.Domain, payload)
	case "DOMAIN-SUFFIX":
		boxRule.DomainSuffix = append(boxRule.DomainSuffix, payload)
	case "DOMAIN-KEYWORD":
		boxRule.DomainKeyword = append(boxRule.DomainKeyword, payload)
	case "DOMAIN-REGEX":
		boxRule.DomainRegex = append(boxRule.DomainRegex, payload)
	case "IP-CIDR", "IP-CIDR6":
		boxRule.IPCIDR = append(boxRule.IPCIDR, payload)
	case "SRC-IP":
		boxRule.SourceIPCIDR = append(boxRule.SourceIPCIDR, payload)
	case "SRC-PORT":
		portRange, err := parseSurgePortRange(payload)
		if err != nil {
			return nil, err
		}
		if portRange.Start == portRange.End {
			boxRule.SourcePort = append(boxRule.SourcePort, portRange.Start)
		} else {
			boxRule.SourcePortRange = append(boxRule.SourcePortRange, F.ToString(portRange.Start, ":", portRange.End))
		}
	case "DEST-PORT":
		portRange, err := parseSurgePortRange(payload)
		if err != nil {
			return nil, err
		}
		if portRange.Start == portRange.End {
			boxRule.Port = append(boxRule.Port, portRange.Start)
		} else {
			boxRule.PortRange = append(boxRule.PortRange, F.ToString(portRange.Start, ":", portRange.End))
		}
	case "PROCESS-NAME":
		boxRule.ProcessName = append(boxRule.ProcessName, payload)
	case "SUBNET":
		subnetType := common.SubstringBefore(payload, ":")
		subnetValue := common.SubstringAfter(payload, ":")
		switch subnetType {
		case "SSID":
			boxRule.WIFISSID = append(boxRule.WIFISSID, subnetValue)
		case "BSSID":
			boxRule.WIFIBSSID = append(boxRule.WIFIBSSID, subnetValue)
		case "TYPE":
			switch subnetValue {
			case "WIFI":
				boxRule.NetworkType = append(boxRule.NetworkType, option.InterfaceType(C.InterfaceTypeWIFI))
			case "WIRED":
				boxRule.NetworkType = append(boxRule.NetworkType, option.InterfaceType(C.InterfaceTypeEthernet))
			case "CELLULAR":
				boxRule.NetworkType = append(boxRule.NetworkType, option.InterfaceType(C.InterfaceTypeCellular))
			default:
				return nil, E.New("unsupported subnet type: ", subnetValue)
			}
		default:
			return nil, E.New("unsupported subnet rule: ", ruleLine)
		}
	case "AND", "OR", "NOT":
		return parseLogicLine(ruleType, payload, FromSurgeLine)
	default:
		return nil, E.New("unsupported rule type: ", ruleType)
	}
	return &option.HeadlessRule{
		Type:           C.RuleTypeDefault,
		DefaultOptions: boxRule,
	}, nil
}

func parseSurgePortRange(portString string) (ranges.Range[uint16], error) {
	if portValue, err := strconv.ParseUint(portString, 10, 16); err == nil {
		return ranges.New(uint16(portValue), uint16(portValue)), nil
	}
	if strings.Contains(portString, "-") {
		portStart := common.SubstringBefore(portString, "-")
		portEnd := common.SubstringAfter(portString, "-")
		portStartValue, err := strconv.ParseUint(portStart, 10, 16)
		if err != nil {
			return ranges.Range[uint16]{}, E.Cause(err, "invalid port range: ", portString)
		}
		portEndValue, err := strconv.ParseUint(portEnd, 10, 16)
		if err != nil {
			return ranges.Range[uint16]{}, E.Cause(err, "invalid port range: ", portString)
		}
		if portStartValue > portEndValue {
			return ranges.Range[uint16]{}, E.Cause(err, "invalid port range: ", portString)
		}
		return ranges.New(uint16(portStartValue), uint16(portEndValue)), nil
	} else if strings.HasPrefix(portString, "<=") {
		portValue, err := strconv.ParseUint(strings.TrimPrefix(portString, "<="), 10, 16)
		if err != nil {
			return ranges.Range[uint16]{}, E.Cause(err, "invalid port range: ", portString)
		}
		return ranges.New(0, uint16(portValue)), nil
	} else if strings.HasPrefix(portString, "<") {
		portValue, err := strconv.ParseUint(strings.TrimPrefix(portString, "<"), 10, 16)
		if err != nil {
			return ranges.Range[uint16]{}, E.Cause(err, "invalid port range: ", portString)
		}
		return ranges.New(0, uint16(portValue-1)), nil
	} else if strings.HasPrefix(portString, ">=") {
		portValue, err := strconv.ParseUint(strings.TrimPrefix(portString, ">="), 10, 16)
		if err != nil {
			return ranges.Range[uint16]{}, E.Cause(err, "invalid port range: ", portString)
		}
		return ranges.New(uint16(portValue), 65535), nil
	} else if strings.HasPrefix(portString, ">") {
		portValue, err := strconv.ParseUint(strings.TrimPrefix(portString, ">"), 10, 16)
		if err != nil {
			return ranges.Range[uint16]{}, E.Cause(err, "invalid port range: ", portString)
		}
		return ranges.New(uint16(portValue+1), 65535), nil
	}
	return ranges.Range[uint16]{}, E.New("invalid port range: ", portString)
}
