package eznode

import (
	"sync/atomic"
)

func (c *Chain) getStats() []ChainNodeStats {
	c.mutex.RLock()
	nodeStats := make([]ChainNodeStats, 0)
	for _, node := range c.nodes {
		node.statsMutex.Lock()
		nodeStats = append(nodeStats, ChainNodeStats{
			Name:          node.name,
			CurrentHits:   node.hits,
			TotalHits:     atomic.LoadUint64(&node.totalHits),
			ResponseStats: node.responseStats,
			Limits:        node.limit.Count,
			Priority:      node.priority,
			Disabled:      node.disabled,
			Fails:         node.fails,
		})
		node.statsMutex.Unlock()
	}
	c.mutex.RUnlock()

	return nodeStats
}

// GetStats returns the stats of chains
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
