package oauth

import (
	config "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"golang.org/x/oauth2"
)

type stateGenerator interface {
	GenerateState() (string, error)
}
type oauthService struct {
	sGenerator  stateGenerator
	oauthConfig *oauth2.Config
}

func NewOauthService(sGenerator stateGenerator, oauthConfig *oauth2.Config) *oauthService {
	return &oauthService{
		sGenerator:  sGenerator,
		oauthConfig: oauthConfig,
	}
}

func (o *oauthService) LoginUrl(ctx context.Context) (string, string, error) {
	config.C(ctx).Info("getting url where user should be redirected when trying to log in")

	state, err := o.sGenerator.GenerateState()
	if err != nil {
		config.C(ctx).Errorf("failed to login user, error %s when generating state", err.Error())
		return "", "", err
	}

	return o.oauthConfig.AuthCodeURL(state), state, nil
}

func (o *oauthService) ExchangeCodeForToken(ctx context.Context, authCode string) (string, error) {
	config.C(ctx).Info("exchanging auth code for access token in oauth service")

	token, err := o.oauthConfig.Exchange(ctx, authCode)
	if err != nil {
		config.C(ctx).Errorf("failed to exchange auth code for access token, error %s", err.Error())
		return "", err
	}

	return token.AccessToken, nil
}
