package eznode

import "fmt"

type EzNodeError struct {
	Message  string
	Metadata ChainResponseMetadata
}

func (e EzNodeError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf(
			"failed. chain id = %s, requested url %s",
			e.Metadata.ChainId,
			e.Metadata.RequestedUrl,
		)
	}

	return e.Message
}
