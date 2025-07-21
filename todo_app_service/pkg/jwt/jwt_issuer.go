package jwt

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
)

type jwtBuilder interface {
	GenerateJWT(context.Context, string, string) (string, error)
	GenerateRefreshToken(context.Context) (string, error)
}

type refreshTokenService interface {
	UpsertRefreshToken(ctx context.Context, refresh *models.Refresh, userEmail string) error
	UpdateRefreshToken(ctx context.Context, refreshToken string, userId string) (*models.Refresh, error)
	GetTokenOwner(context.Context, string) (*models.User, error)
}

type emailSearcher interface {
	GetUserRecordByEmail(context.Context, string) (*models.User, error)
}

type userInfoAggregator interface {
	AggregateUserInfo(context.Context, string) (string, string, error)
}

type jwtIssuer struct {
	builder      jwtBuilder
	aggregator   userInfoAggregator
	service      refreshTokenService
	userSearcher emailSearcher
}

func NewJwtIssuer(builder jwtBuilder, aggregator userInfoAggregator, service refreshTokenService, userSearcher emailSearcher) *jwtIssuer {
	return &jwtIssuer{
		builder:      builder,
		aggregator:   aggregator,
		service:      service,
		userSearcher: userSearcher,
	}
}

func (j *jwtIssuer) GetTokens(ctx context.Context, accessToken string) (*models.CallbackResponse, error) {
	log.C(ctx).Info("getting jwt token in oauth service")

	email, role, err := j.aggregator.AggregateUserInfo(ctx, accessToken)
	if err != nil {
		log.C(ctx).Errorf("failed to aggregate user info in jwt issuer, error %s", err.Error())
		return nil, err
	}

	jwtToken, err := j.builder.GenerateJWT(ctx, email, role)
	if err != nil {
		log.C(ctx).Errorf("failed to create jwt token, error %s when generating...", err.Error())
		return nil, err
	}

	refreshToken, err := j.builder.GenerateRefreshToken(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get tokens, error %s when generatin refresh token", err.Error())
		return nil, err
	}

	user, err := j.userSearcher.GetUserRecordByEmail(ctx, email)
	if err != nil {
		log.C(ctx).Errorf("failed to get user record by email when trying to get api tokens, error %s", err.Error())
		return nil, err
	}

	if err = j.service.UpsertRefreshToken(ctx, &models.Refresh{
		RefreshToken: refreshToken,
		UserId:       user.Id,
	}, user.Email); err != nil {
		log.C(ctx).Errorf("failed to upsert refresh token im jwt issuer, error %s", err.Error())
		return nil, err
	}

	return &models.CallbackResponse{
		RefreshToken: refreshToken,
		JwtToken:     jwtToken,
	}, nil

}

func (j *jwtIssuer) GetRenewedTokens(ctx context.Context, refresh *handler_models.Refresh) (*models.CallbackResponse, error) {
	log.C(ctx).Info("renewing refresh and jwt in oauth service")

	tokenOwner, err := j.service.GetTokenOwner(ctx, refresh.RefreshToken)
	if err != nil {
		log.C(ctx).Errorf("failed to get renewed tokens, error %s when trying to get token owner", err.Error())
		return nil, err
	}

	refreshedJwtToken, err := j.builder.GenerateJWT(ctx, tokenOwner.Email, string(tokenOwner.Role))
	if err != nil {
		log.C(ctx).Errorf("failed to get renewed tokens, error %s when trying to generate new jwt token", err.Error())
		return nil, err
	}

	newRefreshToken, err := j.builder.GenerateRefreshToken(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get renewed token, error %s when trying to generate new refresh token", err.Error())
		return nil, err
	}

	if _, err = j.service.UpdateRefreshToken(ctx, newRefreshToken, tokenOwner.Id); err != nil {
		log.C(ctx).Errorf("failed to get renewed token, error %s when trying to update refresh token", err.Error())
		return nil, err
	}

	return &models.CallbackResponse{
		JwtToken:     refreshedJwtToken,
		RefreshToken: newRefreshToken,
	}, nil
}
