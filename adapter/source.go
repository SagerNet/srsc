package adapter

import "time"

type Source interface {
	Path(urlParams map[string]string) (sourcePath string, err error)
	LastUpdated(path string) time.Time
	Fetch(path string, requestBody FetchRequestBody) (*FetchResponseBody, error)
}

type FetchRequestBody struct {
	ETag        string
	LastUpdated time.Time
}

type FetchResponseBody struct {
	Content     []byte
	NotModified bool
	ETag        string
	LastUpdated time.Time
}
