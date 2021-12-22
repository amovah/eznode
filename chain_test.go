package eznode

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestFindFreeNode(t *testing.T) {
	chainNode1 := NewChainNode(ChainNodeData{
		Name: "Node 1",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 10,
			Per:   2 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	createdChain := NewChain(
		ChainData{
			Id: "test chain",
			Nodes: []*ChainNode{
				chainNode1,
			},
			CheckTickRate: CheckTick{
				TickRate:         100 * time.Millisecond,
				MaxCheckDuration: 1 * time.Second,
			},
			FailureStatusCodes: []int{},
			RetryCount:         2,
		},
	)

	foundNode := createdChain.getFreeNode(make(map[uuid.UUID]bool))

	assert.NotNil(t, foundNode, "should find node")
	if foundNode != nil {
		assert.Equal(t, chainNode1.id, foundNode.id, "should found node id be correct")
	}
}

func TestNotFindNode(t *testing.T) {
	chainNode1 := NewChainNode(ChainNodeData{
		Name: "Node 1",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 10,
			Per:   2 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	createdChain := NewChain(
		ChainData{
			Id: "test-chain",
			Nodes: []*ChainNode{
				chainNode1,
			},
			CheckTickRate: CheckTick{
				TickRate:         100 * time.Millisecond,
				MaxCheckDuration: 1 * time.Second,
			},
			FailureStatusCodes: []int{},
			RetryCount:         2,
		},
	)

	chainNode1.hits = 10

	foundNode := createdChain.getFreeNode(make(map[uuid.UUID]bool))

	assert.Nil(t, foundNode, "should not find node")
}

func TestLoadBalance(t *testing.T) {
	chainNode1 := NewChainNode(ChainNodeData{
		Name: "Node 1",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 10,
			Per:   2 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	chainNode2 := NewChainNode(ChainNodeData{
		Name: "Node 2",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 10,
			Per:   2 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	createdChain := NewChain(
		ChainData{
			Id: "test chain",
			Nodes: []*ChainNode{
				chainNode1,
				chainNode2,
			},
			CheckTickRate: CheckTick{
				TickRate:         100 * time.Millisecond,
				MaxCheckDuration: 1 * time.Second,
			},
			FailureStatusCodes: []int{},
			RetryCount:         2,
		},
	)

	chainNode1.hits = 1
	foundNode := createdChain.getFreeNode(make(map[uuid.UUID]bool))
	assert.Equal(t, chainNode2.id, foundNode.id, "should route to node 2")

	chainNode2.hits = 3
	chainNode1.hits = 2
	foundNode = createdChain.getFreeNode(make(map[uuid.UUID]bool))
	assert.Equal(t, chainNode1.id, foundNode.id, "should route to node 1")
}
