package eznode

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

func generateTrace(nodeName string, err error, resStatus int) NodeTrace {
	nodeTrace := NodeTrace{
		Time:     time.Now(),
		NodeName: nodeName,
	}

	if err != nil {
		nodeTrace.Err = err
		netError, ok := err.(net.Error)
		if errors.Is(err, context.DeadlineExceeded) || (ok && netError.Timeout()) {
			nodeTrace.StatusCode = http.StatusRequestTimeout
			return nodeTrace
		}

		nodeTrace.StatusCode = 0
		return nodeTrace
	}

	nodeTrace.StatusCode = resStatus
	nodeTrace.Err = errors.New(fmt.Sprintf("request failed with status code %v", resStatus))
	return nodeTrace
}

// SendRequest send your request to specific chain
// If chain not found, return error
// Note: make sure which your request should not have host, schema and port
func (e *EzNode) SendRequest(ctx context.Context, chainId string, request *http.Request) (*Response, error) {
	return e.SendRequestSpecific(ctx, chainId, request, []string{})
}

// SendRequestSpecific send your request with specifying which node you want to use
// if your node rely on specific node (usually node which has more history) you can use
// this function to ensure your request will be responded by this node
func (e *EzNode) SendRequestSpecific(ctx context.Context, chainId string, request *http.Request, includeNodeList []string) (*Response, error) {
	selectedChain := e.chains[chainId]
	if selectedChain == nil {
		return nil, errors.New(fmt.Sprintf("cannot find chain id %s", chainId))
	}

	includeNodes := make(map[string]bool)
	for _, includeNode := range includeNodeList {
		includeNodes[includeNode] = true
	}

	excludeNodes := make(map[string]bool)
	nodeTrace := make([]NodeTrace, 0)

	var reqBody []byte
	var err error
	if request.Body != nil {
		reqBody, err = io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
	}

	tryCount := 0
	for tryCount < selectedChain.retryCount {
		selectedNode := selectedChain.getFreeNode(excludeNodes, includeNodes)
		if selectedNode == nil {
			errorMessage := fmt.Sprintf("'%s' chain is at full capacity", selectedChain.id)
			return nil, EzNodeError{
				Message: errorMessage,
				Metadata: ChainResponseMetadata{
					ChainId:      selectedChain.id,
					RequestedUrl: request.URL.String(),
					Retry:        tryCount,
					Trace: append(nodeTrace, NodeTrace{
						Time:       time.Now(),
						StatusCode: http.StatusTooManyRequests,
						Err:        errors.New(errorMessage),
					}),
				},
			}
		}

		clonedReq := request.Clone(context.Background())
		clonedReq.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		clonedReq = selectedNode.middleware(clonedReq)
		ctxTimeout, cancelTimeout := context.WithTimeout(ctx, selectedNode.requestTimeout)
		defer cancelTimeout()

		res, err := e.apiCaller.DoRequest(ctxTimeout, clonedReq)
		isValid := isResponseValid(selectedChain.failureStatusCodes, res, err)
		go releaseResource(selectedChain, selectedNode)
		go collectMetric(selectedNode, res, err, isValid)
		if isValid {
			res.Metadata = ChainResponseMetadata{
				ChainId:      selectedChain.id,
				RequestedUrl: request.URL.String(),
				Retry:        tryCount,
				Trace: append(nodeTrace, NodeTrace{
					Time:       time.Now(),
					NodeName:   selectedNode.name,
					StatusCode: res.StatusCode,
					Err:        nil,
				}),
			}
			return res, nil
		}

		resStatusCode := 0
		if res != nil {
			resStatusCode = res.StatusCode
		}

		nodeTrace = append(nodeTrace, generateTrace(selectedNode.name, err, resStatusCode))
		excludeNodes[selectedNode.name] = true
	}

	httpStatusCode := http.StatusFailedDependency
	errorMessage := "reached max retries, " + http.StatusText(httpStatusCode)

	return nil, EzNodeError{
		Message: errorMessage,
		Metadata: ChainResponseMetadata{
			ChainId:      selectedChain.id,
			RequestedUrl: request.URL.String(),
			Retry:        tryCount,
			Trace: append(nodeTrace, NodeTrace{
				StatusCode: httpStatusCode,
				Err:        errors.New(errorMessage),
			}),
		},
	}
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
