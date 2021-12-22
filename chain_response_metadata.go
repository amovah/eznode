package eznode

import "github.com/google/uuid"

type ChainResponseMetadata struct {
	ChainId      string
	RequestedUrl string
	Retry        int
	Trace        []NodeTrace
}

type NodeTrace struct {
	NodeName   string
	NodeId     uuid.UUID
	StatusCode int
	Err        error
}
