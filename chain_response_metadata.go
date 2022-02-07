package eznode

import "time"

// ChainResponseMetadata is a structure that contains metadata about the response
type ChainResponseMetadata struct {
	// ChainId is the chain id of the chain that the response is for
	ChainId string
	// RequestedUrl is the url that was requested
	RequestedUrl string
	// Retry is the number of retries that have been attempted
	Retry int
	// Trace is request to response trace
	Trace []NodeTrace
}

// NodeTrace is a structure that contains the trace of a request
type NodeTrace struct {
	// NodeName is the node that the request was sent to
	NodeName string
	// StatusCode is the status code of the response
	StatusCode int
	// Err is the error that occurred
	Err error
	// Time is the time that the request was sent
	Time time.Time
}
