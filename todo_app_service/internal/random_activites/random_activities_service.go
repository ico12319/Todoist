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

//go:generate mockery --name=httpService --exported --output=./mocks --outpkg=mocks --filename=http_service.go --with-expecter=true
type httpService interface {
	GetHttpResponse(ctx context.Context, httpMethod string, url string, body io.Reader) (*http.Response, error)
}

type service struct {
	apiUrl      string
	httpService httpService
}

func NewService(apiUrl string, httpService httpService) *service {
	return &service{
		apiUrl:      apiUrl,
		httpService: httpService,
	}
}

func (s *service) Suggest(ctx context.Context) (*models.RandomActivity, error) {
	log.C(ctx).Info("suggesting activity in random activity service")

	resp, err := s.httpService.GetHttpResponse(ctx, http.MethodGet, s.apiUrl, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response when trying to suggest activity, error %s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).Warnf("bad http status code received when trying to call bored-api, expected %d actual %d", http.StatusOK, resp.StatusCode)
		return nil, errors.New("error when trying to suggest an activity")
	}

	defer resp.Body.Close()

	var randomActivity models.RandomActivity
	if err = json.NewDecoder(resp.Body).Decode(&randomActivity); err != nil {
		log.C(ctx).Errorf("failed to decode response body, error %s", err.Error())
		return nil, err
	}

	return &randomActivity, nil
}
