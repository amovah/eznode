package storage

import (
	"github.com/amovah/eznode/stats"
)

type ErrorHandler func(error)

type Storage interface {
	Save(ErrorHandler, []stats.ChainStats)
	Load() ([]stats.ChainStats, error)
}
