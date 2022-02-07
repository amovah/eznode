package eznode

import "fmt"

// EzNodeError is the error type for EzNode
type EzNodeError struct {
	// Message is the error message
	Message string
	// Metadata is the error metadata
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
