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
	nodes         []*ChainNode
	mutex         *sync.RWMutex
	checkTickRate checkTick
}

func createMiddleware(node *chain, unit *ChainNode) RequestMiddleware {
	mainMiddleware := unit.middleware

	return func(request *http.Request) *http.Request {
		node.mutex.Lock()
		unit.hits += 1
		node.mutex.Unlock()

		defer func() {
			go func() {
				time.Sleep(unit.limit.per)
				node.mutex.Lock()
				unit.hits -= 1
				node.mutex.Unlock()
			}()
		}()

		return mainMiddleware(request)
	}
}

func NewChain(
	initNodes []*ChainNode,
	checkTickRate checkTick,
) *chain {
	createdBlockNode := &chain{
		mutex:         &sync.RWMutex{},
		checkTickRate: checkTickRate,
	}

	nodes := initNodes
	for _, node := range nodes {
		node.middleware = createMiddleware(createdBlockNode, node)
	}

	createdBlockNode.nodes = nodes

	return createdBlockNode
}

func (c *chain) getNodeRequester() *ChainNode {
	return c.findNodeWithTimeout()
}

func (c *chain) findNodeWithTimeout() *ChainNode {
	firstLoadNode := c.findNode()
	if firstLoadNode != nil {
		return firstLoadNode
	}

	ticker := time.NewTicker(c.checkTickRate.tickRate)
	foundNodeChannel := make(chan *ChainNode)
	tickerDone := make(chan bool)
	go func() {
		for {
			select {
			case <-tickerDone:
				foundNodeChannel <- nil
				return
			case <-ticker.C:
				foundNode := c.findNode()
				if foundNode != nil {
					foundNodeChannel <- foundNode
					return
				}
			}
		}
	}()

	go func() {
		time.Sleep(c.checkTickRate.maxCheckDuration)
		ticker.Stop()
		tickerDone <- true
	}()

	foundNode := <-foundNodeChannel
	return foundNode
}

func (c *chain) findNode() *ChainNode {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	sort.Slice(c.nodes, func(i, j int) bool {
		if c.nodes[i].priority == c.nodes[j].priority {
			return c.nodes[i].hits < c.nodes[j].hits
		}

		return c.nodes[i].priority > c.nodes[j].priority
	})
	for _, node := range c.nodes {
		if node.hits < node.limit.count {
			return node
		}
	}

	return nil
}
