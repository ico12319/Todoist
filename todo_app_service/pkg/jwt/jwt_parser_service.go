package jwt

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

//go:generate mockery --name=parser --exported --output=./mocks --outpkg=mocks --filename=jwt_parser.go --with-expecter=true
type parser interface {
	ParseWithClaims(string, *Claims) (*jwt.Token, *Claims, error)
}

type JwtParserService struct {
	parser parser
}

func NewJwtParseService(parser parser) *JwtParserService {
	return &JwtParserService{parser: parser}
}

func (j *JwtParserService) ParseJWT(ctx context.Context, tokenString string) (*Claims, error) {
	log.C(ctx).Info("parsing jwt token")

	claims := &Claims{}

	token, claims, err := j.parser.ParseWithClaims(tokenString, claims)
	if err != nil {
		log.C(ctx).Errorf("failed to parse jwt with claims, error %s", err.Error())
		return nil, utils.DetermineJWTErrorWhenParsingWithClaims(ctx, err)
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		log.C(ctx).Errorf("wrong signing method")
		return nil, errors.New("unexpected signing method")
	}

	if !token.Valid {
		log.C(ctx).Error("the parsed jwt token is not valid")
		return nil, errors.New("token is no longer valid")
	}

	return claims, nil
}
