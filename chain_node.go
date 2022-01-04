package eznode

import (
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type ChainNode struct {
	name           string
	url            *url.URL
	limit          ChainNodeLimit
	requestTimeout time.Duration
	hits           uint
	totalHits      uint64
	responseStats  map[int]uint64
	statsMutex     *sync.Mutex
	priority       int
	middleware     RequestMiddleware
	disabled       bool
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

	if chainNodeData.Limit.Count < 1 {
		log.Fatal("limit.count cannot be less than 1")
	}

	if chainNodeData.Limit.Per < 1 {
		log.Fatal("limit.per cannot be less than 1")
	}

	if chainNodeData.RequestTimeout < 1 {
		log.Fatal("requestTimeout cannot be less than 1")
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

		if chainNodeData.Middleware != nil {
			return chainNodeData.Middleware(request)
		}

		return request
	}

	return &ChainNode{
		name:           chainNodeData.Name,
		url:            parsedUrl,
		limit:          chainNodeData.Limit,
		requestTimeout: chainNodeData.RequestTimeout,
		hits:           0,
		totalHits:      0,
		responseStats:  make(map[int]uint64),
		statsMutex:     &sync.Mutex{},
		priority:       chainNodeData.Priority,
		middleware:     middleware,
		disabled:       false,
	}
}
