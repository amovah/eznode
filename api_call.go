package eznode

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Response struct {
	StatusCode int
	Body       []byte
	Headers    *http.Header
	Metadata   ChainResponseMetadata
}

type ApiCaller interface {
	doRequest(context context.Context, request *http.Request) (*Response, error)
}

type apiCallerClient struct {
	client *http.Client
}

func (a *apiCallerClient) doRequest(ctx context.Context, request *http.Request) (*Response, error) {
	res, err := a.client.Do(request.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	response := &Response{
		StatusCode: res.StatusCode,
		Body:       resBody,
		Headers:    &res.Header,
	}

	return response, nil
}

func createHttpClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}

	return client
}
