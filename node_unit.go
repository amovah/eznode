package eznode

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

type NodeUnitLimit struct {
	count    uint
	duration time.Duration
}

type NodeUnit struct {
	url            *url.URL
	limit          NodeUnitLimit
	requestTimeout time.Duration
	hits           uint
	priority       uint
	requester      RequestMiddleware
}

func NewNodeUnit(
	rawUrl string,
	limit NodeUnitLimit,
	requestTimeout time.Duration,
	priority uint,
	middleware RequestMiddleware,
) *NodeUnit {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		log.Fatal(err)
	}

	requester := func(request *http.Request) *http.Request {
		newParsedUrl, err := url.Parse(parsedUrl.String() + request.URL.String())
		if err != nil {
			log.Fatal(err)
		}

		request.URL = newParsedUrl

		return middleware(request)
	}

	return &NodeUnit{
		url:            parsedUrl,
		limit:          limit,
		requestTimeout: requestTimeout,
		hits:           0,
		priority:       priority,
		requester:      requester,
	}
}
