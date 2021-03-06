package eznode

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadWithEmptyStorageShouldBeEmpty(t *testing.T) {
	t.Parallel()

	chainNode1 := NewChainNode(NewChainNodeConfig{
		Name: "Node 1",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 1,
			Per:   2 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
	})

	createdChain := NewChain(
		NewChainConfig{
			Id: "test-chain",
			Nodes: []*ChainNode{
				chainNode1,
			},
			CheckTickRate: CheckTick{
				TickRate:         100 * time.Millisecond,
				MaxCheckDuration: 200 * time.Millisecond,
			},
			FailureStatusCodes: []int{},
			RetryCount:         2,
		},
	)

	ezNode := NewEzNode(
		[]*Chain{
			createdChain,
		},
	)
	ezNode.LoadStats([]ChainStats{})

	assert.Equal(t, uint(0), ezNode.chains["test-chain"].nodes[0].hits)
	assert.Equal(t, uint64(0), ezNode.chains["test-chain"].nodes[0].totalHits)
	assert.Equal(t, 0, len(ezNode.chains["test-chain"].nodes[0].responseStats))
}

func TestShouldLoadCorrectly(t *testing.T) {
	t.Parallel()

	currentHit := uint(2)
	totalHits := uint64(141)

	chainNode1 := NewChainNode(NewChainNodeConfig{
		Name: "Node 1",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 1,
			Per:   2 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	createdChain := NewChain(
		NewChainConfig{
			Id: "test-chain",
			Nodes: []*ChainNode{
				chainNode1,
			},
			CheckTickRate: CheckTick{
				TickRate:         100 * time.Millisecond,
				MaxCheckDuration: 200 * time.Millisecond,
			},
			FailureStatusCodes: []int{},
			RetryCount:         2,
		},
	)

	ezNode := NewEzNode(
		[]*Chain{
			createdChain,
		},
	)
	ezNode.LoadStats([]ChainStats{
		{
			Id: "test-chain",
			Nodes: []ChainNodeStats{
				{
					Name:        "Node 1",
					CurrentHits: currentHit,
					TotalHits:   totalHits,
					ResponseStats: map[int]uint64{
						0:   10,
						200: 5,
					},
					Limits:   0,
					Priority: 0,
				},
			},
		},
	})

	assert.Equal(t, uint(0), ezNode.chains["test-chain"].nodes[0].hits)
	assert.Equal(t, totalHits, ezNode.chains["test-chain"].nodes[0].totalHits)
	assert.Equal(t, uint64(10), ezNode.chains["test-chain"].nodes[0].responseStats[0])
	assert.Equal(t, uint64(5), ezNode.chains["test-chain"].nodes[0].responseStats[200])
}
