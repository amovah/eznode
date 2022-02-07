package eznode

import (
	"context"
	"errors"
	"fmt"
	"net"
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
					Time:       time.Now(),
					StatusCode: http.StatusTooManyRequests,
					Err:        errors.New(errorMessage),
				}),
			},
		}
	}

	createdRequest := selectedNode.middleware(request.Clone(context.Background()))
	ctx, cancelTimeout := context.WithTimeout(context.Background(), selectedNode.requestTimeout)
	defer cancelTimeout()

	res, err := e.apiCaller.DoRequest(ctx, createdRequest)
	go collectMetric(
		selectedNode,
		res,
		err,
		isResponseValid(selectedChain.failureStatusCodes, res, err),
	)
	go releaseResource(selectedChain, selectedNode)

	if isResponseValid(selectedChain.failureStatusCodes, res, err) {
		res.Metadata = ChainResponseMetadata{
			ChainId:      selectedChain.id,
			RequestedUrl: request.URL.String(),
			Retry:        tryCount,
			Trace: append(errorTrace, NodeTrace{
				Time:       time.Now(),
				NodeName:   selectedNode.name,
				StatusCode: res.StatusCode,
				Err:        nil,
			}),
		}
		return res, nil
	}

	nodeTrace := NodeTrace{
		Time:     time.Now(),
		NodeName: selectedNode.name,
	}
	if err != nil {
		nodeTrace.Err = err
		netError, ok := err.(net.Error)
		if errors.Is(err, context.DeadlineExceeded) || (ok && netError.Timeout()) {
			nodeTrace.StatusCode = http.StatusRequestTimeout
			errorTrace = append(errorTrace, nodeTrace)
		} else {
			nodeTrace.StatusCode = 0
			errorTrace = append(errorTrace, nodeTrace)
		}
	} else {
		nodeTrace.StatusCode = res.StatusCode
		nodeTrace.Err = errors.New(fmt.Sprintf("request failed with status code %v", res.StatusCode))
		errorTrace = append(errorTrace, nodeTrace)
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

func collectMetric(
	selectedNode *ChainNode,
	res *Response,
	err error,
	isValid bool,
) {
	atomic.AddUint64(&selectedNode.totalHits, 1)

	selectedNode.statsMutex.Lock()
	defer selectedNode.statsMutex.Unlock()
	if err != nil {
		netError, ok := err.(net.Error)
		if errors.Is(err, context.DeadlineExceeded) || (ok && netError.Timeout()) {
			selectedNode.responseStats[http.StatusRequestTimeout] += 1
		} else {
			selectedNode.responseStats[0] += 1
		}
	} else {
		selectedNode.responseStats[res.StatusCode] += 1
	}

	if !isValid {
		selectedNode.fails += 1
	}
}

func releaseResource(selectedChain *Chain, selectedNode *ChainNode) {
	time.Sleep(selectedNode.limit.Per)
	selectedChain.mutex.Lock()
	selectedNode.hits -= 1
	selectedChain.mutex.Unlock()
}

func isResponseValid(failureStatusCodes map[int]bool, res *Response, err error) bool {
	return err == nil && !(failureStatusCodes[res.StatusCode])
}
