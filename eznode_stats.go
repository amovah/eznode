package eznode

import (
	"github.com/amovah/eznode/stats"
	"sync/atomic"
)

func (c *Chain) getStats() []stats.ChainNodeStats {
	c.mutex.RLock()
	nodeStats := make([]stats.ChainNodeStats, 0)
	for _, node := range c.nodes {
		nodeStats = append(nodeStats, stats.ChainNodeStats{
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

func (e *EzNode) GetStats() []stats.ChainStats {
	chainStats := make([]stats.ChainStats, 0)

	for _, chain := range e.chains {
		chainStats = append(chainStats, stats.ChainStats{
			Id:    chain.id,
			Nodes: chain.getStats(),
		})
	}

	return chainStats
}
