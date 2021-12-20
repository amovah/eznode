package eznode

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestNewChain(t *testing.T) {
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

	assert.Equal(t, chainNode1.id, foundNode.id)
}
