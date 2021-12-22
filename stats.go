package eznode

import "github.com/google/uuid"

type ChainNodeStats struct {
	Id     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Hits   uint      `json:"hits"`
	Limits uint      `json:"limits"`
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
			Id:     node.id,
			Name:   node.name,
			Hits:   node.hits,
			Limits: node.limit.Count,
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
