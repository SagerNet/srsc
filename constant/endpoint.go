package constant

import "time"

const DefaultTTL = 600 * time.Second

const (
	EndpointTypeFile     = "file"
	EndpointSourceLocal  = "local"
	EndpointSourceRemote = "remote"
)
