package eznode

import (
	"sync/atomic"
)

type ChainNodeStats struct {
	Name        string `json:"name"`
	CurrentHits uint   `json:"current_hits"`
	TotalHits   uint64 `json:"total_hits"`
	Limits      uint   `json:"limits"`
	Priority    int    `json:"priority"`
}

type ChainStats struct {
	Id    string           `json:"id"`
	Nodes []ChainNodeStats `json:"nodes"`
}

func (c *Chain) getStats() []ChainNodeStats {
	c.mutex.RLock()
	nodeStats := make([]ChainNodeStats, 0)
	for _, node := range c.nodes {
		nodeStats = append(nodeStats, ChainNodeStats{
			Name:        node.name,
			CurrentHits: node.hits,
			TotalHits:   atomic.LoadUint64(&node.totalHits),
			Limits:      node.limit.Count,
			Priority:    node.priority,
		})
	}
	c.mutex.RUnlock()

	return nodeStats
}

func (e *EzNode) GetStats() []ChainStats {
	chainStats := make([]ChainStats, 0)

	for _, chain := range e.chains {
		chainStats = append(chainStats, ChainStats{
			Id:    chain.id,
			Nodes: chain.getStats(),
		})
	}

	return chainStats
}
