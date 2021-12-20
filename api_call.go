package eznode

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Response struct {
	statusCode int
	body       []byte
	headers    *http.Header
}

type apiCaller interface {
	doRequest(context context.Context, request *http.Request) (*Response, error)
}

type apiCallerClient struct {
	client *http.Client
}

func (a *apiCallerClient) doRequest(context context.Context, request *http.Request) (*Response, error) {
	responseChannel := make(chan *Response)
	errorChannel := make(chan error)

	go func() {
		defer close(errorChannel)
		defer close(responseChannel)

		res, err := a.requestSlow(request)
		if err != nil {
			errorChannel <- err
			return
		}

		responseChannel <- res
	}()

	select {
	case <-context.Done():
		return nil, context.Err()
	case r := <-responseChannel:
		return r, nil
	case err := <-errorChannel:
		return nil, err
	}
}

func (a *apiCallerClient) requestSlow(request *http.Request) (*Response, error) {
	res, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	response := &Response{
		statusCode: res.StatusCode,
		body:       resBody,
		headers:    &res.Header,
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
