package eznode

import (
	"github.com/amovah/eznode/stats"
	"github.com/amovah/eznode/storage"
	"log"
	"sync"
	"time"
)

type Option func(*EzNode)

func NewEzNode(chains []*Chain, options ...Option) (*EzNode, error) {
	chainHashMap := make(map[string]*Chain)
	for _, userChain := range chains {
		chainHashMap[userChain.id] = userChain
	}

	ezNode := &EzNode{
		chains: chainHashMap,
		apiCaller: &apiCallerClient{
			client: createHttpClient(),
		},
		store: store{
			storage:      storage.NewTemporaryStorage(),
			interval:     60 * time.Second,
			ticker:       &time.Ticker{},
			done:         make(chan bool),
			isRun:        false,
			mutex:        &sync.Mutex{},
			errorHandler: func(_ error) {},
		},
	}

	for _, option := range options {
		option(ezNode)
	}

	loadedStats, err := ezNode.store.storage.Load()
	if err != nil {
		return nil, err
	}

	loadedStatsMap := make(map[string]map[string]*stats.ChainNodeStats)
	for _, stat := range loadedStats {
		loadedStatsMap[stat.Id] = make(map[string]*stats.ChainNodeStats)
		for _, node := range stat.Nodes {
			loadedStatsMap[stat.Id][node.Name] = &node
		}
	}

	for _, chain := range ezNode.chains {
		if loadedStatsMap[chain.id] != nil {
			for _, node := range chain.nodes {
				loadedNode := loadedStatsMap[chain.id][node.name]
				if loadedNode != nil {
					node.hits = loadedNode.CurrentHits
					node.totalHits = loadedNode.TotalHits
				}
			}
		}
	}

	return ezNode, nil
}

func WithApiClient(apiCaller ApiCaller) Option {
	return func(ezNode *EzNode) {
		ezNode.apiCaller = apiCaller
	}
}

func WithStore(
	storage storage.Storage,
	interval time.Duration,
	errorHandler storage.ErrorHandler,
) Option {
	if interval <= 0 {
		log.Fatal("save interval cannot be less than 0 [less than 10 seconds is not recommended]")
	}

	return func(ezNode *EzNode) {
		ezNode.store.storage = storage
		ezNode.store.interval = interval
		ezNode.store.errorHandler = errorHandler
	}
}
