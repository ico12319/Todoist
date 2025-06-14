package tokens

import (
	"context"
	"errors"
	"fmt"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/golang-jwt/jwt/v5"
)

type jwtParserService struct{}

func NewJwtParseService() *jwtParserService {
	return &jwtParserService{}
}

func (j *jwtParserService) ParseJWT(ctx context.Context, tokenString string) (*Claims, error) {
	log.C(ctx).Info("parsing jwt token")
	configManager := log.GetInstance()

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return configManager.JwtConfig.Secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			log.C(ctx).Errorf("failed to parse jwt, error %s invalid signature", err.Error())
			return nil, fmt.Errorf("invalid signature")
		}
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired")
		}
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	if !token.Valid {
		log.C(ctx).Error("the parsed jwt token is not valid")
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
