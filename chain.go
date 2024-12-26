package eznode

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// RequestMiddleware is a function that is called before the request is processed.
type RequestMiddleware func(*http.Request) *http.Request

type Chain struct {
	id                 string
	nodes              []*ChainNode
	mutex              *sync.RWMutex
	checkTickRate      CheckTick
	failureStatusCodes map[int]bool
	retryCount         int
}

type NewChainConfig struct {
	// Id of the chain
	Id string
	// List of nodes in the chain
	Nodes []*ChainNode
	// tick rate for checking nodes availability
	CheckTickRate CheckTick
	// list of http status codes which recognized as failure
	FailureStatusCodes []int
	// number of retries for failed requests
	RetryCount int
}

// NewChain creates new Chain
// If FailureStatusCodes is not specified, default list of status codes is used
func NewChain(
	chainData NewChainConfig,
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

	seenName := make(map[string]bool)
	for _, node := range chainData.Nodes {
		if seenName[node.name] {
			log.Fatal("you cannot have duplicate name")
		}

		seenName[node.name] = true
	}

	failureStatusCodes := make(map[int]bool)
	if chainData.FailureStatusCodes != nil {
		for _, statusCode := range chainData.FailureStatusCodes {
			failureStatusCodes[statusCode] = true
		}
	} else {
		for _, statusCode := range DefaultFailureStatusCodes {
			failureStatusCodes[statusCode] = true
		}
	}

	return &Chain{
		id:                 chainData.Id,
		mutex:              &sync.RWMutex{},
		checkTickRate:      chainData.CheckTickRate,
		failureStatusCodes: failureStatusCodes,
		retryCount:         chainData.RetryCount,
		nodes:              chainData.Nodes,
	}
}

func (c *Chain) getFreeNode(excludeNodes map[string]bool, includeNodes map[string]bool) *ChainNode {
	if len(excludeNodes) == len(c.nodes) {
		return nil
	}

	firstLoadNode := c.findNode(excludeNodes, includeNodes)
	if firstLoadNode != nil {
		return firstLoadNode
	}

	deadlineToFind := time.After(c.checkTickRate.MaxCheckDuration)
	ticker := time.NewTicker(c.checkTickRate.TickRate)

	for {
		select {
		case <-deadlineToFind:
			return nil
		case <-ticker.C:
			foundNode := c.findNode(excludeNodes, includeNodes)
			if foundNode != nil {
				return foundNode
			}
		}
	}
}

func (c *Chain) findNode(excludeNodes map[string]bool, includeNodes map[string]bool) *ChainNode {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	selectedNode := &ChainNode{
		priority: -1,
	}

	for _, node := range c.nodes {
		if !excludeNodes[node.name] &&
			((len(includeNodes) > 0 && includeNodes[node.name]) || len(includeNodes) == 0) &&
			node.priority >= selectedNode.priority &&
			node.hits < node.limit.Count &&
			(node.hits <= selectedNode.hits || selectedNode.priority == -1) &&
			!node.disabled {
			selectedNode = node
		}
	}

	if selectedNode.priority == -1 {
		return nil
	}

	selectedNode.hits += 1
	return selectedNode
}
