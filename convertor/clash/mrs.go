package clash

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/netip"
	"sort"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/rw"
	"github.com/sagernet/srsc/adapter"
	"github.com/sagernet/srsc/convertor/internal/meta_cidr"
	"github.com/sagernet/srsc/convertor/internal/meta_domainset"

	"github.com/klauspost/compress/zstd"
	"golang.org/x/exp/slices"
)

var MrsMagicBytes = [4]byte{'M', 'R', 'S', 1} // MRSv1

func fromMrs(content []byte) ([]adapter.Rule, error) {
	decoder, err := zstd.NewReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	defer decoder.Close()
	var header [4]byte
	_, err = io.ReadFull(decoder, header[:])
	if err != nil {
		return nil, err
	}
	if header != MrsMagicBytes {
		return nil, E.New("invalid MrsMagic bytes")
	}
	var behavior byte
	err = binary.Read(decoder, binary.BigEndian, &behavior)
	if err != nil {
		return nil, err
	}
	var length int64
	err = binary.Read(decoder, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, E.New("invalid reserved length: ", length)
	} else if length > 0 {
		err = rw.SkipN(decoder, int(length))
		if err != nil {
			return nil, E.Cause(err, "discard reserved bytes")
		}
	}
	switch behavior {
	case 0:
		var domainSet *trie.DomainSet
		domainSet, err = trie.ReadDomainSetBin(decoder)
		if err != nil {
			return nil, err
		}
		var keys []string
		domainSet.Foreach(func(key string) bool {
			keys = append(keys, key)
			return true
		})
		sort.Strings(keys)
		var rule adapter.DefaultRule
		for _, key := range keys {
			if _, ok := slices.BinarySearch(keys, "+."+key); ok {
				continue
			}
			if strings.HasPrefix(key, "+.") {
				rule.DomainSuffix = append(rule.DomainSuffix, strings.TrimPrefix(key, "+."))
			} else {
				if strings.Contains(key, "+") || strings.Contains(key, "*") {
					continue
				}
				rule.Domain = append(rule.Domain, key)
			}
		}
		return []adapter.Rule{{Type: C.RuleTypeDefault, DefaultOptions: rule}}, nil
	case 1:
		var ipCidrSet *cidr.IpCidrSet
		ipCidrSet, err = cidr.ReadIpCidrSet(decoder)
		if err != nil {
			return nil, err
		}
		return []adapter.Rule{{
			Type: C.RuleTypeDefault,
			DefaultOptions: adapter.DefaultRule{
				DefaultHeadlessRule: option.DefaultHeadlessRule{
					IPCIDR: common.Map(ipCidrSet.ToIPSet().Prefixes(), netip.Prefix.String),
				},
			},
		}}, nil
	default:
		return nil, E.New("invalid behavior: ", behavior)
	}
}

func toMrs(behavior string, rules []adapter.Rule) ([]byte, error) {
	var output bytes.Buffer
	encoder, err := zstd.NewWriter(&output, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, err
	}
	_, err = encoder.Write(MrsMagicBytes[:])
	if err != nil {
		return nil, err
	}
	if behavior == "domain" {
		encoder.Write([]byte{0})
		err = binary.Write(encoder, binary.BigEndian, int64(len(rules[0].DefaultOptions.Domain)+len(rules[0].DefaultOptions.DomainSuffix)))
		if err != nil {
			return nil, err
		}
	} else {
		encoder.Write([]byte{1})
		err = binary.Write(encoder, binary.BigEndian, int64(len(rules[0].DefaultOptions.IPCIDR)))
		if err != nil {
			return nil, err
		}
	}
	err = binary.Write(encoder, binary.BigEndian, int64(0))
	if err != nil {
		return nil, err
	}
	if len(rules[0].DefaultOptions.IPCIDR) > 0 {
		ipCidrTrie := cidr.NewIpCidrSet()
		for _, rule := range rules[0].DefaultOptions.IPCIDR {
			ipCidrTrie.AddIpCidrForString(rule)
		}
		err = ipCidrTrie.WriteBin(encoder)
		if err != nil {
			return nil, E.Cause(err, "compile mrs")
		}
	} else {
		domainTrie := trie.New[struct{}]()
		for _, rule := range rules[0].DefaultOptions.Domain {
			domainTrie.Insert(rule, struct{}{})
		}
		for _, rule := range rules[0].DefaultOptions.DomainSuffix {
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
	return output.Bytes(), nil
}
