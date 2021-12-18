package eznode

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"
)

type NodeUnit struct {
	url            *url.URL
	limit          uint
	limitDuration  time.Duration
	requestTimeout time.Duration
	hits           uint
	requester      Requester
}

func NewNodeUnit(
	rawUrl string,
	limit uint,
	limitDuration time.Duration,
	requestTimeout time.Duration,
	middleware RequestMiddleware,
	apiClient apiCaller,
) *NodeUnit {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		log.Fatal(err)
	}

	requester := func(request *http.Request) (*Response, error) {
		newParsedUrl, err := url.Parse(parsedUrl.String() + request.URL.String())
		if err != nil {
			log.Fatal(err)
		}

		request.URL = newParsedUrl

		ctx, cancelFunc := context.WithTimeout(context.Background(), requestTimeout)
		defer cancelFunc()

		return apiClient.doRequest(ctx, middleware(request))
	}

	return &NodeUnit{
		url:            parsedUrl,
		limit:          limit,
		limitDuration:  limitDuration,
		requestTimeout: requestTimeout,
		hits:           0,
		requester:      requester,
	}
}
