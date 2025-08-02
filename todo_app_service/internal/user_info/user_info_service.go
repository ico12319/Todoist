package user_info

import (
	"Todo-List/internProject/todo_app_service/internal/gitHub"
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"errors"
)

type githubService interface {
	GetUserOrganizations(ctx context.Context, accessToken string) ([]*gitHub.Organization, error)
	GetUserInfo(ctx context.Context, accessToken string) (*gitHub.UserInfo, error)
	GetUserInfoPrivate(ctx context.Context, accessToken string) (*gitHub.UserInfo, error)
}

type service struct {
	gService githubService
}

func NewUserInfoService(gService githubService) *service {
	return &service{gService: gService}
}

func (u *service) DetermineUserGitHubEmail(ctx context.Context, accessToken string) (string, error) {
	log.C(ctx).Info("determining email in oauth service")

	userInfo, err := u.gService.GetUserInfo(ctx, accessToken)
	if err != nil {
		log.C(ctx).Errorf("failed to get user organizations, error %s when calling github service", err.Error())
		return "", nil
	}

	if userInfo.Email == nil {
		log.C(ctx).Info("user does not have public email associated with his account, getting private ones...")
		userInfo, err = u.gService.GetUserInfoPrivate(ctx, accessToken)
	}

	if userInfo == nil || userInfo.Email == nil {
		log.C(ctx).Info("user does not have private emails associated with his account, unable to create user")

		return "", errors.New("user should have at least one private or public email associated with his account")
	}

	return *userInfo.Email, nil
}

func (u *service) GetUserAppRole(ctx context.Context, accessToken string) (string, error) {
	userOrganizations, err := u.gService.GetUserOrganizations(ctx, accessToken)
	if err != nil {
		log.C(ctx).Errorf("failed to get user organizations, error %s when calling github service", err.Error())
		return "", nil
	}

	role, err := utils.DetermineRole(userOrganizations)
	if err != nil {
		log.C(ctx).Errorf("failed to determine user role, error %s", err.Error())
		return "", nil
	}

	return role, nil
}
