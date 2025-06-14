package oauth

import (
	"context"
	config "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/handler_models"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
)

type stateGenerator interface {
	GenerateState() (string, error)
}

type jwtBuilder interface {
	GenerateJWT(ctx context.Context, email string, role string) (string, error)
}

type refreshTokenBuilder interface {
	GenerateRefreshToken(ctx context.Context) (string, error)
}

type userInfoRetriever interface {
	DetermineUserGitHubEmail(ctx context.Context, accessToken string) (string, error)
	GetUserAppRole(ctx context.Context, accessToken string) (string, error)
}

type refreshTokenService interface {
	CreateRefreshToken(ctx context.Context, email string, refreshToken string) (*models.Refresh, error)
	UpdateRefreshToken(ctx context.Context, refreshToken string, userId string) (*models.Refresh, error)
	GetTokenOwner(ctx context.Context, refreshToken string) (*models.User, error)
}
type oauthService struct {
	refreshService refreshTokenService
	uInfoService   userInfoRetriever
	sGenerator     stateGenerator
	refreshBuilder refreshTokenBuilder
	jwtBuilder     jwtBuilder
	configManager  *config.Config
}

func NewOauthService(uInfoService userInfoRetriever, refreshService refreshTokenService, sGenerator stateGenerator, refreshBuilder refreshTokenBuilder, jwtBuilder jwtBuilder, configManger *config.Config) *oauthService {
	return &oauthService{uInfoService: uInfoService, refreshService: refreshService, sGenerator: sGenerator, refreshBuilder: refreshBuilder, jwtBuilder: jwtBuilder, configManager: configManger}
}

func (o *oauthService) LoginUrl(ctx context.Context) (string, string, error) {
	config.C(ctx).Info("getting url where user should be redirected when trying to log in")

	state, err := o.sGenerator.GenerateState()
	if err != nil {
		config.C(ctx).Errorf("failed to login user, error %s when generating state", err.Error())
		return "", "", err
	}

	return o.configManager.OauthConfig.AuthCodeURL(state), state, nil
}

func (o *oauthService) ExchangeCodeForToken(ctx context.Context, authCode string) (string, error) {
	config.C(ctx).Info("exchanging auth code for access token in oauth service")

	token, err := o.configManager.OauthConfig.Exchange(ctx, authCode)
	if err != nil {
		config.C(ctx).Errorf("failed to exchange auth code for access token, error %s", err.Error())
		return "", err
	}

	return token.AccessToken, nil
}

func (o *oauthService) GetTokens(ctx context.Context, accessToken string) (*models.CallbackResponse, error) {
	config.C(ctx).Info("getting jwt token in oauth service")

	userEmail, err := o.uInfoService.DetermineUserGitHubEmail(ctx, accessToken)
	if err != nil {
		config.C(ctx).Errorf("failed to get jwt, error %s when trying to determine user's github email", err.Error())
		return nil, err
	}

	userRole, err := o.uInfoService.GetUserAppRole(ctx, accessToken)
	if err != nil {
		config.C(ctx).Errorf("failed to get jwt, error %s when trying to get user's app role", err.Error())
		return nil, err
	}

	jwtToken, err := o.jwtBuilder.GenerateJWT(ctx, userEmail, userRole)
	if err != nil {
		config.C(ctx).Errorf("failed to create jwt token, error %s when generating...", err.Error())
		return nil, err
	}

	refreshToken, err := o.refreshBuilder.GenerateRefreshToken(ctx)
	if err != nil {
		config.C(ctx).Errorf("failed to get tokens, error %s when generatin refresh token", err.Error())
		return nil, err
	}

	if _, err = o.refreshService.CreateRefreshToken(ctx, userEmail, refreshToken); err != nil {
		config.C(ctx).Errorf("failed to create refresh token, error %s when generating...", err.Error())
		return nil, err
	}

	return &models.CallbackResponse{
		RefreshToken: refreshToken,
		JwtToken:     jwtToken,
	}, nil
}

func (o *oauthService) GetRenewedTokens(ctx context.Context, refresh *handler_models.Refresh) (*models.CallbackResponse, error) {
	config.C(ctx).Info("renewing refresh and jwt in oauth service")

	tokenOwner, err := o.refreshService.GetTokenOwner(ctx, refresh.RefreshToken)
	if err != nil {
		config.C(ctx).Errorf("failed to get renewed tokens, error %s when trying to get token owner", err.Error())
		return nil, err
	}

	refreshedJwtToken, err := o.jwtBuilder.GenerateJWT(ctx, tokenOwner.Email, string(tokenOwner.Role))
	if err != nil {
		config.C(ctx).Errorf("failed to get renewed tokens, error %s when trying to generate new jwt token", err.Error())
		return nil, err
	}

	newRefreshToken, err := o.refreshBuilder.GenerateRefreshToken(ctx)
	if err != nil {
		config.C(ctx).Errorf("failed to get renewed token, error %s when trying to generate new refresh token", err.Error())
		return nil, err
	}

	if _, err = o.refreshService.UpdateRefreshToken(ctx, newRefreshToken, tokenOwner.Id); err != nil {
		config.C(ctx).Errorf("failed to get renewed token, error %s when trying to update refresh token", err.Error())
		return nil, err
	}

	return &models.CallbackResponse{
		JwtToken:     refreshedJwtToken,
		RefreshToken: newRefreshToken,
	}, nil
}
