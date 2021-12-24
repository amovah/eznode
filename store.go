package eznode

import (
	"github.com/amovah/eznode/storage"
	"sync"
	"time"
)

type store struct {
	storage      storage.Storage
	interval     time.Duration
	ticker       *time.Ticker
	done         chan bool
	isRun        bool
	mutex        *sync.Mutex
	errorHandler storage.ErrorHandler
}
