package eznode

import (
	"net/http"
	"testing"
	"time"
)

func TestNewChain(t *testing.T) {
	chain := NewChain(
		[]*ChainNode{
			newChainNode(
				"http://example.com",
				NodeUnitLimit{
					count: 25,
					per:   5 * time.Second,
				},
				2*time.Second,
				2,
				func(request *http.Request) *http.Request {
					return request
				},
			),
		},
		CheckTick{
			tickRate:         1 * time.Second,
			maxCheckDuration: 5 * time.Second,
		},
	)

	chain.getFreeNode()
}
