package eznode

// LoadStats loads stats and sync stats to eznode core
func (e *EzNode) LoadStats(loadedStats []ChainStats) {
	loadedStatsMap := make(map[string]map[string]ChainNodeStats)
	for _, stat := range loadedStats {
		loadedStatsMap[stat.Id] = make(map[string]ChainNodeStats)
		for _, node := range stat.Nodes {
			loadedStatsMap[stat.Id][node.Name] = node
		}
	}

	for _, chain := range e.chains {
		chain.mutex.Lock()
		if loadedStatsMap[chain.id] != nil {
			for _, node := range chain.nodes {
				loadedNode := loadedStatsMap[chain.id][node.name]
				node.totalHits = loadedNode.TotalHits
				if loadedNode.ResponseStats != nil {
					node.statsMutex.Lock()
					node.responseStats = loadedNode.ResponseStats
					node.fails = loadedNode.Fails
					node.statsMutex.Unlock()
				}
			}
		}
		chain.mutex.Unlock()
	}
}
