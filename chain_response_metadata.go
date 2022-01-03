package eznode

import "time"

type ChainResponseMetadata struct {
	ChainId      string
	RequestedUrl string
	Retry        int
	Trace        []NodeTrace
}

type NodeTrace struct {
	NodeName   string
	StatusCode int
	Err        error
	Time       time.Time
}
