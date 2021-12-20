package eznode

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestMiddleware func(*http.Request) *http.Request
type Requester func(request *http.Request) (*Response, error)

type Chain struct {
	id                 string
	nodes              []*ChainNode
	mutex              *sync.RWMutex
	checkTickRate      CheckTick
	failureStatusCodes map[int]bool
	retryCount         int
}

func createMiddleware(node *Chain, unit *ChainNode) RequestMiddleware {
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

type ChainData struct {
	id                     string
	initNodes              []*ChainNode
	checkTickRate          CheckTick
	initFailureStatusCodes []int
	retry                  int
}

func NewChain(
	chainData ChainData,
) *Chain {
	if chainData.id == "" {
		log.Fatal("id cannot be empty")
	}

	if chainData.checkTickRate.TickRate < 50*time.Millisecond {
		log.Fatal("tick rate cannot be less than 50 millisecond")
	}

	if chainData.checkTickRate.MaxCheckDuration < chainData.checkTickRate.TickRate {
		log.Fatal("max check duration must be greater than tick rate")
	}

	if chainData.retry < 0 {
		log.Fatal("retry must be greater than -1")
	}

	failureStatusCodes := make(map[int]bool)
	for _, statusCode := range chainData.initFailureStatusCodes {
		failureStatusCodes[statusCode] = true
	}

	createdChain := &Chain{
		id:            chainData.id,
		mutex:         &sync.RWMutex{},
		checkTickRate: chainData.checkTickRate,
	}

	nodes := chainData.initNodes
	for _, node := range nodes {
		node.middleware = createMiddleware(createdChain, node)
	}

	createdChain.nodes = nodes

	return createdChain
}

func (c *Chain) getFreeNode() *ChainNode {
	firstLoadNode := c.findNode()
	if firstLoadNode != nil {
		return firstLoadNode
	}

	ticker := time.NewTicker(c.checkTickRate.TickRate)
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
		time.Sleep(c.checkTickRate.MaxCheckDuration)
		ticker.Stop()
		tickerDone <- true
	}()

	foundNode := <-foundNodeChannel
	return foundNode
}

func (c *Chain) findNode() *ChainNode {
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
