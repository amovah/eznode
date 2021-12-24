package eznode

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestLoadWithEmptyStorageShouldBeEmpty(t *testing.T) {
	chainNode1 := NewChainNode(ChainNodeData{
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
		ChainData{
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
}

func TestShouldLoadCorrectly(t *testing.T) {
	currentHit := uint(2)
	totalHits := uint64(141)

	chainNode1 := NewChainNode(ChainNodeData{
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
		ChainData{
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
					Limits:      0,
					Priority:    0,
				},
			},
		},
	})

	assert.Equal(t, currentHit, ezNode.chains["test-chain"].nodes[0].hits)
	assert.Equal(t, totalHits, ezNode.chains["test-chain"].nodes[0].totalHits)
}

func TestStartStopSyncStore(t *testing.T) {
	saveCallCount := 0

	chainNode1 := NewChainNode(ChainNodeData{
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
		ChainData{
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
		WithSyncInterval(
			500*time.Millisecond,
		),
	)

	ezNode.StartSyncStorage(func(chainStats []ChainStats) {
		saveCallCount += 1
	})
	time.Sleep(2000 * time.Millisecond)
	ezNode.StopSyncStorage()
	assert.Equal(t, 4, saveCallCount)
	time.Sleep(2 * time.Second)
	assert.Equal(t, 4, saveCallCount)
}