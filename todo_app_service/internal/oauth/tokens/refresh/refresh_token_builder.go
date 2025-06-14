package refresh

import (
	"context"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type timeGenerator interface {
	Now() time.Time
}
type refreshTokenBuilder struct {
	generator timeGenerator
}

func NewRefreshTokenBuilder(generator timeGenerator) *refreshTokenBuilder {
	return &refreshTokenBuilder{generator: generator}
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

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := jwtToken.SignedString(configManager.JwtConfig.Secret)
	if err != nil {
		log.C(ctx).Errorf("failed to sign jwt, error %s", err.Error())
		return "", err
	}

	return signedToken, nil
}
