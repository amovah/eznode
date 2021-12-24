package storage

import (
	"github.com/amovah/eznode/stats"
)

type TemporaryStorage struct {
}

func NewTemporaryStorage() *TemporaryStorage {
	return &TemporaryStorage{}
}

func (m *TemporaryStorage) Save(_ ErrorHandler, _ []stats.ChainStats) {
}

func (m *TemporaryStorage) Load() ([]stats.ChainStats, error) {
	return []stats.ChainStats{}, nil
}
