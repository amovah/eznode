package eznode

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

type mockApiCall struct {
	returnFunc   func(*http.Request) (*Response, error)
	validateFunc func(*http.Request)
}

func (m mockApiCall) doRequest(context context.Context, request *http.Request) (*Response, error) {
	m.validateFunc(request)
	return m.returnFunc(request)
}

func TestCallRightRequest(t *testing.T) {
	mockedApiCall := mockApiCall{
		returnFunc: func(request *http.Request) (*Response, error) {
			return &Response{
				StatusCode: 0,
				Body:       nil,
				Headers:    &http.Header{},
			}, nil
		},
		validateFunc: func(request *http.Request) {
			assert.Equal(t, "http://example.com/", request.URL.String())
			assert.Equal(t, "GET", request.Method)
		},
	}

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

	ezNode := NewEzNode([]*Chain{createdChain}, WithApiClient(mockedApiCall))
	request, _ := http.NewRequest("GET", "/", nil)
	ezNode.SendRequest("test-chain", request)
}

func TestRetry(t *testing.T) {
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

	chainNode3 := NewChainNode(ChainNodeData{
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
		ChainData{
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

	chainNode3 := NewChainNode(ChainNodeData{
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
		ChainData{
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

	chainNode3 := NewChainNode(ChainNodeData{
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
		ChainData{
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

func TestReturnErrorType(t *testing.T) {
	mockedApiCall := mockApiCall{
		returnFunc: func(request *http.Request) (*Response, error) {
			return nil, errors.New("test error")
		},
		validateFunc: func(request *http.Request) {
			assert.Equal(t, "http://example.com/", request.URL.String())
			assert.Equal(t, "GET", request.Method)
		},
	}

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

	ezNode := NewEzNode([]*Chain{createdChain}, WithApiClient(mockedApiCall))
	request, _ := http.NewRequest("GET", "/", nil)
	_, err := ezNode.SendRequest("test-chain", request)
	_, ok := err.(EzNodeError)
	if !ok {
		assert.Fail(t, "error must be type of EzNodeError")
	}
}
