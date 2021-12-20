package eznode

import (
	"net/http"
)

type EzNode struct {
	chains    []*chain
	apiCaller apiCaller
}

type Option func(*EzNode)

func NewEzNode(chains []*chain, options ...Option) *EzNode {
	ezNode := &EzNode{
		chains: chains,
		apiCaller: &apiCallerClient{
			client: createHttpClient(),
		},
	}

	for _, option := range options {
		option(ezNode)
	}

	return ezNode
}

func WithClient(client *http.Client) Option {
	return func(ezNode *EzNode) {
		ezNode.apiCaller = &apiCallerClient{
			client: client,
		}
	}
}
