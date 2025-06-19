package constant

import "time"

const DefaultTTL = 5 * time.Minute

const (
	EndpointTypeFile     = "file"
	EndpointSourceLocal  = "local"
	EndpointSourceRemote = "remote"
)
