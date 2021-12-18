package eznode

import (
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

func createDecoratedRequest(node *chain, unit *NodeUnit) Requester {
	mainRequester := unit.requester

	return func(request *http.Request) (*Response, error) {
		node.mutex.Lock()
		unit.hits += 1
		node.mutex.Unlock()

		defer func() {
			go func() {
				time.Sleep(unit.limitDuration)
				node.mutex.Lock()
				unit.hits -= 1
				node.mutex.Unlock()
			}()
		}()

		return mainRequester(request)
	}
}

func NewBlockNode(
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

func (b *chain) GetNodeRequester() Requester {
	node := b.findNodeWithTimeout()
	if node != nil {
		return node.requester
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
		return b.nodes[i].hits < b.nodes[j].hits
	})
	for _, node := range b.nodes {
		if node.hits < node.limit {
			return node
		}
	}

	return nil
}
