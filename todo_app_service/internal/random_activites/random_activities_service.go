package random_activites

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type httpService interface {
	NewRequestWithContext(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error)
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type service struct {
	apiUrl    string
	requester httpService
	client    httpClient
}

func NewService(apiUrl string, requester httpService, client httpClient) *service {
	return &service{
		apiUrl:    apiUrl,
		requester: requester,
		client:    client,
	}
}

func (s *service) Suggest(ctx context.Context) (*models.RandomActivity, error) {
	log.C(ctx).Info("suggesting activity in random activity service")

	resp, err := s.getBoredApiResponse(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var randomActivity models.RandomActivity
	if err = json.NewDecoder(resp.Body).Decode(&randomActivity); err != nil {
		log.C(ctx).Errorf("failed to decode response body, error %s", err.Error())
		return nil, err
	}

	return &randomActivity, nil
}

func (s *service) getBoredApiResponse(ctx context.Context) (*http.Response, error) {
	log.C(ctx).Info("trying to get http response from bored api in random_activities_api")

	req, err := s.requester.NewRequestWithContext(ctx, http.MethodGet, s.apiUrl, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to make http request to bored api, error %s", err.Error())
		return nil, errors.New("error when trying to make http request")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response from bored api, error %s", err.Error())
		return nil, errors.New("error when trying to get http response")
	}

	return resp, nil
}
