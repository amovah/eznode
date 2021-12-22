package eznode

import "github.com/google/uuid"

type ChainResponseMetadata struct {
	ChainId      string
	RequestedUrl string
	Retry        int
	ErrorTrace   []NodeErrorTrace
}

type NodeErrorTrace struct {
	NodeName   string
	NodeId     uuid.UUID
	StatusCode int
	Err        error
}
