package eznode

import (
	"context"
	"errors"
	"fmt"
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

	return e.tryRequest(selectedChain, request, 0)
}

func (e *EzNode) tryRequest(selectedChain *Chain, request *http.Request, tryCount int) (*Response, error) {
	selectedNode := selectedChain.getFreeNode()
	if selectedNode == nil {
		return nil, errors.New(fmt.Sprintf("chain id %s is at full capacity", selectedChain.id))
	}

	createdRequest := selectedNode.middleware(request.Clone(context.Background()))
	ctx, cancelTimeout := context.WithTimeout(context.Background(), selectedNode.requestTimeout)
	defer cancelTimeout()

	res, err := e.apiCaller.doRequest(ctx, createdRequest)
	if isResponseValid(selectedChain.failureStatusCodes, res, err) {
		return res, err
	}

	if tryCount >= selectedChain.retryCount {
		return res, err
	}

	return e.tryRequest(selectedChain, request, tryCount+1)
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
