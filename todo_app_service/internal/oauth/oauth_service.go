package oauth

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"errors"
	"golang.org/x/oauth2"
)

//go:generate mockery --name=stateGenerator --exported --output=./mocks --outpkg=mocks --filename=state_generator.go --with-expecter=true
type stateGenerator interface {
	GenerateState() (string, error)
}

//go:generate mockery --name=oauthConfigurator --exported --output=./mocks --outpkg=mocks --filename=oauth_configurator.go --with-expecter=true
type oauthConfigurator interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type service struct {
	sGenerator  stateGenerator
	oauthConfig oauthConfigurator
}

func NewService(sGenerator stateGenerator, oauthConfig oauthConfigurator) *service {
	return &service{
		sGenerator:  sGenerator,
		oauthConfig: oauthConfig,
	}
}

func (s *service) LoginUrl(ctx context.Context) (string, string, error) {
	log.C(ctx).Info("getting url where user should be redirected when trying to log in")

	state, err := s.sGenerator.GenerateState()
	if err != nil {
		log.C(ctx).Errorf("failed to login user, error %s when generating state", err.Error())
		return "", "", err
	}

	if len(state) == 0 {
		log.C(ctx).Warn("length of state is 0...")
		return "", "", errors.New("zero length state generated")
	}

	return s.oauthConfig.AuthCodeURL(state), state, nil
}

func (s *service) ExchangeCodeForToken(ctx context.Context, authCode string) (string, error) {
	log.C(ctx).Info("exchanging auth code for access token in oauth service")

	if len(authCode) == 0 {
		log.C(ctx).Warn("length og auth code is 0...")
		return "", errors.New("empty authorization code")
	}

	token, err := s.oauthConfig.Exchange(ctx, authCode)
	if err != nil {
		log.C(ctx).Errorf("failed to exchange auth code for access token, error %s", err.Error())
		return "", err
	}

	return token.AccessToken, nil
}
