package eznode

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFindFreeNode(t *testing.T) {
	t.Parallel()

	chainNode1 := NewChainNode(NewChainParam{
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
		NewChainParams{
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

	foundNode := createdChain.getFreeNode(make(map[string]bool), make(map[string]bool))

	assert.NotNil(t, foundNode, "should find node")
	if foundNode != nil {
		assert.Equal(t, chainNode1.name, foundNode.name, "should found node name be correct")
	}
}

func TestNotFindNode(t *testing.T) {
	t.Parallel()

	chainNode1 := NewChainNode(NewChainParam{
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
		NewChainParams{
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

	foundNode := createdChain.getFreeNode(make(map[string]bool), make(map[string]bool))

	assert.Nil(t, foundNode, "should not find node")
}

func TestDisableNode(t *testing.T) {
	t.Parallel()

	chainNode1 := NewChainNode(NewChainParam{
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
		NewChainParams{
			Id: "test-chain",
			Nodes: []*ChainNode{
				chainNode1,
			},
			CheckTickRate: CheckTick{
				TickRate:         100 * time.Millisecond,
				MaxCheckDuration: 200 * time.Millisecond,
			},
			RetryCount: 2,
		},
	)

	createdChain.disableNode("Node 1")
	foundNode := createdChain.getFreeNode(make(map[string]bool), make(map[string]bool))

	assert.Nil(t, foundNode, "should not find node")

	createdChain.enableNode("Node 1")
	foundNode = createdChain.getFreeNode(make(map[string]bool), make(map[string]bool))

	assert.NotNil(t, foundNode, "should not find node")
}

func TestDisableNodeWithTime(t *testing.T) {
	t.Parallel()

	chainNode1 := NewChainNode(NewChainParam{
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
		NewChainParams{
			Id: "test-chain",
			Nodes: []*ChainNode{
				chainNode1,
			},
			CheckTickRate: CheckTick{
				TickRate:         100 * time.Millisecond,
				MaxCheckDuration: 200 * time.Millisecond,
			},
			RetryCount: 2,
		},
	)

	createdChain.disableNodeWithTime("Node 1", 2*time.Second)
	foundNode := createdChain.getFreeNode(make(map[string]bool), make(map[string]bool))
	assert.Nil(t, foundNode, "should not find node")

	time.Sleep(1 * time.Second)
	foundNode = createdChain.getFreeNode(make(map[string]bool), make(map[string]bool))
	assert.Nil(t, foundNode, "should not find node")

	time.Sleep(1 * time.Second)
	foundNode = createdChain.getFreeNode(make(map[string]bool), make(map[string]bool))
	assert.NotNil(t, foundNode, "should find node")
}

func TestLoadBalance(t *testing.T) {
	t.Parallel()

	chainNode1 := NewChainNode(NewChainParam{
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

	chainNode2 := NewChainNode(NewChainParam{
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
		NewChainParams{
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
	foundNode := createdChain.getFreeNode(make(map[string]bool), make(map[string]bool))
	assert.Equal(t, chainNode2.name, foundNode.name, "should route to node 2")

	chainNode2.hits = 3
	chainNode1.hits = 2
	foundNode = createdChain.getFreeNode(make(map[string]bool), make(map[string]bool))
	assert.Equal(t, chainNode1.name, foundNode.name, "should route to node 1")
}

func TestShouldFilterNode(t *testing.T) {
	t.Parallel()

	chainNode1 := NewChainNode(NewChainParam{
		Name: "Node 1",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 3,
			Per:   10 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	chainNode2 := NewChainNode(NewChainParam{
		Name: "Node 2",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 10,
			Per:   10 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	createdChain := NewChain(
		NewChainParams{
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

	includeNodes := make(map[string]bool)
	includeNodes[chainNode1.name] = true

	foundNode := createdChain.getFreeNode(make(map[string]bool), includeNodes)
	assert.Equal(t, chainNode1.name, foundNode.name)
	foundNode = createdChain.getFreeNode(make(map[string]bool), includeNodes)
	assert.Equal(t, chainNode1.name, foundNode.name)
	foundNode = createdChain.getFreeNode(make(map[string]bool), includeNodes)
	assert.Equal(t, chainNode1.name, foundNode.name)

	foundNode = createdChain.getFreeNode(make(map[string]bool), includeNodes)
	assert.Nil(t, foundNode)
	foundNode = createdChain.getFreeNode(make(map[string]bool), includeNodes)
	assert.Nil(t, foundNode)
}
