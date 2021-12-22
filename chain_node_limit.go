package eznode

import (
	"time"
)

type ChainNodeLimit struct {
	Count uint
	Per   time.Duration
}
