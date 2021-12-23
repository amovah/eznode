package eznode

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"sync/atomic"
	"time"
)

type EzNode struct {
	chains    map[string]*Chain
	apiCaller ApiCaller
}

func (e *EzNode) SendRequest(chainId string, request *http.Request) (*Response, error) {
	selectedChain := e.chains[chainId]
	if selectedChain == nil {
		return nil, errors.New(fmt.Sprintf("cannot find chain id %s", chainId))
	}

	excludeNodes := make(map[uuid.UUID]bool)
	errorTrace := make([]NodeTrace, 0)
	return e.tryRequest(selectedChain, request, 0, excludeNodes, errorTrace)
}

func (e *EzNode) tryRequest(
	selectedChain *Chain,
	request *http.Request,
	tryCount int,
	excludeNodes map[uuid.UUID]bool,
	errorTrace []NodeTrace,
) (*Response, error) {
	selectedNode := selectedChain.getFreeNode(excludeNodes)
	if selectedNode == nil {
		errorMessage := fmt.Sprintf("'%s' chain is at full capacity", selectedChain.id)
		return nil, EzNodeError{
			Message: errorMessage,
			Metadata: ChainResponseMetadata{
				ChainId:      selectedChain.id,
				RequestedUrl: request.URL.String(),
				Retry:        tryCount,
				Trace: append(errorTrace, NodeTrace{
					StatusCode: http.StatusTooManyRequests,
					Err:        errors.New(errorMessage),
				}),
			},
		}
	}

	createdRequest := selectedNode.middleware(request.Clone(context.Background()))
	ctx, cancelTimeout := context.WithTimeout(context.Background(), selectedNode.requestTimeout)
	defer cancelTimeout()

	go func() {
		atomic.AddUint64(&selectedNode.totalHits, 1)
	}()
	res, err := e.apiCaller.doRequest(ctx, createdRequest)
	go func() {
		time.Sleep(selectedNode.limit.Per)
		selectedChain.mutex.Lock()
		selectedNode.hits -= 1
		selectedChain.mutex.Unlock()
	}()

	if isResponseValid(selectedChain.failureStatusCodes, res, err) {
		res.Metadata = ChainResponseMetadata{
			ChainId:      selectedChain.id,
			RequestedUrl: request.URL.String(),
			Retry:        tryCount,
			Trace: append(errorTrace, NodeTrace{
				NodeName:   selectedNode.name,
				NodeId:     selectedNode.id,
				StatusCode: res.StatusCode,
				Err:        nil,
			}),
		}
		return res, nil
	}

	if err != nil {
		errorTrace = append(errorTrace, NodeTrace{
			NodeName: selectedNode.name,
			NodeId:   selectedNode.id,
			Err:      err,
		})
	} else {
		errorTrace = append(errorTrace, NodeTrace{
			NodeName:   selectedNode.name,
			NodeId:     selectedNode.id,
			StatusCode: res.StatusCode,
			Err:        errors.New(fmt.Sprintf("request failed with status code %v", res.StatusCode)),
		})
	}

	if tryCount >= selectedChain.retryCount {
		httpStatusCode := http.StatusFailedDependency
		errorMessage := "reached max retries, " + http.StatusText(httpStatusCode)

		return &Response{}, EzNodeError{
			Message: errorMessage,
			Metadata: ChainResponseMetadata{
				ChainId:      selectedChain.id,
				RequestedUrl: request.URL.String(),
				Retry:        tryCount,
				Trace: append(errorTrace, NodeTrace{
					StatusCode: httpStatusCode,
					Err:        errors.New(errorMessage),
				}),
			},
		}
	}

	excludeNodes[selectedNode.id] = true
	return e.tryRequest(selectedChain, request, tryCount+1, excludeNodes, errorTrace)
}

func isResponseValid(failureStatusCodes map[int]bool, res *Response, err error) bool {
	return err == nil && !(failureStatusCodes[res.StatusCode])
}

type Option func(*EzNode)

func NewEzNode(chains []*Chain, options ...Option) *EzNode {
	chainHashMap := make(map[string]*Chain)
	for _, userChain := range chains {
		chainHashMap[userChain.id] = userChain
	}

	ezNode := &EzNode{
		chains: chainHashMap,
		apiCaller: &apiCallerClient{
			client: createHttpClient(),
		},
	}

	for _, option := range options {
		option(ezNode)
	}

	return ezNode
}

func WithApiClient(apiCaller ApiCaller) Option {
	return func(ezNode *EzNode) {
		ezNode.apiCaller = apiCaller
	}
}
