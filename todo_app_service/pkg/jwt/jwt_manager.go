package jwt

import (
	config "Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/golang-jwt/jwt/v5"
)

// wrapper around jwt package functions, used for better testing
type jwtManager struct {
	configManager *config.Config
}

func NewJwtManager() *jwtManager {
	return &jwtManager{configManager: config.GetInstance()}
}

func (*jwtManager) GetSignedJWT(jwt *jwt.Token, key interface{}) (string, error) {
	return jwt.SignedString(key)
}

func (*jwtManager) GetJWTWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token {
	return jwt.NewWithClaims(method, claims)
}

func (j *jwtManager) ParseWithClaims(tokenString string, claims *Claims) (*jwt.Token, *Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.configManager.JwtConfig.Secret, nil
	})

	if err != nil {
		return nil, nil, err
	}

	return token, claims, nil
}
