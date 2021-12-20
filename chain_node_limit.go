package eznode

import (
	"time"
)

type ChainNodeLimit struct {
	count uint
	per   time.Duration
}

func NewChainNodeLimit(count uint, per time.Duration) ChainNodeLimit {
	return ChainNodeLimit{
		count: count,
		per:   per,
	}
}
