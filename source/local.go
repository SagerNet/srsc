package source

import (
	"context"
	"os"
	"text/template"
	"time"

	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/srsc/adapter"
	"github.com/sagernet/srsc/option"
)

var _ adapter.Source = (*Local)(nil)

type Local struct {
	pathTemplate *template.Template
}

func NewLocal(ctx context.Context, options option.SourceOptions) (*Local, error) {
	pathTemplate, err := template.New("local path").Parse(options.LocalOptions.Path)
	if err != nil {
		return nil, err
	}
	return &Local{
		pathTemplate: pathTemplate,
	}, nil
}

func (s *Local) Path(urlParams map[string]string) (sourcePath string, err error) {
	pathBuffer := buf.New()
	defer pathBuffer.Release()
	err = s.pathTemplate.Execute(pathBuffer, urlParams)
	if err != nil {
		return
	}
	sourcePath = string(pathBuffer.Bytes())
	return
}

func (s *Local) LastUpdated(path string) time.Time {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return fileInfo.ModTime()
}

func (s *Local) Fetch(path string, requestBody adapter.FetchRequestBody) (body *adapter.FetchResponseBody, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	return &adapter.FetchResponseBody{
		Content:     content,
		LastUpdated: s.LastUpdated(path),
	}, nil
}
