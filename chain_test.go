package eznode

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestFindFreeNode(t *testing.T) {
	chainNode1 := NewChainNode(ChainNodeData{
		name:           "Node 1",
		url:            "http://example.com",
		limit:          NewChainNodeLimit(10, 2*time.Second),
		requestTimeout: 1 * time.Second,
		priority:       1,
		middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	createdChain := NewChain(
		[]*ChainNode{
			chainNode1,
		},
		NewCheckTick(1*time.Second, 5*time.Second),
	)

	foundNode := createdChain.getFreeNode()

	assert.NotNil(t, foundNode, "should find node")
	if foundNode != nil {
		assert.Equal(t, chainNode1.id, foundNode.id, "should found node id be correct")
	}
}

func TestNotFindNode(t *testing.T) {
	chainNode1 := NewChainNode(ChainNodeData{
		name:           "Node 1",
		url:            "http://example.com",
		limit:          NewChainNodeLimit(10, 2*time.Second),
		requestTimeout: 1 * time.Second,
		priority:       1,
		middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	createdChain := NewChain(
		[]*ChainNode{
			chainNode1,
		},
		NewCheckTick(500*time.Millisecond, 1*time.Second),
	)

	chainNode1.hits = 10

	foundNode := createdChain.getFreeNode()

	assert.Nil(t, foundNode, "should not find node")
}

func TestLoadBalance(t *testing.T) {
	chainNode1 := NewChainNode(ChainNodeData{
		name:           "Node 1",
		url:            "http://example.com",
		limit:          NewChainNodeLimit(10, 2*time.Second),
		requestTimeout: 1 * time.Second,
		priority:       1,
		middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	chainNode2 := NewChainNode(ChainNodeData{
		name:           "Node 2",
		url:            "http://example.com",
		limit:          NewChainNodeLimit(10, 2*time.Second),
		requestTimeout: 1 * time.Second,
		priority:       1,
		middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	createdChain := NewChain(
		[]*ChainNode{
			chainNode1,
			chainNode2,
		},
		NewCheckTick(500*time.Millisecond, 1*time.Second),
	)

	chainNode1.hits = 1
	foundNode := createdChain.getFreeNode()
	assert.Equal(t, chainNode2.id, foundNode.id, "should route to node 2")

	chainNode2.hits = 3
	chainNode1.hits = 2
	foundNode = createdChain.getFreeNode()
	assert.Equal(t, chainNode1.id, foundNode.id, "should route to node 1")
}
