package eznode

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockApiCall struct {
	returnFunc   func(*http.Request) (*Response, error)
	validateFunc func(*http.Request)
}

func (m mockApiCall) DoRequest(ctx context.Context, request *http.Request) (*Response, error) {
	m.validateFunc(request)
	return m.returnFunc(request)
}

func TestCallRightRequest(t *testing.T) {
	t.Parallel()

	mockedApiCall := mockApiCall{
		returnFunc: func(request *http.Request) (*Response, error) {
			return &Response{
				StatusCode: 200,
				Body:       nil,
				Headers:    &http.Header{},
			}, nil
		},
		validateFunc: func(request *http.Request) {
			assert.Equal(t, "http://example.com/", request.URL.String())
			assert.Equal(t, "GET", request.Method)
		},
	}

	chainNode1 := NewChainNode(NewChainNodeConfig{
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
		NewChainConfig{
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

	ezNode := NewEzNode([]*Chain{createdChain}, WithApiClient(mockedApiCall))
	request, _ := http.NewRequest("GET", "/", nil)
	ezNode.SendRequest("test-chain", request)
}

func TestRetry(t *testing.T) {
	t.Parallel()

	retryCount := 0
	chainMaxTry := 2
	seenUrls := make(map[string]bool)

	mockedApiCall := mockApiCall{
		returnFunc: func(request *http.Request) (*Response, error) {
			retryCount += 1
			return &Response{}, errors.New("error")
		},
		validateFunc: func(request *http.Request) {
			if seenUrls[request.URL.String()] {
				assert.Fail(t, "url seen already")
			}
		},
	}

	chainNode1 := NewChainNode(NewChainNodeConfig{
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

	chainNode2 := NewChainNode(NewChainNodeConfig{
		Name: "Node 2",
		Url:  "http://example2.com",
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

	chainNode3 := NewChainNode(NewChainNodeConfig{
		Name: "Node 3",
		Url:  "http://example3.com",
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
		NewChainConfig{
			Id: "test-chain",
			Nodes: []*ChainNode{
				chainNode1,
				chainNode2,
				chainNode3,
			},
			CheckTickRate: CheckTick{
				TickRate:         500 * time.Millisecond,
				MaxCheckDuration: 1 * time.Second,
			},
			FailureStatusCodes: []int{},
			RetryCount:         chainMaxTry,
		},
	)

	ezNode := NewEzNode([]*Chain{createdChain}, WithApiClient(mockedApiCall))
	request, _ := http.NewRequest("GET", "/", nil)
	ezNode.SendRequest("test-chain", request)
	assert.Equal(t, chainMaxTry+1, retryCount)
}

func TestFailOnFailureStatusCodes(t *testing.T) {
	t.Parallel()

	retryCount := 0
	chainMaxTry := 2

	mockedApiCall := mockApiCall{
		returnFunc: func(request *http.Request) (*Response, error) {
			retryCount += 1
			return &Response{
				StatusCode: http.StatusNotFound,
				Body:       nil,
				Headers:    nil,
			}, nil
		},
		validateFunc: func(request *http.Request) {
		},
	}

	chainNode1 := NewChainNode(NewChainNodeConfig{
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

	chainNode2 := NewChainNode(NewChainNodeConfig{
		Name: "Node 2",
		Url:  "http://example2.com",
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

	chainNode3 := NewChainNode(NewChainNodeConfig{
		Name: "Node 3",
		Url:  "http://example3.com",
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
		NewChainConfig{
			Id: "test-chain",
			Nodes: []*ChainNode{
				chainNode1,
				chainNode2,
				chainNode3,
			},
			CheckTickRate: CheckTick{
				TickRate:         500 * time.Millisecond,
				MaxCheckDuration: 1 * time.Second,
			},
			FailureStatusCodes: []int{http.StatusNotFound},
			RetryCount:         chainMaxTry,
		},
	)

	ezNode := NewEzNode([]*Chain{createdChain}, WithApiClient(mockedApiCall))
	request, _ := http.NewRequest("GET", "/", nil)
	ezNode.SendRequest("test-chain", request)
	assert.Equal(t, chainMaxTry+1, retryCount)
}

func TestNoMoreTryWhenCheckedAll(t *testing.T) {
	t.Parallel()

	retryCount := 0
	chainMaxTry := 5

	mockedApiCall := mockApiCall{
		returnFunc: func(request *http.Request) (*Response, error) {
			retryCount += 1
			return &Response{
				StatusCode: http.StatusNotFound,
				Body:       nil,
				Headers:    nil,
			}, nil
		},
		validateFunc: func(request *http.Request) {
		},
	}

	chainNode1 := NewChainNode(NewChainNodeConfig{
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

	chainNode2 := NewChainNode(NewChainNodeConfig{
		Name: "Node 2",
		Url:  "http://example2.com",
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

	chainNode3 := NewChainNode(NewChainNodeConfig{
		Name: "Node 3",
		Url:  "http://example3.com",
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
		NewChainConfig{
			Id: "test-chain",
			Nodes: []*ChainNode{
				chainNode1,
				chainNode2,
				chainNode3,
			},
			CheckTickRate: CheckTick{
				TickRate:         500 * time.Millisecond,
				MaxCheckDuration: 1 * time.Second,
			},
			FailureStatusCodes: []int{http.StatusNotFound},
			RetryCount:         chainMaxTry,
		},
	)

	ezNode := NewEzNode([]*Chain{createdChain}, WithApiClient(mockedApiCall))
	request, _ := http.NewRequest("GET", "/", nil)
	ezNode.SendRequest("test-chain", request)
	assert.Equal(t, 3, retryCount)
}

func TestLockAndReleaseResource(t *testing.T) {
	t.Parallel()

	mockedApiCall := mockApiCall{
		returnFunc: func(request *http.Request) (*Response, error) {
			return &Response{
				StatusCode: 200,
				Body:       nil,
				Headers:    &http.Header{},
			}, nil
		},
		validateFunc: func(request *http.Request) {
		},
	}

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

	ezNode := NewEzNode([]*Chain{createdChain}, WithApiClient(mockedApiCall))

	request, _ := http.NewRequest("GET", "/", nil)
	res, err := ezNode.SendRequest("test-chain", request)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, uint(1), ezNode.chains["test-chain"].nodes[0].hits)

	res, err = ezNode.SendRequest("test-chain", request)
	assert.NotNil(t, err)
	assert.Equal(t, uint(1), ezNode.chains["test-chain"].nodes[0].hits)

	time.Sleep(2 * time.Second)
	assert.Equal(t, uint(0), ezNode.chains["test-chain"].nodes[0].hits)
	res, err = ezNode.SendRequest("test-chain", request)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, uint(1), ezNode.chains["test-chain"].nodes[0].hits)
}

func TestConcurrentRequests(t *testing.T) {
	t.Parallel()

	mockedApiCall := mockApiCall{
		returnFunc: func(request *http.Request) (*Response, error) {
			time.Sleep(50 * time.Millisecond)
			return &Response{
				StatusCode: 200,
				Body:       nil,
				Headers:    &http.Header{},
			}, nil
		},
		validateFunc: func(request *http.Request) {
		},
	}

	chainNode1 := NewChainNode(NewChainNodeConfig{
		Name: "Node 1",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 100,
			Per:   1 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	chainNode2 := NewChainNode(NewChainNodeConfig{
		Name: "Node 2",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 25,
			Per:   1 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       1,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	chainNode3 := NewChainNode(NewChainNodeConfig{
		Name: "Node 3",
		Url:  "http://example.com",
		Limit: ChainNodeLimit{
			Count: 60,
			Per:   60 * time.Second,
		},
		RequestTimeout: 1 * time.Second,
		Priority:       0,
		Middleware: func(request *http.Request) *http.Request {
			return request
		},
	})

	createdChain := NewChain(
		NewChainConfig{
			Id: "test-chain",
			Nodes: []*ChainNode{
				chainNode1,
				chainNode3,
				chainNode2,
			},
			CheckTickRate: CheckTick{
				TickRate:         100 * time.Millisecond,
				MaxCheckDuration: 5 * time.Second,
			},
			FailureStatusCodes: []int{},
			RetryCount:         2,
		},
	)

	ezNode := NewEzNode([]*Chain{createdChain}, WithApiClient(mockedApiCall))
	request, _ := http.NewRequest("GET", "/", nil)

	w := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		go func() {
			w.Add(1)
			ezNode.SendRequest("test-chain", request)
			w.Done()
		}()
	}
	w.Wait()

	for i := 0; i < 100; i++ {
		go func() {
			w.Add(1)
			ezNode.SendRequest("test-chain", request)
			w.Done()
		}()
	}

	for _, node := range ezNode.chains["test-chain"].nodes {
		assert.True(t, node.hits <= node.limit.Count)
	}
	w.Wait()

	for _, node := range ezNode.chains["test-chain"].nodes {
		assert.True(t, node.hits <= node.limit.Count)
	}
}
