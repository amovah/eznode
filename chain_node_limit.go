package eznode

import (
	"time"
)

// ChainNodeLimit determine the limit of node, how many request can be sent to one node
type ChainNodeLimit struct {
	// Count is max number of request can be sent to this node per Per
	Count uint
	// Per is time period of the limit
	Per time.Duration
}
