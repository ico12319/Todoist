package oauth

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/oauth/githubModels"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/utils"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
)

type githubService interface {
	GetUserOrganizations(ctx context.Context, accessToken string) ([]*githubModels.Organization, error)
	GetUserInfo(ctx context.Context, accessToken string) (*githubModels.UserInfo, error)
	GetUserInfoPrivate(ctx context.Context, accessToken string) (*githubModels.UserInfo, error)
}

type userInfoService struct {
	gService githubService
}

func NewUserInfoService(gService githubService) *userInfoService {
	return &userInfoService{gService: gService}
}

func (u *userInfoService) DetermineUserGitHubEmail(ctx context.Context, accessToken string) (string, error) {
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

		return "", fmt.Errorf("user should have at least one private or public email associated with his account")
	}

	return *userInfo.Email, nil
}

func (u *userInfoService) GetUserAppRole(ctx context.Context, accessToken string) (string, error) {
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
