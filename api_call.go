package eznode

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Response is the response from an API call (eznode final result)
type Response struct {
	// StatusCode is the HTTP status code of the response
	StatusCode int
	// Body is the response body
	Body []byte
	// Headers is the response headers
	Headers *http.Header
	// Metadata is the response metadata, it includes trace of request which it takes to get the response
	// also it includes the error and which node it was sent to
	Metadata ChainResponseMetadata
}

// ApiCaller is the interface for making API calls
type ApiCaller interface {
	DoRequest(context context.Context, request *http.Request) (*Response, error)
}

type apiCallerClient struct {
	client *http.Client
}

func (a *apiCallerClient) DoRequest(ctx context.Context, request *http.Request) (*Response, error) {
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
