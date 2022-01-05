package eznode

import (
	"time"
)

type StatsSyncer func([]ChainStats)

func (e *EzNode) StartSyncStats(syncer StatsSyncer) {
	e.syncStorage.mutex.Lock()
	defer e.syncStorage.mutex.Unlock()

	if !e.syncStorage.isRun {
		e.syncStorage.ticker = time.NewTicker(e.syncStorage.interval)

		go func() {
			for {
				select {
				case <-e.syncStorage.done:
					return
				case <-e.syncStorage.ticker.C:
					if syncer != nil {
						syncer(e.GetStats())
					}
				}
			}
		}()

		e.syncStorage.isRun = true
	}
}

func (e *EzNode) StopSyncStats() {
	e.syncStorage.mutex.Lock()
	defer e.syncStorage.mutex.Unlock()

	if e.syncStorage.isRun {
		e.syncStorage.ticker.Stop()
		e.syncStorage.done <- true
		e.syncStorage.isRun = false
	}
}

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