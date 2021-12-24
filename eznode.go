package eznode

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type EzNode struct {
	chains      map[string]*Chain
	apiCaller   ApiCaller
	syncStorage syncStorage
}

func (e *EzNode) SendRequest(chainId string, request *http.Request) (*Response, error) {
	selectedChain := e.chains[chainId]
	if selectedChain == nil {
		return nil, errors.New(fmt.Sprintf("cannot find chain id %s", chainId))
	}

	excludeNodes := make(map[string]bool)
	errorTrace := make([]NodeTrace, 0)
	return e.tryRequest(selectedChain, request, 0, excludeNodes, errorTrace)
}

func (e *EzNode) tryRequest(
	selectedChain *Chain,
	request *http.Request,
	tryCount int,
	excludeNodes map[string]bool,
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

	res, err := e.apiCaller.doRequest(ctx, createdRequest)
	collectMetric(selectedNode, res, err)
	releaseResource(selectedChain, selectedNode)

	if isResponseValid(selectedChain.failureStatusCodes, res, err) {
		res.Metadata = ChainResponseMetadata{
			ChainId:      selectedChain.id,
			RequestedUrl: request.URL.String(),
			Retry:        tryCount,
			Trace: append(errorTrace, NodeTrace{
				NodeName:   selectedNode.name,
				StatusCode: res.StatusCode,
				Err:        nil,
			}),
		}
		return res, nil
	}

	if err != nil {
		errorTrace = append(errorTrace, NodeTrace{
			NodeName: selectedNode.name,
			Err:      err,
		})
	} else {
		errorTrace = append(errorTrace, NodeTrace{
			NodeName:   selectedNode.name,
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

	excludeNodes[selectedNode.name] = true
	return e.tryRequest(selectedChain, request, tryCount+1, excludeNodes, errorTrace)
}

func collectMetric(selectedNode *ChainNode, res *Response, err error) {
	go func() {
		atomic.AddUint64(&selectedNode.totalHits, 1)
		if err != nil {
			responseStats := selectedNode.responseStats[res.StatusCode]
			atomic.AddUint64(&responseStats, 1)
		} else {
			responsesStats := selectedNode.responseStats[0]
			atomic.AddUint64(&responsesStats, 1)
		}
	}()
}

func releaseResource(selectedChain *Chain, selectedNode *ChainNode) {
	go func() {
		time.Sleep(selectedNode.limit.Per)
		selectedChain.mutex.Lock()
		selectedNode.hits -= 1
		selectedChain.mutex.Unlock()
	}()
}

func isResponseValid(failureStatusCodes map[int]bool, res *Response, err error) bool {
	return err == nil && !(failureStatusCodes[res.StatusCode])
}
