package eznode

import (
	"time"
)

// ChainNodeLimit determine the limit of node, how many request can be sent to one node
type ChainNodeLimit struct {
	Count uint
	Per   time.Duration
}
