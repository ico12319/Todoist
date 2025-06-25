package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

//go:generate mockery --name=timeGenerator --exported --output=./mocks --outpkg=mocks --filename=time_generator.go --with-expecter=true
type timeGenerator interface {
	Now() time.Time
}

//go:generate mockery --name=jwtGetter --exported --output=./mocks --outpkg=mocks --filename=jwt_getter.go --with-expecter=true
type jwtGetter interface {
	GetJWTWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token
	GetSignedJWT(jwt *jwt.Token, key interface{}) (string, error)
}

type refreshTokenBuilder struct {
	generator   timeGenerator
	tokenGetter jwtGetter
}

func NewRefreshTokenBuilder(generator timeGenerator, tokenGetter jwtGetter) *refreshTokenBuilder {
	return &refreshTokenBuilder{generator: generator, tokenGetter: tokenGetter}
}

func (r *refreshTokenBuilder) GenerateRefreshToken(ctx context.Context) (string, error) {
	configManager := log.GetInstance()

	log.C(ctx).Info("generating refresh token")
	expirationTime := r.generator.Now().Add(150 * time.Hour)

	claims := &Claims{
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(r.generator.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	jwtToken := r.tokenGetter.GetJWTWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := r.tokenGetter.GetSignedJWT(jwtToken, configManager.JwtConfig.Secret)
	if err != nil {
		log.C(ctx).Errorf("failed to sign jwt, error %s", err.Error())
		return "", utils.DetermineErrorWhenSigningJWT(err)
	}

	return signedToken, nil
}
