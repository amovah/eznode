package eznode

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

type EzNode struct {
	chains    map[string]*Chain
	apiCaller ApiCaller
}

func (e *EzNode) SendRequest(chainId string, request *http.Request) (*Response, error) {
	selectedChain := e.chains[chainId]
	if selectedChain == nil {
		return nil, errors.New(fmt.Sprintf("cannot find a chain with id %s", chainId))
	}

	excludeNodes := make(map[uuid.UUID]bool)
	errorTrace := make([]NodeErrorTrace, 0)
	return e.tryRequest(selectedChain, request, 0, excludeNodes, errorTrace)
}

func (e *EzNode) tryRequest(
	selectedChain *Chain,
	request *http.Request,
	tryCount int,
	excludeNodes map[uuid.UUID]bool,
	errorTrace []NodeErrorTrace,
) (*Response, error) {
	selectedNode := selectedChain.getFreeNode(excludeNodes)
	if selectedNode == nil {
		return nil, EzNodeError{
			Message: fmt.Sprintf("'%s' is at full capacity", selectedChain.id),
			Metadata: ChainResponseMetadata{
				ChainId:      selectedChain.id,
				RequestedUrl: request.URL.String(),
				Retry:        tryCount,
				ErrorTrace: append(errorTrace, NodeErrorTrace{
					Err: errors.New(fmt.Sprintf("'%s' is at full capacity", selectedChain.id)),
				}),
			},
		}
	}

	createdRequest := selectedNode.middleware(request.Clone(context.Background()))
	ctx, cancelTimeout := context.WithTimeout(context.Background(), selectedNode.requestTimeout)
	defer cancelTimeout()

	res, err := e.apiCaller.doRequest(ctx, createdRequest)
	if isResponseValid(selectedChain.failureStatusCodes, res, err) || tryCount >= selectedChain.retryCount {
		metadata := ChainResponseMetadata{
			ChainId:      selectedChain.id,
			RequestedUrl: request.URL.String(),
			Retry:        tryCount,
			ErrorTrace: append(errorTrace, NodeErrorTrace{
				NodeName: selectedNode.name,
				NodeId:   selectedNode.id,
				Err:      err,
			}),
		}

		if err != nil {
			return res, EzNodeError{
				Message:  err.Error(),
				Metadata: metadata,
			}
		}

		res.Metadata = metadata
		return res, nil
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
