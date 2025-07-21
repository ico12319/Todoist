package jwt_test

import (
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/tokens"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const (
	SIGNED_JWT_STRING = "signed jwt string"
	EMAIL             = "email@email.com"
)

var (
	validUserId             = uuid.FromBytesOrNil([]byte("valid user id"))
	mockIssuedTime          = time.Now()
	mockExpiryTime          = time.Now().Add(30 * time.Minute)
	errMalformedToken       = errors.New("empty or malformed token")
	errInvalidSigningMethod = errors.New("unexpected signing method")
	errInvalidToken         = errors.New("token is no longer valid")
	errTokenExpired         = errors.New("token expired")
)

func initClaims() *jwt.Claims {
	return &jwt.Claims{
		UserId: validUserId.String(),
		Email:  EMAIL,
		Role:   string(constants.Admin),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(mockIssuedTime),
			ExpiresAt: jwt.NewNumericDate(mockExpiryTime),
		},
	}
}

func initJWT(claims *jwt.Claims, method jwt.SigningMethod, valid bool) *jwt.Token {
	return &jwt.Token{
		Claims: claims,
		Method: method,
		Valid:  valid,
	}
}
