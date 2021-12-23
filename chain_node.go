package eznode

import (
	"github.com/google/uuid"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ChainNode struct {
	id             uuid.UUID
	name           string
	url            *url.URL
	limit          ChainNodeLimit
	requestTimeout time.Duration
	hits           uint
	totalHits      uint64
	priority       int
	middleware     RequestMiddleware
}

type ChainNodeData struct {
	Name           string
	Url            string
	Limit          ChainNodeLimit
	RequestTimeout time.Duration
	Priority       int
	Middleware     RequestMiddleware
}

func NewChainNode(
	chainNodeData ChainNodeData,
) *ChainNode {
	if chainNodeData.Name == "" {
		log.Fatal("name cannot be empty")
	}

	parsedUrl, err := url.Parse(chainNodeData.Url)
	if err != nil {
		log.Fatal(err)
	}

	if chainNodeData.Limit.Count <= 0 {
		log.Fatal("limit.count cannot be less than 0")
	}

	if chainNodeData.Limit.Per <= 0 {
		log.Fatal("limit.per cannot be less than 0")
	}

	if chainNodeData.RequestTimeout < 1*time.Second {
		log.Fatal("requestTimeout cannot be less than 1 second")
	}

	if chainNodeData.Priority < 0 {
		log.Fatal("priority cannot be less than 0")
	}

	middleware := func(request *http.Request) *http.Request {
		newParsedUrl, err := url.Parse(parsedUrl.String() + request.URL.String())
		if err != nil {
			log.Fatal(err)
		}

		request.URL = newParsedUrl

		return chainNodeData.Middleware(request)
	}

	return &ChainNode{
		id:             uuid.New(),
		name:           chainNodeData.Name,
		url:            parsedUrl,
		limit:          chainNodeData.Limit,
		requestTimeout: chainNodeData.RequestTimeout,
		hits:           0,
		priority:       chainNodeData.Priority,
		middleware:     middleware,
	}
}
