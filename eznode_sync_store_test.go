package eznode

import (
	"github.com/amovah/eznode/stats"
	"github.com/amovah/eznode/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

type mockStorage struct {
	save func(storage.ErrorHandler, []stats.ChainStats)
	load func() ([]stats.ChainStats, error)
}

func (m *mockStorage) Save(handler storage.ErrorHandler, stats []stats.ChainStats) {
	m.save(handler, stats)
}

func (m *mockStorage) Load() ([]stats.ChainStats, error) {
	return m.load()
}

func TestLoadWithEmptyStorageShouldBeEmpty(t *testing.T) {
	mockedStorage := mockStorage{
		save: func(handler storage.ErrorHandler, chainStats []stats.ChainStats) {},
		load: func() ([]stats.ChainStats, error) {
			return []stats.ChainStats{}, nil
		},
	}

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

	ezNode, err := NewEzNode(
		[]*Chain{
			createdChain,
		},
		WithStore(
			&mockedStorage,
			time.Second,
			func(_ error) {},
		),
	)

	assert.Nil(t, err)
	assert.Equal(t, uint(0), ezNode.chains["test-chain"].nodes[0].hits)
	assert.Equal(t, uint64(0), ezNode.chains["test-chain"].nodes[0].totalHits)
}

func TestShouldLoadCorrectly(t *testing.T) {
	currentHit := uint(2)
	totalHits := uint64(141)

	mockedStorage := mockStorage{
		save: func(handler storage.ErrorHandler, chainStats []stats.ChainStats) {},
		load: func() ([]stats.ChainStats, error) {
			return []stats.ChainStats{
				{
					Id: "test-chain",
					Nodes: []stats.ChainNodeStats{
						{
							Name:        "Node 1",
							CurrentHits: currentHit,
							TotalHits:   totalHits,
							Limits:      0,
							Priority:    0,
						},
					},
				},
			}, nil
		},
	}

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

	ezNode, err := NewEzNode(
		[]*Chain{
			createdChain,
		},
		WithStore(
			&mockedStorage,
			time.Second,
			func(_ error) {},
		),
	)

	assert.Nil(t, err)
	assert.Equal(t, currentHit, ezNode.chains["test-chain"].nodes[0].hits)
	assert.Equal(t, totalHits, ezNode.chains["test-chain"].nodes[0].totalHits)
}

func TestStartStopSyncStore(t *testing.T) {
	saveCallCount := 0
	mockedStorage := mockStorage{
		save: func(handler storage.ErrorHandler, chainStats []stats.ChainStats) {
			saveCallCount += 1
		},
		load: func() ([]stats.ChainStats, error) {
			return []stats.ChainStats{}, nil
		},
	}

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

	ezNode, err := NewEzNode(
		[]*Chain{
			createdChain,
		},
		WithStore(
			&mockedStorage,
			500*time.Millisecond,
			func(_ error) {},
		),
	)
	assert.Nil(t, err)

	ezNode.StartSyncStore()
	time.Sleep(2000 * time.Millisecond)
	ezNode.StopSyncStore()
	assert.Equal(t, 4, saveCallCount)
	time.Sleep(2 * time.Second)
	assert.Equal(t, 4, saveCallCount)
}
