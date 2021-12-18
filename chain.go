package eznode

import (
	"context"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestMiddleware func(*http.Request) *http.Request
type Requester func(request *http.Request) (*Response, error)

type checkTick struct {
	tickRate         time.Duration
	maxCheckDuration time.Duration
}

type chain struct {
	nodes         []*NodeUnit
	mutex         *sync.RWMutex
	checkTickRate checkTick
}

func createDecoratedRequest(node *chain, unit *NodeUnit) RequestMiddleware {
	mainRequester := unit.requester

	return func(request *http.Request) *http.Request {
		node.mutex.Lock()
		unit.hits += 1
		node.mutex.Unlock()

		defer func() {
			go func() {
				time.Sleep(unit.limit.duration)
				node.mutex.Lock()
				unit.hits -= 1
				node.mutex.Unlock()
			}()
		}()

		return mainRequester(request)
	}
}

func NewChain(
	initNodes []*NodeUnit,
	checkTickRate checkTick,
) *chain {
	createdBlockNode := &chain{
		mutex:         &sync.RWMutex{},
		checkTickRate: checkTickRate,
	}

	nodes := initNodes
	for _, node := range nodes {
		node.requester = createDecoratedRequest(createdBlockNode, node)
	}

	createdBlockNode.nodes = nodes

	return createdBlockNode
}

func (b *chain) getNodeRequester(apiClient apiCaller) Requester {
	node := b.findNodeWithTimeout()
	if node != nil {
		return func(request *http.Request) (*Response, error) {
			return apiClient.doRequest(context.Background(), node.requester(request))
		}
	}

	return nil
}

func (b *chain) findNodeWithTimeout() *NodeUnit {
	firstLoad := b.findNode()
	if firstLoad != nil {
		return firstLoad
	}

	ticker := time.NewTicker(b.checkTickRate.tickRate)
	foundNodeChannel := make(chan *NodeUnit)
	tickerDone := make(chan bool)
	go func() {
		for {
			select {
			case <-tickerDone:
				foundNodeChannel <- nil
				return
			case <-ticker.C:
				foundNode := b.findNode()
				if foundNode != nil {
					foundNodeChannel <- foundNode
					return
				}
			}
		}
	}()

	go func() {
		time.Sleep(b.checkTickRate.maxCheckDuration)
		ticker.Stop()
		tickerDone <- true
	}()

	foundNode := <-foundNodeChannel
	return foundNode
}

func (b *chain) findNode() *NodeUnit {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	sort.Slice(b.nodes, func(i, j int) bool {
		if b.nodes[i].priority == b.nodes[j].priority {
			return b.nodes[i].hits < b.nodes[j].hits
		}

		return b.nodes[i].priority > b.nodes[j].priority
	})
	for _, node := range b.nodes {
		if node.hits < node.limit.count {
			return node
		}
	}

	return nil
}
