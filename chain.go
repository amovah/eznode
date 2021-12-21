package eznode

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestMiddleware func(*http.Request) *http.Request

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
	Id                 string
	Nodes              []*ChainNode
	CheckTickRate      CheckTick
	FailureStatusCodes []int
	RetryCount         int
}

func NewChain(
	chainData ChainData,
) *Chain {
	if chainData.Id == "" {
		log.Fatal("id cannot be empty")
	}

	if chainData.CheckTickRate.TickRate < 50*time.Millisecond {
		log.Fatal("tick rate cannot be less than 50 millisecond")
	}

	if chainData.CheckTickRate.MaxCheckDuration < chainData.CheckTickRate.TickRate {
		log.Fatal("max check duration must be greater than tick rate")
	}

	if chainData.RetryCount < 0 {
		log.Fatal("retry must be greater than -1")
	}

	failureStatusCodes := make(map[int]bool)
	for _, statusCode := range chainData.FailureStatusCodes {
		failureStatusCodes[statusCode] = true
	}

	createdChain := &Chain{
		id:                 chainData.Id,
		mutex:              &sync.RWMutex{},
		checkTickRate:      chainData.CheckTickRate,
		failureStatusCodes: failureStatusCodes,
		retryCount:         chainData.RetryCount,
	}

	nodes := chainData.Nodes
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
