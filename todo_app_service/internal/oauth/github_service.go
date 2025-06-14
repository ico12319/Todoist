package oauth

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/utils"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/oauth/githubModels"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"net/http"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type service struct {
	client httpClient
}

func NewService(client httpClient) *service {
	return &service{client: client}
}

func (s *service) GetUserOrganizations(ctx context.Context, accessToken string) ([]*githubModels.Organization, error) {
	log.C(ctx).Info("getting user organizations")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/orgs", nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get user otganizations, error %s when making http request", err.Error())
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := s.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get user organizations, error %s when trying to get http response", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	userOrganizations, err := utils.DecodeJsonResponse[[]*githubModels.Organization](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to get user organizations, error %s when trying to decode JSON response body", err.Error())
		return nil, err
	}

	return userOrganizations, nil
}

func (s *service) GetUserInfo(ctx context.Context, accessToken string) (*githubModels.UserInfo, error) {
	resp, err := s.getUserInfoResponse(ctx, accessToken, "https://api.github.com/user")
	if err != nil {
		log.C(ctx).Errorf("failed to get user info http response, error %s when calling get user info response method", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	userInfo, err := utils.DecodeJsonResponse[*githubModels.UserInfo](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to get user info, error %s when trying to decode JSON response body", err.Error())
		return nil, err
	}
	return userInfo, nil
}

func (s *service) GetUserInfoPrivate(ctx context.Context, accessToken string) (*githubModels.UserInfo, error) {
	resp, err := s.getUserInfoResponse(ctx, accessToken, "https://api.github.com/user/emails")
	if err != nil {
		log.C(ctx).Errorf("failed to get private user info http response, error %s when calling get user info response method", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	privateUserInfo, err := utils.DecodeJsonResponse[[]*githubModels.UserInfo](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to get private user info, error %s when trying to decode JSON response body", err.Error())
		return nil, err
	}

	if len(privateUserInfo) == 0 {
		log.C(ctx).Info("private user info is empty!")
		return nil, nil
	}

	return privateUserInfo[0], nil
}

func (s *service) getUserInfoResponse(ctx context.Context, accessToken string, url string) (*http.Response, error) {
	log.C(ctx).Info("getting user info")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get user info, error %s when making http request", err.Error())
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := s.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get user info, error %s when trying to get http response", err.Error())
		return nil, err
	}
	return resp, err
}
