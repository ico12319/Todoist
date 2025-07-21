package gql_http_helpers

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"io"
	"net/http"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

//go:generate mockery --name=requestAuthSetter --exported --output=./mocks --outpkg=mocks --filename=request_auth_setter.go --with-expecter=true
type requestDecorator interface {
	DecorateRequest(context.Context, *http.Request) (*http.Request, error)
}

//go:generate mockery --name=httpRequester --exported --output=./mocks --outpkg=mocks --filename=http_requester2.go --with-expecter=true
type requester interface {
	NewRequestWithContext(context.Context, string, string, io.Reader) (*http.Request, error)
}

type httpResponseGetter struct {
	client    httpClient
	decorator requestDecorator
	requester requester
}

func NewHttpResponseGetter(client httpClient, decorator requestDecorator, requester requester) *httpResponseGetter {
	return &httpResponseGetter{
		client:    client,
		decorator: decorator,
		requester: requester,
	}
}

func (h *httpResponseGetter) GetHttpResponse(ctx context.Context, httpMethod string, url string, body io.Reader) (*http.Response, error) {
	log.C(ctx).Info("getting http response in http response getter")

	req, err := h.requester.NewRequestWithContext(ctx, httpMethod, url, body)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s when trying to make http request", err.Error())
		return nil, err
	}

	req, err = h.decorator.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s when trying to decorate request", err.Error())
		return nil, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s when trying to do http request", err.Error())
		return nil, err
	}

	return resp, nil
}
