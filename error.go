package eznode

import "fmt"

type EzNodeError struct {
	Message  string
	Metadata ChainResponseMetadata
}

func (e EzNodeError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf(
			"failed, node id = %s, node name = %s, chain id = %s, requested url %s",
			e.Metadata.nodeId,
			e.Metadata.nodeName,
			e.Metadata.chainId,
			e.Metadata.requestedUrl,
		)
	}

	return e.Message
}
