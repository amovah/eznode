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
	priority       uint
	middleware     RequestMiddleware
}

type ChainNodeData struct {
	name           string
	url            string
	limit          ChainNodeLimit
	requestTimeout time.Duration
	priority       uint
	middleware     RequestMiddleware
}

func NewChainNode(
	chainNodeData ChainNodeData,
) *ChainNode {
	if chainNodeData.name == "" {
		log.Fatal("name cannot be empty")
	}

	parsedUrl, err := url.Parse(chainNodeData.url)
	if err != nil {
		log.Fatal(err)
	}

	if chainNodeData.limit.count <= 0 {
		log.Fatal("limit.count cannot be less than 0")
	}

	if chainNodeData.limit.per <= 0 {
		log.Fatal("limit.per cannot be less than 0")
	}

	if chainNodeData.requestTimeout < 1*time.Second {
		log.Fatal("requestTimeout cannot be less than 1 second")
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
		id:             uuid.New(),
		name:           chainNodeData.name,
		url:            parsedUrl,
		limit:          chainNodeData.limit,
		requestTimeout: chainNodeData.requestTimeout,
		hits:           0,
		priority:       chainNodeData.priority,
		middleware:     middleware,
	}
}
