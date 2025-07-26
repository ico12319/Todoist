package http_helpers

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"io"
	"net/http"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type requestDecorator interface {
	DecorateRequest(context.Context, *http.Request) (*http.Request, error)
}

type service struct {
	client    httpClient
	decorator requestDecorator
}

func NewService(client httpClient, decorator requestDecorator) *service {
	return &service{
		client:    client,
		decorator: decorator,
	}
}

func (*service) NewRequestWithContext(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, url, body)
}

func (*service) SetCookie(w http.ResponseWriter, cookie *http.Cookie) {
	http.SetCookie(w, cookie)
}

func (*service) Redirect(w http.ResponseWriter, r *http.Request, url string, httpStatusCode int) {
	http.Redirect(w, r, url, httpStatusCode)
}

func (s *service) GetHttpResponse(ctx context.Context, httpMethod string, url string, body io.Reader) (*http.Response, error) {
	log.C(ctx).Info("getting http response in http response getter")

	req, err := s.NewRequestWithContext(ctx, httpMethod, url, body)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s when trying to make http request", err.Error())
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s when trying to do http request", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *service) GetHttpResponseWithAuthHeader(ctx context.Context, httpMethod string, url string, body io.Reader) (*http.Response, error) {
	log.C(ctx).Info("getting http response in http response getter")

	req, err := s.NewRequestWithContext(ctx, httpMethod, url, body)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s when trying to make http request", err.Error())
		return nil, err
	}

	req, err = s.decorator.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s when trying to decorate request", err.Error())
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s when trying to do http request", err.Error())
		return nil, err
	}

	return resp, nil
}
