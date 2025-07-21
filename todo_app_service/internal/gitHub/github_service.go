package gitHub

import (
	"Todo-List/internProject/graphQL_service/graph/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"fmt"
	"io"
	"net/http"
)

type httpRequester interface {
	NewRequestWithContext(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error)
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type service struct {
	requester httpRequester
	client    httpClient
}

func NewService(requester httpRequester, client httpClient) *service {
	return &service{requester: requester,
		client: client,
	}
}

func (s *service) GetUserOrganizations(ctx context.Context, accessToken string) ([]*Organization, error) {
	log.C(ctx).Info("getting user organizations")

	resp, err := s.getGithubResponse(ctx, accessToken, "https://api.github.com/user/orgs")
	if err != nil {
		log.C(ctx).Errorf("failed to get user organizations, error %s when trying to get http response", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	userOrganizations, err := utils.DecodeJsonResponse[[]*Organization](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to get user organizations, error %s when trying to decode JSON response body", err.Error())
		return nil, err
	}

	return userOrganizations, nil
}

func (s *service) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	resp, err := s.getGithubResponse(ctx, accessToken, "https://api.github.com/user")
	if err != nil {
		log.C(ctx).Errorf("failed to get user info http response, error %s when calling get user info response method", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	userInfo, err := utils.DecodeJsonResponse[*UserInfo](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to get user info, error %s when trying to decode JSON response body", err.Error())
		return nil, err
	}
	return userInfo, nil
}

func (s *service) GetUserInfoPrivate(ctx context.Context, accessToken string) (*UserInfo, error) {
	resp, err := s.getGithubResponse(ctx, accessToken, "https://api.github.com/user/emails")
	if err != nil {
		log.C(ctx).Errorf("failed to get private user info http response, error %s when calling get user info response method", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	privateUserInfo, err := utils.DecodeJsonResponse[[]*UserInfo](resp)
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

func (s *service) getGithubResponse(ctx context.Context, accessToken string, url string) (*http.Response, error) {
	log.C(ctx).Info("getting user info")

	req, err := s.requester.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
