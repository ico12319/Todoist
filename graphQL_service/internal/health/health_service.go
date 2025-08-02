package health

import (
	"Todo-List/internProject/graphQL_service/internal/gql_errors"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"io"
	"net/http"
)

type httpService interface {
	GetHttpResponse(ctx context.Context, httpMethod string, url string, body io.Reader) (*http.Response, error)
}

type service struct {
	httpService httpService
}

func NewService(httpService httpService) *service {
	return &service{
		httpService: httpService,
	}
}

func (s *service) CheckRESTProbes(ctx context.Context, url string) error {
	log.C(ctx).Info("checking whether REST api is alive in health service in GQL")

	resp, err := s.httpService.GetHttpResponse(ctx, http.MethodGet, url, nil)
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
