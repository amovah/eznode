package eznode

import (
	"github.com/google/uuid"
	"log"
	"net/http"
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

func createMiddleware(chainNode *ChainNode) RequestMiddleware {
	mainMiddleware := chainNode.middleware

	return func(request *http.Request) *http.Request {
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

	nodes := chainData.Nodes
	for _, node := range nodes {
		node.middleware = createMiddleware(node)
	}

	return &Chain{
		id:                 chainData.Id,
		mutex:              &sync.RWMutex{},
		checkTickRate:      chainData.CheckTickRate,
		failureStatusCodes: failureStatusCodes,
		retryCount:         chainData.RetryCount,
		nodes:              nodes,
	}
}

func (c *Chain) getFreeNode(excludeNodes map[uuid.UUID]bool) *ChainNode {
	if len(excludeNodes) == len(c.nodes) {
		return nil
	}

	firstLoadNode := c.findNode(excludeNodes)
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
				foundNode := c.findNode(excludeNodes)
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

func (c *Chain) findNode(excludeNodes map[uuid.UUID]bool) *ChainNode {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	selectedNode := &ChainNode{
		priority: -1,
	}

	for _, node := range c.nodes {
		if !excludeNodes[node.id] && node.priority >= selectedNode.priority && node.hits < node.limit.Count {
			if node.hits <= selectedNode.hits || selectedNode.priority == -1 {
				selectedNode = node
			}
		}
	}

	if selectedNode.priority == -1 {
		return nil
	}

	selectedNode.hits += 1
	return selectedNode
}
