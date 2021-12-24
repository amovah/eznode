package eznode

import (
	"sync"
	"time"
)

type syncStorage struct {
	interval time.Duration
	ticker   *time.Ticker
	done     chan bool
	isRun    bool
	mutex    *sync.Mutex
}
