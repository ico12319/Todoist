package tokens

import (
	config "Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/golang-jwt/jwt/v5"
)

type jwtParser struct {
	configManager *config.Config
}

func NewJwtParser() *jwtParser {
	return &jwtParser{
		configManager: config.GetInstance(),
	}
}

func (j *jwtParser) ParseWithClaims(tokenString string, claims *Claims) (*jwt.Token, *Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.configManager.JwtConfig.Secret, nil
	})

	if err != nil {
		return nil, nil, err
	}

	return token, claims, nil
}
