package eznode

type ChainResponseMetadata struct {
	ChainId      string
	RequestedUrl string
	Retry        int
	ErrorTrace   []NodeErrorTrace
}

type NodeErrorTrace struct {
	NodeName string
	NodeId   string
	Err      error
}
