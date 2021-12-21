package eznode

import "github.com/google/uuid"

type ChainResponseMetadata struct {
	chainId      string
	nodeName     string
	nodeId       uuid.UUID
	requestedUrl string
	retries      int
}
