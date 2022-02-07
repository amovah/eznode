package eznode

import (
	"log"
	"sync"
	"time"
)

// Option is a functional parameter for NewEzNode
type Option func(*EzNode)

// NewEzNode creates a new EzNode
func NewEzNode(chains []*Chain, options ...Option) *EzNode {
	chainHashMap := make(map[string]*Chain)
	for _, userChain := range chains {
		chainHashMap[userChain.id] = userChain
	}

	ezNode := &EzNode{
		chains: chainHashMap,
		apiCaller: &apiCallerClient{
			client: createHttpClient(),
		},
		syncStorage: syncStorage{
			interval: 60 * time.Second,
			ticker:   &time.Ticker{},
			done:     make(chan bool),
			isRun:    false,
			mutex:    &sync.Mutex{},
		},
	}

	for _, option := range options {
		option(ezNode)
	}

	return ezNode
}

// WithApiClient sets the api client
func WithApiClient(apiCaller ApiCaller) Option {
	return func(ezNode *EzNode) {
		ezNode.apiCaller = apiCaller
	}
}

// WithSyncInterval sets the sync interval for calling sync stats function
func WithSyncInterval(
	interval time.Duration,
) Option {
	if interval <= 0 {
		log.Fatal("save interval cannot be less than 0 [less than 10 seconds is not recommended]")
	}

	return func(ezNode *EzNode) {
		ezNode.syncStorage.interval = interval
	}
}
