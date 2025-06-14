package srsc

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/sagernet/sing-box/common/listener"
	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/log"
	boxOption "github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/logger"
	N "github.com/sagernet/sing/common/network"
	aTLS "github.com/sagernet/sing/common/tls"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/srsc/adapter"
	"github.com/sagernet/srsc/cache"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/endpoint"
	"github.com/sagernet/srsc/option"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/http2"
)

type Server struct {
	createdAt  time.Time
	ctx        context.Context
	logger     logger.ContextLogger
	logFactory log.Factory
	listener   *listener.Listener
	tlsConfig  tls.ServerConfig
	httpServer *http.Server
	cache      adapter.Cache
}

type Options struct {
	Context context.Context
	Logger  logger.ContextLogger
	option.Options
}

func NewServer(options Options) (*Server, error) {
	createdAt := time.Now()
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = service.ContextWithDefaultRegistry(ctx)
	if options.Logger == nil {
		logFactory, err := log.New(log.Options{
			Context:  ctx,
			Options:  common.PtrValueOrDefault(options.Log),
			BaseTime: createdAt,
		})
		if err != nil {
			return nil, E.Cause(err, "create log factory")
		}
		options.Logger = logFactory.Logger()
		// TODO: improve log
	}
	serviceCache, err := cache.New(ctx, common.PtrValueOrDefault(options.Cache))
	if err != nil {
		return nil, E.Cause(err, "create cache")
	}
	service.MustRegister[adapter.Cache](ctx, serviceCache)
	chiRouter := chi.NewRouter()
	s := &Server{
		createdAt: createdAt,
		ctx:       ctx,
		logger:    options.Logger,
		listener: listener.New(listener.Options{
			Context: ctx,
			Logger:  options.Logger,
			Network: []string{N.NetworkTCP},
			Listen: boxOption.ListenOptions{
				Listen:     options.Listen,
				ListenPort: options.ListenPort,
			},
		}),
		httpServer: &http.Server{
			Handler: chiRouter,
		},
		cache: serviceCache,
	}
	if options.Endpoints == nil || options.Endpoints.Size() == 0 {
		return nil, E.New("missing endpoints")
	}
	for index, entry := range options.Endpoints.Entries() {
		switch entry.Value.Type {
		case C.EndpointTypeFile:
			handler, err := endpoint.NewFileEndpoint(ctx, options.Logger, index, entry.Value.FileOptions)
			if err != nil {
				return nil, err
			}
			chiRouter.Get(entry.Key, handler.ServeHTTP)
		default:
			return nil, E.New("unknown endpoint type: " + entry.Value.Type)
		}
	}
	if options.TLS != nil {
		tlsConfig, err := tls.NewServer(ctx, options.Logger, common.PtrValueOrDefault(options.TLS))
		if err != nil {
			return nil, err
		}
		s.tlsConfig = tlsConfig
	}
	return s, nil
}

func (s *Server) Start() error {
	if s.cache != nil {
		err := s.cache.Start()
		if err != nil {
			return E.Cause(err, "start cache")
		}
	}
	if s.tlsConfig != nil {
		err := s.tlsConfig.Start()
		if err != nil {
			return E.Cause(err, "create TLS config")
		}
	}
	tcpListener, err := s.listener.ListenTCP()
	if err != nil {
		return err
	}
	if s.tlsConfig != nil {
		if !common.Contains(s.tlsConfig.NextProtos(), http2.NextProtoTLS) {
			s.tlsConfig.SetNextProtos(append([]string{"h2"}, s.tlsConfig.NextProtos()...))
		}
		tcpListener = aTLS.NewListener(tcpListener, s.tlsConfig)
	}
	go func() {
		err = s.httpServer.Serve(tcpListener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("serve error: ", err)
		}
	}()
	s.logger.Info("srsc started (", F.Seconds(time.Since(s.createdAt).Seconds()), "s)")
	return nil
}

func (s *Server) Close() error {
	return common.Close(
		common.PtrOrNil(s.httpServer),
		common.PtrOrNil(s.listener),
		s.tlsConfig,
		s.cache,
	)
}
