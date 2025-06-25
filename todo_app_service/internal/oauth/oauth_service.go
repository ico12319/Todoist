package oauth

import (
	config "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
)

type ChannelResult[T any] struct {
	Result T
	Err    error
}

type stateGenerator interface {
	GenerateState() (string, error)
}

type jwtBuilder interface {
	GenerateJWT(context.Context, string, string) (string, error)
}

type refreshTokenBuilder interface {
	GenerateRefreshToken(context.Context) (string, error)
}

type userInfoRetriever interface {
	DetermineUserGitHubEmail(context.Context, string) (string, error)
	GetUserAppRole(context.Context, string) (string, error)
}

type refreshTokenService interface {
	CreateRefreshToken(context.Context, string, string) (*models.Refresh, error)
	UpdateRefreshToken(context.Context, string, string) (*models.Refresh, error)
	GetTokenOwner(context.Context, string) (*models.User, error)
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

	emailChan := make(chan ChannelResult[string], 1)
	roleChan := make(chan ChannelResult[string], 1)

	go func() {
		userEmail, err := o.uInfoService.DetermineUserGitHubEmail(ctx, accessToken)
		emailChan <- ChannelResult[string]{userEmail, err}
	}()

	go func() {
		userRole, err := o.uInfoService.GetUserAppRole(ctx, accessToken)
		roleChan <- ChannelResult[string]{userRole, err}
	}()

	var email string
	var role string

	for i := 0; i < constants.GOROUTINES_COUNT; i++ {
		select {
		case chRes := <-emailChan:
			if chRes.Err != nil {
				config.C(ctx).Errorf("failed to get jwt, error %s when trying to determine user's github email", chRes.Err.Error())
				return nil, chRes.Err
			}
			email = chRes.Result

		case chRes := <-roleChan:
			if chRes.Err != nil {
				config.C(ctx).Errorf("failed to get jwt, error %s when trying to get user's app role", chRes.Err.Error())
				return nil, chRes.Err
			}
			role = chRes.Result

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	jwtToken, err := o.jwtBuilder.GenerateJWT(ctx, email, role)
	if err != nil {
		config.C(ctx).Errorf("failed to create jwt token, error %s when generating...", err.Error())
		return nil, err
	}

	refreshToken, err := o.refreshBuilder.GenerateRefreshToken(ctx)
	if err != nil {
		config.C(ctx).Errorf("failed to get tokens, error %s when generatin refresh token", err.Error())
		return nil, err
	}

	if _, err = o.refreshService.CreateRefreshToken(ctx, email, refreshToken); err != nil {
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
