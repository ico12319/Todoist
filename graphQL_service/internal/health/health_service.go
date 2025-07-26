package health

import (
	"Todo-List/internProject/graphQL_service/internal/gql_errors"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"io"
	"net/http"
)

type httpService interface {
	NewRequestWithContext(ctx context.Context, httpMethod string, url string, reader io.Reader) (*http.Request, error)
}

type httpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type service struct {
	requester httpService
	client    httpClient
}

func NewService(requester httpService, client httpClient) *service {
	return &service{
		requester: requester,
		client:    client,
	}
}

func (s *service) CheckRESTProbes(ctx context.Context, url string) error {
	log.C(ctx).Info("checking whether REST api is alive in health service in GQL")

	req, err := s.requester.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to make http request to REST api in health service in GQL, error %s", err.Error())
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to receive http response from REST api in health service in GQL, error %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).Debugf("bad http status code recieved expected %d, actual %d", http.StatusOK, resp.StatusCode)
		return gql_errors.NewRestHealthError(resp.StatusCode)
	}

	return nil
}
