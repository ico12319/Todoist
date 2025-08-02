package gitHub

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type httpService interface {
	GetHttpResponseWithAccessCode(ctx context.Context, httpMethod string, url string, body io.Reader, accessToken string) (*http.Response, error)
}
type service struct {
	httpService httpService
}

func NewService(httpService httpService) *service {
	return &service{
		httpService: httpService,
	}
}

func (s *service) GetUserOrganizations(ctx context.Context, accessToken string) ([]*Organization, error) {
	log.C(ctx).Info("getting user organizations in github service")

	url := "https://api.github.com/user/orgs"

	resp, err := s.httpService.GetHttpResponseWithAccessCode(ctx, http.MethodGet, url, nil, accessToken)
	if err != nil {
		log.C(ctx).Errorf("failed to get user organizations, error %s when trying to get http response", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var userOrganizations []*Organization
	if err = json.NewDecoder(resp.Body).Decode(&userOrganizations); err != nil {
		log.C(ctx).Errorf("failed to get user organizations, error %s when trying to decode JSON response body", err.Error())
		return nil, err
	}

	return userOrganizations, nil
}

func (s *service) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	log.C(ctx).Info("getting user info in github service")

	url := "https://api.github.com/user"

	resp, err := s.httpService.GetHttpResponseWithAccessCode(ctx, http.MethodGet, url, nil, accessToken)
	if err != nil {
		log.C(ctx).Errorf("failed to get user info http response, error %s when calling get user info response method", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo UserInfo
	if err = json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.C(ctx).Errorf("failed to get user info, error %s when trying to decode JSON response body", err.Error())
		return nil, err
	}

	return &userInfo, nil
}

func (s *service) GetUserInfoPrivate(ctx context.Context, accessToken string) (*UserInfo, error) {
	log.C(ctx).Info("getting user private info in github service")

	url := "https://api.github.com/user/emails"

	resp, err := s.httpService.GetHttpResponseWithAccessCode(ctx, http.MethodGet, url, nil, accessToken)
	if err != nil {
		log.C(ctx).Errorf("failed to get private user info http response, error %s when calling get user info response method", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var privateUserInfo []*UserInfo
	if err = json.NewDecoder(resp.Body).Decode(&privateUserInfo); err != nil {
		log.C(ctx).Errorf("failed to get private user info, error %s when trying to decode JSON response body", err.Error())
		return nil, err
	}

	if len(privateUserInfo) == 0 {
		log.C(ctx).Info("private user info is empty!")
		return nil, nil
	}

	return privateUserInfo[0], nil
}
