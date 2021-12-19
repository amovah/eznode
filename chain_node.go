package eznode

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"
)

type NodeUnitLimit struct {
	count uint
	per   time.Duration
}

type ChainNode struct {
	name           string
	url            *url.URL
	limit          NodeUnitLimit
	requestTimeout time.Duration
	hits           uint
	priority       uint
	middleware     RequestMiddleware
}

type ChainNodeData struct {
	name           string
	url            string
	limit          NodeUnitLimit
	requestTimeout time.Duration
	priority       uint
	middleware     RequestMiddleware
}

func NewChainNode(
	chainNodeData ChainNodeData,
) (*ChainNode, error) {
	if chainNodeData.name != "" {
		return nil, errors.New("name cannot be empty")
	}

	parsedUrl, err := url.Parse(chainNodeData.url)
	if err != nil {
		return nil, err
	}

	if chainNodeData.limit.per <= 0 {
		return nil, errors.New("limit per cannot less and equal than 0")
	}

	if chainNodeData.requestTimeout < 1*time.Second {
		return nil, errors.New("request timeout cannot be less than 1 second")
	}

	middleware := func(request *http.Request) *http.Request {
		newParsedUrl, err := url.Parse(parsedUrl.String() + request.URL.String())
		if err != nil {
			log.Fatal(err)
		}

		request.URL = newParsedUrl

		return chainNodeData.middleware(request)
	}

	return &ChainNode{
		name:           chainNodeData.name,
		url:            parsedUrl,
		limit:          chainNodeData.limit,
		requestTimeout: chainNodeData.requestTimeout,
		hits:           0,
		priority:       chainNodeData.priority,
		middleware:     middleware,
	}, nil
}
