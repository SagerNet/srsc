package endpoint

import (
	"context"
	"net/http"
	"os"

	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/srsc/adapter"
	C "github.com/sagernet/srsc/constant"
	"github.com/sagernet/srsc/convertor"
	"github.com/sagernet/srsc/option"
	"github.com/sagernet/srsc/source"

	"github.com/go-chi/chi/v5"
)

var _ http.Handler = (*FileEndpoint)(nil)

type FileEndpoint struct {
	ctx             context.Context
	logger          logger.ContextLogger
	cache           adapter.Cache
	index           int
	source          adapter.Source
	sourceConvertor adapter.Convertor
	targetConvertor adapter.Convertor
	convertOptions  option.ConvertOptions
}

func NewFileEndpoint(ctx context.Context, logger logger.ContextLogger, index int, options option.FileEndpoint) (*FileEndpoint, error) {
	ep := &FileEndpoint{
		ctx:            ctx,
		logger:         logger,
		cache:          service.FromContext[adapter.Cache](ctx),
		index:          index,
		convertOptions: options.ConvertOptions,
	}
	endpointSource, err := source.New(ctx, options.SourceOptions)
	if err != nil {
		return nil, E.Cause(err, "create source")
	}
	ep.source = endpointSource
	sourceConvertor, loaded := convertor.Convertors[options.SourceType]
	if !loaded {
		return nil, E.New("unknown source type: ", options.SourceType)
	}
	ep.sourceConvertor = sourceConvertor
	targetConvertor, loaded := convertor.Convertors[options.TargetType]
	if !loaded {
		return nil, E.New("unknown target type: ", options.TargetType)
	}
	ep.targetConvertor = targetConvertor
	return ep, nil
}

func (f *FileEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f.serveHTTP0(w, r)
	if err != nil {
		f.logger.Error("handle ", r.RemoteAddr, " - ", r.Header.Get("User-Agent"), " \"", r.Method, " ", r.URL, " ", r.Proto, "\": ", err)
	} else {
		f.logger.Debug("accepted ", r.RemoteAddr, " - ", r.Header.Get("User-Agent"), " \"", r.Method, " ", r.URL, " ", r.Proto, "\"")
	}
}

func (f *FileEndpoint) serveHTTP0(w http.ResponseWriter, r *http.Request) error {
	convertOptions := adapter.ConvertOptions{
		Options:  f.convertOptions,
		Metadata: C.DetectMetadata(r.UserAgent()),
	}
	var urlParams map[string]string // TODO: improve performance
	rawURLParams := chi.RouteContext(r.Context()).URLParams
	if len(rawURLParams.Keys) > 0 {
		urlParams = make(map[string]string)
		for i, key := range rawURLParams.Keys {
			urlParams[key] = rawURLParams.Values[i]
		}
	}
	cachePath, err := f.source.Path(urlParams)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return E.Cause(err, "evaluate source path")
	}
	cacheKey := F.ToString("file.", f.index, ".", cachePath)
	cachedBinary, err := f.cache.LoadBinary(cacheKey)
	if err != nil && !os.IsNotExist(err) {
		w.WriteHeader(http.StatusInternalServerError)
		return E.Cause(err, "load cache binary")
	}
	lastUpdated := f.source.LastUpdated(cachePath)
	if cachedBinary != nil && !lastUpdated.IsZero() && cachedBinary.LastUpdated.Equal(lastUpdated) {
		return f.writeCache(w, cachedBinary, convertOptions)
	}

	var fetchBody adapter.FetchRequestBody
	if cachedBinary != nil {
		fetchBody.ETag = cachedBinary.LastEtag
		fetchBody.LastUpdated = cachedBinary.LastUpdated
	}
	response, err := f.source.Fetch(cachePath, fetchBody)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return E.Cause(err, "fetch source")
	}
	if response.NotModified {
		if cachedBinary == nil {
			w.WriteHeader(http.StatusBadGateway)
			return E.New("fetch source: unexpected not modified response")
		}
		if response.LastUpdated != cachedBinary.LastUpdated {
			cachedBinary.LastUpdated = response.LastUpdated
			err = f.cache.SaveBinary(cacheKey, cachedBinary)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return E.Cause(err, "save cache binary")
			}
		}
		return f.writeCache(w, cachedBinary, convertOptions)
	}
	if len(response.Content) == 0 {
		w.WriteHeader(http.StatusBadGateway)
		return E.Cause(err, "fetch source: empty content")
	}
	// binary := response.Content
	// if f.sourceConvertor.Type() != f.targetConvertor.Type() {
	var rules []adapter.Rule
	rules, err = f.sourceConvertor.From(f.ctx, response.Content, convertOptions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return E.Cause(err, "decode source")
	}
	binary, err := f.targetConvertor.To(f.ctx, rules, convertOptions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return E.Cause(err, "encode target")
	}
	//}
	cache := &adapter.SavedBinary{
		Content:     binary,
		LastUpdated: response.LastUpdated,
		LastEtag:    response.ETag,
	}
	err = f.cache.SaveBinary(cacheKey, cache)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return E.Cause(err, "save cache binary")
	}
	return f.writeCache(w, cache, convertOptions)
}

func (f *FileEndpoint) writeCache(w http.ResponseWriter, cachedBinary *adapter.SavedBinary, convertOptions adapter.ConvertOptions) error {
	w.Header().Set("Content-Type", f.targetConvertor.ContentType(convertOptions)+"; charset=utf-8")
	w.Header().Set("Content-Length", F.ToString(len(cachedBinary.Content)))
	if cachedBinary.LastEtag != "" {
		w.Header().Set("ETag", cachedBinary.LastEtag)
	}
	_, err := w.Write(cachedBinary.Content)
	if err != nil {
		return E.Cause(err, "write cached content")
	}
	return nil
}
