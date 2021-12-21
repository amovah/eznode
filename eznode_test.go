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
	t            *testing.T
	returnFunc   func(*http.Request) (*Response, error)
	validateFunc func(*http.Request)
}

func (m mockApiCall) doRequest(context context.Context, request *http.Request) (*Response, error) {
	m.validateFunc(request)
	return m.returnFunc(request)
}

func TestCallRightRequest(t *testing.T) {
	mockedApiCall := mockApiCall{
		t: t,
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
		Name:           "Node 1",
		Url:            "http://example.com",
		Limit:          NewChainNodeLimit(10, 2*time.Second),
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

func TestRetryMode(t *testing.T) {
	retryCount := 0
	chainMaxTry := 5

	mockedApiCall := mockApiCall{
		t: t,
		returnFunc: func(request *http.Request) (*Response, error) {
			retryCount += 1
			return &Response{}, errors.New("error")
		},
		validateFunc: func(request *http.Request) {
		},
	}

	chainNode1 := NewChainNode(ChainNodeData{
		Name:           "Node 1",
		Url:            "http://example.com",
		Limit:          NewChainNodeLimit(10, 2*time.Second),
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
	chainMaxTry := 5

	mockedApiCall := mockApiCall{
		t: t,
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
		Name:           "Node 1",
		Url:            "http://example.com",
		Limit:          NewChainNodeLimit(10, 2*time.Second),
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