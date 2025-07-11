package source

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"text/template"
	"time"

	boxAdapter "github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/dns/transport/local"
	boxOption "github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/buf"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	aTLS "github.com/sagernet/sing/common/tls"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/option"
)

var _ adapter.Source = (*Remote)(nil)

type Remote struct {
	ctx          context.Context
	pathTemplate *template.Template
	httpClient   *http.Client
	userAgent    string
	ttl          time.Duration
}

func NewRemote(ctx context.Context, options option.SourceOptions) (*Remote, error) {
	pathTemplate := template.New("remote URL")
	pathTemplate.Funcs(template.FuncMap{
		"toLower": strings.ToLower,
		"toUpper": strings.ToUpper,
	})
	_, err := pathTemplate.Parse(options.RemoteOptions.URL)
	if err != nil {
		return nil, err
	}
	var serverAddress string
	if serverURL, err := url.Parse(options.RemoteOptions.URL); err == nil {
		if hostname := serverURL.Hostname(); M.IsDomainName(hostname) {
			serverAddress = hostname
		}
	}
	remoteDialer, err := dialer.NewDefault(ctx, options.RemoteOptions.DialerOptions)
	if err != nil {
		return nil, err
	}
	var tlsConfig tls.Config
	if options.RemoteOptions.TLS != nil && options.RemoteOptions.TLS.Enabled {
		tlsConfig, err = tls.NewClient(ctx, serverAddress, common.PtrValueOrDefault(options.RemoteOptions.TLS))
		if err != nil {
			return nil, E.Cause(err, "create TLS config")
		}
	}
	dnsTransport := common.Must1(local.NewTransport(ctx, logger.NOP(), "", boxOption.LocalDNSServerOptions{}))
	dnsClient := dns.NewClient(dns.ClientOptions{
		Logger: logger.NOP(),
	})
	var httpTransport *http.Transport
	if tlsConfig != nil {
		httpTransport = &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				destination := M.ParseSocksaddr(addr)
				var conn net.Conn
				if destination.IsFqdn() {
					var addresses []netip.Addr
					addresses, err = dnsClient.Lookup(ctx, dnsTransport, destination.Fqdn, boxAdapter.DNSQueryOptions{}, nil)
					if err != nil {
						return nil, err
					}
					conn, err = N.DialParallel(ctx, remoteDialer, network, destination, addresses, false, 0)
				} else {
					conn, err = remoteDialer.DialContext(ctx, network, destination)
				}
				if err != nil {
					return nil, err
				}
				tlsConn, err := aTLS.ClientHandshake(ctx, conn, tlsConfig)
				if err != nil {
					conn.Close()
					return nil, err
				}
				return tlsConn, nil
			},
			ForceAttemptHTTP2: true,
		}
	} else {
		httpTransport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				destination := M.ParseSocksaddr(addr)
				var conn net.Conn
				if destination.IsFqdn() {
					var addresses []netip.Addr
					addresses, err = dnsClient.Lookup(ctx, dnsTransport, destination.Fqdn, boxAdapter.DNSQueryOptions{}, nil)
					if err != nil {
						return nil, err
					}
					conn, err = N.DialParallel(ctx, remoteDialer, network, destination, addresses, false, 0)
				} else {
					conn, err = remoteDialer.DialContext(ctx, network, destination)
				}
				if err != nil {
					return nil, err
				}
				return conn, nil
			},
			ForceAttemptHTTP2: true,
		}
	}
	var userAgent string
	if options.RemoteOptions.UserAgent != "" {
		userAgent = options.RemoteOptions.UserAgent
	} else {
		userAgent = F.ToString("srsc/", C.Version, "(sing-box ", C.CoreVersion(), ")")
	}
	var ttl time.Duration
	if options.RemoteOptions.TTL > 0 {
		ttl = options.RemoteOptions.TTL.Build()
	} else {
		ttl = C.DefaultTTL
	}
	return &Remote{
		ctx:          ctx,
		pathTemplate: pathTemplate,
		httpClient: &http.Client{
			Transport: httpTransport,
		},
		userAgent: userAgent,
		ttl:       ttl,
	}, nil
}

func (s *Remote) Path(urlParams map[string]string) (sourcePath string, err error) {
	pathBuffer := buf.New()
	defer pathBuffer.Release()
	err = s.pathTemplate.Execute(pathBuffer, urlParams)
	if err != nil {
		return
	}
	sourcePath = string(pathBuffer.Bytes())
	return
}

func (s *Remote) LastUpdated(_ string) time.Time {
	return time.Time{}
}

func (s *Remote) Fetch(path string, requestBody adapter.FetchRequestBody) (body *adapter.FetchResponseBody, err error) {
	if time.Now().Sub(requestBody.LastUpdated) < s.ttl {
		return &adapter.FetchResponseBody{
			NotModified: true,
			LastUpdated: requestBody.LastUpdated,
		}, nil
	}
	request, err := http.NewRequestWithContext(s.ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, E.Cause(err, "create HTTP request")
	}
	request.Header.Set("User-Agent", s.userAgent)
	if requestBody.ETag != "" {
		request.Header.Set("If-None-Match", requestBody.ETag)
	}
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, E.Cause(err, "fetch source: exchange HTTP request")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotModified {
		return &adapter.FetchResponseBody{
			NotModified: true,
			LastUpdated: time.Now(),
		}, nil
	} else if response.StatusCode != http.StatusOK {
		return nil, E.New("fetch source: unexpected HTTP response: " + response.Status)
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		err = E.Cause(err, "fetch source: read HTTP response")
		return
	}
	newETag := response.Header.Get("ETag")
	return &adapter.FetchResponseBody{
		Content:     content,
		ETag:        newETag,
		LastUpdated: time.Now(),
	}, nil
}
